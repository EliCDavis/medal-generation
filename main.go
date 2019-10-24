package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
	"github.com/golang/freetype/truetype"
	"github.com/pradeep-pyro/triangle"
	"golang.org/x/image/font"
)

func makeSquareWithTexture(
	bottomLeft vector.Vector3,
	topLeft vector.Vector3,
	topRight vector.Vector3,
	bottomRight vector.Vector3,

	bottomLeftTexture vector.Vector2,
	topLeftTexture vector.Vector2,
	topRightTexture vector.Vector2,
	bottomRightTexture vector.Vector2,
) []mesh.Polygon {
	polys := make([]mesh.Polygon, 2)

	poly, _ := mesh.NewPolygonWithTexture(
		[]vector.Vector3{bottomLeft, topLeft, bottomRight},
		[]vector.Vector3{bottomLeft, topLeft, bottomRight},
		[]vector.Vector2{bottomLeftTexture, topLeftTexture, bottomRightTexture},
	)

	polys[0] = poly

	poly, _ = mesh.NewPolygonWithTexture(
		[]vector.Vector3{topLeft, topRight, bottomRight},
		[]vector.Vector3{topLeft, topRight, bottomRight},
		[]vector.Vector2{topLeftTexture, topRightTexture, bottomRightTexture},
	)

	polys[1] = poly
	return polys
}

func fill(shapes []mesh.Shape) ([]mesh.Polygon, error) {

	for _, shape := range shapes {
		if len(shape.GetPoints()) < 3 {
			return nil, errors.New("Can't make a polygon with less than 3 points")
		}
	}

	numOfPoints := 0
	pointsPrefixSum := make([]int, len(shapes))
	for i, shape := range shapes {
		pointsPrefixSum[i] = numOfPoints
		numOfPoints += len(shape.GetPoints())
	}

	flatPoints := make([][2]float64, numOfPoints)

	for shapeIndex, shape := range shapes {
		for pointIndex, point := range shape.GetPoints() {
			flatPoints[pointIndex+pointsPrefixSum[shapeIndex]][0] = point.X()
			flatPoints[pointIndex+pointsPrefixSum[shapeIndex]][1] = point.Y()
		}
	}

	segments := make([][2]int32, numOfPoints)

	for shapeIndex, shape := range shapes {
		for pointIndex := 0; pointIndex < len(shape.GetPoints()); pointIndex++ {
			i := pointIndex + pointsPrefixSum[shapeIndex]
			segments[i][0] = int32(i)
			segments[i][1] = int32(((pointIndex + 1) % len(shape.GetPoints())) + pointsPrefixSum[shapeIndex])
		}
	}

	// Hole represented by a point lying inside it
	var holes = make([][2]float64, len(shapes))
	for i := range shapes {
		holes[i][0] = 0
		holes[i][1] = 0.0
	}

	v, faces := triangle.ConstrainedDelaunay(flatPoints, segments, holes)

	betterPolys := make([]mesh.Polygon, len(faces))
	for i, face := range faces {
		ourVerts := make([]vector.Vector3, 0)
		ourVerts = append(ourVerts, vector.NewVector3(v[face[0]][0], 0, v[face[0]][1]))
		ourVerts = append(ourVerts, vector.NewVector3(v[face[1]][0], 0, v[face[1]][1]))
		ourVerts = append(ourVerts, vector.NewVector3(v[face[2]][0], 0, v[face[2]][1]))
		poly, _ := mesh.NewPolygon(ourVerts, ourVerts)
		betterPolys[i] = poly
	}
	return betterPolys, nil
}

func makeBottomPlate(resolution int, radius float64) []mesh.Polygon {
	polys := make([]mesh.Polygon, resolution)

	angleIncrement := (1.0 / float64(resolution)) * 2.0 * math.Pi
	for sideIndex := 0; sideIndex < resolution; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		angleNext := angleIncrement * (float64(sideIndex) + 1)

		points := []vector.Vector3{
			vector.NewVector3(math.Cos(angle)*radius, 0, math.Sin(angle)*radius),
			vector.NewVector3(math.Cos(angleNext)*radius, 0, math.Sin(angleNext)*radius),
			vector.NewVector3(0, 0, 0),
		}

		tex := []vector.Vector2{
			vector.NewVector2(0, 0),
			vector.NewVector2(1, 0),
			vector.NewVector2(1, 1),
		}

		poly, _ := mesh.NewPolygonWithTexture(points, points, tex)

		polys[sideIndex] = poly

	}

	return polys
}

func makeTopPlate(resolution int, radius, height float64) []mesh.Polygon {
	polys := make([]mesh.Polygon, resolution)

	angleIncrement := (1.0 / float64(resolution)) * 2.0 * math.Pi
	for sideIndex := 0; sideIndex < resolution; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		angleNext := angleIncrement * (float64(sideIndex) + 1)

		points := []vector.Vector3{
			vector.NewVector3(0, height, 0),
			vector.NewVector3(math.Cos(angleNext)*radius, height, math.Sin(angleNext)*radius),
			vector.NewVector3(math.Cos(angle)*radius, height, math.Sin(angle)*radius),
		}

		tex := []vector.Vector2{
			vector.NewVector2(0, 0),
			vector.NewVector2(1, 0),
			vector.NewVector2(1, 1),
		}

		poly, _ := mesh.NewPolygonWithTexture(points, points, tex)

		polys[sideIndex] = poly
	}

	return polys
}

func carve(width float64, height float64, shapes []mesh.Shape) ([]mesh.Polygon, error) {

	for _, shape := range shapes {
		if len(shape.GetPoints()) < 3 {
			return nil, errors.New("Can't make a polygon with less than 3 points")
		}
	}

	numOfPoints := 4
	pointsPrefixSum := make([]int, len(shapes))
	for i, shape := range shapes {
		pointsPrefixSum[i] = numOfPoints
		numOfPoints += len(shape.GetPoints())
	}

	flatPoints := make([][2]float64, numOfPoints)

	flatPoints[0] = [2]float64{0.0, 0.0}
	flatPoints[1] = [2]float64{0.0, height}
	flatPoints[2] = [2]float64{width, height}
	flatPoints[3] = [2]float64{width, 0.0}

	for shapeIndex, shape := range shapes {
		for pointIndex, point := range shape.GetPoints() {
			flatPoints[pointIndex+pointsPrefixSum[shapeIndex]][0] = point.X()
			flatPoints[pointIndex+pointsPrefixSum[shapeIndex]][1] = point.Y()
		}
	}

	segments := make([][2]int32, numOfPoints)

	segments[0] = [2]int32{0, 1}
	segments[1] = [2]int32{1, 2}
	segments[2] = [2]int32{2, 3}
	segments[3] = [2]int32{3, 0}

	for shapeIndex, shape := range shapes {
		for pointIndex := 0; pointIndex < len(shape.GetPoints()); pointIndex++ {
			i := pointIndex + pointsPrefixSum[shapeIndex]
			segments[i][0] = int32(i)
			segments[i][1] = int32(((pointIndex + 1) % len(shape.GetPoints())) + pointsPrefixSum[shapeIndex])
		}
	}

	// Hole represented by a point lying inside it
	var holes = make([][2]float64, len(shapes))
	for i, shape := range shapes {
		pointInShape := shape.RandomPointInShape()
		holes[i][0] = pointInShape.X()
		holes[i][1] = pointInShape.Y()
	}

	v, faces := triangle.ConstrainedDelaunay(flatPoints, segments, holes)

	betterPolys := make([]mesh.Polygon, len(faces))
	for i, face := range faces {
		ourVerts := make([]vector.Vector3, 3)
		ourVerts[0] = vector.NewVector3(v[face[0]][0], 0, v[face[0]][1])
		ourVerts[1] = vector.NewVector3(v[face[1]][0], 0, v[face[1]][1])
		ourVerts[2] = vector.NewVector3(v[face[2]][0], 0, v[face[2]][1])
		poly, _ := mesh.NewPolygon(ourVerts, ourVerts)
		betterPolys[i] = poly
	}
	return betterPolys, nil
}

// makeRing makes a single ring of faces.
func makeRing(resolution int, startingHeight, endingHeight, bottomRadius, topRadius float64) []mesh.Polygon {
	polys := make([]mesh.Polygon, resolution*2)

	numTimesForTextureToRepeat := 8

	angleIncrement := (1.0 / float64(resolution)) * 2.0 * math.Pi
	for sideIndex := 0; sideIndex < resolution; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		angleNext := angleIncrement * (float64(sideIndex) + 1)

		resPerTextToRepeat := resolution / numTimesForTextureToRepeat
		bottomLeftUV := vector.NewVector2(math.Min(float64(sideIndex)/float64(resPerTextToRepeat), 1), 0)
		topLeftUV := vector.NewVector2(math.Min(float64(sideIndex)/float64(resPerTextToRepeat), 1), 1.0)
		topRightUV := vector.NewVector2(math.Min(float64(sideIndex+1)/float64(resPerTextToRepeat), 1), 1.0)
		bottomRightUV := vector.NewVector2(math.Min(float64(sideIndex+1)/float64(resPerTextToRepeat), 1), 0)

		// outer
		square := makeSquareWithTexture(
			vector.NewVector3(math.Cos(angle)*bottomRadius, startingHeight, math.Sin(angle)*bottomRadius),
			vector.NewVector3(math.Cos(angle)*topRadius, endingHeight, math.Sin(angle)*topRadius),
			vector.NewVector3(math.Cos(angleNext)*topRadius, endingHeight, math.Sin(angleNext)*topRadius),
			vector.NewVector3(math.Cos(angleNext)*bottomRadius, startingHeight, math.Sin(angleNext)*bottomRadius),
			bottomLeftUV,
			topLeftUV,
			topRightUV,
			bottomRightUV,
		)

		polys[(sideIndex * 2)] = square[0]
		polys[(sideIndex*2)+1] = square[1]

	}

	return polys
}

// MakeMedalion creates a 3D object that represents a medal
func MakeMedalion(medalionThickness, designImpression float64) (mesh.Model, error) {

	defer timeTrack(time.Now(), "Creating Medal")

	startingRadius := 1.

	// How much extra radius will be added to the side of the medal as it bulges
	maxRadiusBulge := .1

	// how many rings we will use to aproximate the side of the medal bulging out
	bulgeResolution := 10

	// How many lines we will use to "draw" a circle
	sides := 64

	polys := make([]mesh.Polygon, 0)

	sinIncrement := 1. / float64(bulgeResolution)
	ringHeightIncrement := medalionThickness * sinIncrement
	for b := 0.; b < float64(bulgeResolution); b += 1.0 {
		// resolution int, startingHeight, endingHeight, bottomRadius, topRadius float64
		polys = append(polys, makeRing(
			sides,
			ringHeightIncrement*b,
			ringHeightIncrement*(b+1),
			(maxRadiusBulge*math.Sin(math.Pi*sinIncrement*b))+startingRadius,
			(math.Sin(math.Pi*sinIncrement*(b+1))*maxRadiusBulge)+startingRadius)...)
	}

	polys = append(polys, makeBottomPlate(sides, startingRadius)...)

	ringBorder := 0.05
	polys = append(polys, makeRing(sides, medalionThickness, medalionThickness, startingRadius, startingRadius-ringBorder)...)
	polys = append(polys, makeRing(sides, medalionThickness, medalionThickness-designImpression, startingRadius-ringBorder, startingRadius-ringBorder)...)
	polys = append(polys, makeTopPlate(sides, startingRadius-ringBorder, medalionThickness-designImpression)...)

	return mesh.NewModel(polys)
}

func TextToShape(textToWrite string) ([][]mesh.Shape, error) {

	defer timeTrack(time.Now(), fmt.Sprintf("Generating Text: %s", textToWrite))

	fontByteData, err := ioutil.ReadFile("./sample.ttf")

	if err != nil {
		return nil, err
	}

	parsedFont, err := truetype.Parse(fontByteData)

	if err != nil {
		return nil, err
	}

	finalWord := make([][]mesh.Shape, len(textToWrite))

	accumulatedWidth := 0.

	for charIndex, char := range textToWrite {

		glyph := truetype.GlyphBuf{}
		glyph.Load(parsedFont, 100, parsedFont.Index(char), font.HintingNone)

		letterPoints := make([]vector.Vector2, len(glyph.Points))
		for i, p := range glyph.Points {
			letterPoints[i] = vector.NewVector2(float64(p.X), float64(p.Y))
		}

		shape, err := mesh.NewShape(letterPoints)
		if err != nil {
			continue
		}

		shrunkShape := shape.Scale(.01)

		bottomLeftBounds, topRightBounds := shrunkShape.GetBounds()
		accumulatedWidth += (topRightBounds.X() - bottomLeftBounds.X())

		finalWord[charIndex] = []mesh.Shape{shrunkShape.Translate(vector.NewVector2(accumulatedWidth, 0))}
	}

	return finalWord, nil
}

func saveMedal(medal mesh.Model, name string) error {
	defer timeTrack(time.Now(), "Saving Medal")

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	err = medal.Save(w)
	if err != nil {
		return err
	}
	return w.Flush()
}

func ExtrudeShape(shapes []mesh.Shape, dist float64) (mesh.Model, error) {

	polys, err := fill(shapes)
	if err != nil {
		return mesh.Model{}, err
	}

	model, err := mesh.NewModel(polys)
	if err != nil {
		return mesh.Model{}, err
	}

	otherEnd := model.Translate(vector.NewVector3(0, dist, 0))

	stitching := make([]mesh.Polygon, 0)
	for polyIndex := 0; polyIndex < len(polys); polyIndex++ {

		verticesStart := polys[polyIndex].GetVertices()
		verticesEnd := otherEnd.GetFaces()[polyIndex].GetVertices()

		for v := 0; v < len(verticesStart); v++ {
			endingIndex := (v + 1) % len(verticesStart)
			stitching = append(stitching, makeSquareWithTexture(
				verticesStart[v],
				verticesEnd[v],
				verticesEnd[endingIndex],
				verticesStart[endingIndex],
				vector.NewVector2(0, 0),
				vector.NewVector2(0, 1),
				vector.NewVector2(1, 1),
				vector.NewVector2(1, 0),
			)...)
		}

	}

	stiches, err := mesh.NewModel(stitching)

	if err != nil {
		return mesh.Model{}, err
	}

	return model.Merge(otherEnd).Merge(stiches), nil
}

func TextToModel(text string, extrusion float64, letterShapeModifier func([][]mesh.Shape) []mesh.Shape) (mesh.Model, error) {
	letterShapes, err := TextToShape(text)
	if err != nil {
		return mesh.Model{}, err
	}

	model, err := ExtrudeShape(letterShapeModifier(letterShapes), extrusion)
	if err != nil {
		return mesh.Model{}, err
	}

	return model.Scale(vector.NewVector3(-.4, .4, .4), model.GetCenterOfBoundingBox()), nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func main() {

	medallionThickness := .6

	medallionImpression := .1

	medal, err := MakeMedalion(medallionThickness, medallionImpression)

	if err != nil {
		panic(err)
	}

	topText, err := TextToModel("Hello", medallionImpression, func(letters [][]mesh.Shape) []mesh.Shape {
		rotatedShapes := make([]mesh.Shape, 0)
		angleIncrements := math.Pi / float64(len(letters))

		for i, letterShape := range letters {
			curAngle := (math.Pi / 2.0) - (angleIncrements * float64(i)) - (angleIncrements / 2.0)
			centerOfShapes := mesh.CenterOfBoundingBoxOfShapes(letterShape)
			for _, shape := range letterShape {
				repositioned := shape.Translate(centerOfShapes.MultByConstant(-1).Add(vector.NewVector2(0.0, 1.6)))
				rotatedShapes = append(rotatedShapes, repositioned.Rotate(curAngle, vector.Vector2Zero()))
			}
		}
		return rotatedShapes
	})

	if err != nil {
		panic(err)
	}

	bottomText, err := TextToModel("dlroW", medallionImpression, func(letters [][]mesh.Shape) []mesh.Shape {
		rotatedShapes := make([]mesh.Shape, 0)
		angleIncrements := math.Pi / float64(len(letters))

		for i, letterShape := range letters {
			curAngle := (math.Pi / 2.0) - (angleIncrements * float64(i)) - (angleIncrements / 2.0)
			centerOfShapes := mesh.CenterOfBoundingBoxOfShapes(letterShape)
			for _, shape := range letterShape {
				repositioned := shape.Translate(centerOfShapes.MultByConstant(-1).Add(vector.NewVector2(0.0, -1.6)))
				rotatedShapes = append(rotatedShapes, repositioned.Rotate(curAngle, vector.Vector2Zero()))
			}
		}
		return rotatedShapes
	})

	if err != nil {
		panic(err)
	}

	topTextCentered := topText.Translate(vector.NewVector3(
		-topText.GetCenterOfBoundingBox().X(),
		medallionThickness-medallionImpression,
		-.5,
	))

	bottomTextCentered := bottomText.Translate(vector.NewVector3(
		-bottomText.GetCenterOfBoundingBox().X(),
		medallionThickness-medallionImpression,
		.5,
	))

	err = saveMedal(medal.Merge(topTextCentered).Merge(bottomTextCentered), "out.obj")

	if err != nil {
		panic(err)
	}

}
