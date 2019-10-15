package main

import (
	"bufio"
	"errors"
	"log"
	"math"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
	"github.com/pradeep-pyro/triangle"
)

func check(e error) {
	if e != nil {
		log.Panicln(e.Error())
		panic(e)
	}
}

func makeSquare(
	bottomLeft vector.Vector3,
	topLeft vector.Vector3,
	topRight vector.Vector3,
	bottomRight vector.Vector3,

	bottomLeftTexture vector.Vector2,
	topLeftTexture vector.Vector2,
	topRightTexture vector.Vector2,
	bottomRightTexture vector.Vector2,
) ([]mesh.Polygon, error) {
	polys := make([]mesh.Polygon, 2)

	poly, err := mesh.NewPolygonWithTexture(
		[]vector.Vector3{bottomLeft, topLeft, bottomRight},
		[]vector.Vector3{bottomLeft, topLeft, bottomRight},
		[]vector.Vector2{bottomLeftTexture, topLeftTexture, bottomRightTexture},
	)
	if err != nil {
		return nil, err
	}
	polys[0] = poly

	poly, err = mesh.NewPolygonWithTexture(
		[]vector.Vector3{topLeft, topRight, bottomRight},
		[]vector.Vector3{topLeft, topRight, bottomRight},
		[]vector.Vector2{topLeftTexture, topRightTexture, bottomRightTexture},
	)
	if err != nil {
		return nil, err
	}
	polys[1] = poly
	return polys, nil
}

func fill(width float64, height float64, shapes []mesh.Shape) ([]mesh.Polygon, error) {

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
		square, err := makeSquare(
			vector.NewVector3(math.Cos(angle)*bottomRadius, startingHeight, math.Sin(angle)*bottomRadius),
			vector.NewVector3(math.Cos(angle)*topRadius, endingHeight, math.Sin(angle)*topRadius),
			vector.NewVector3(math.Cos(angleNext)*topRadius, endingHeight, math.Sin(angleNext)*topRadius),
			vector.NewVector3(math.Cos(angleNext)*bottomRadius, startingHeight, math.Sin(angleNext)*bottomRadius),
			bottomLeftUV,
			topLeftUV,
			topRightUV,
			bottomRightUV,
		)

		check(err)
		polys[(sideIndex * 2)] = square[0]
		polys[(sideIndex*2)+1] = square[1]

	}

	return polys
}

func main() {

	startingRadius := 1.

	// How much extra radius will be added to the side of the medal as it bulges
	maxRadiusBulge := .2

	// how many rings we will use to aproximate the side of the medal bulging out
	bulgeResolution := 10

	ringHeight := .8

	// How many lines we will use to "draw" a circle
	sides := 64

	polys := make([]mesh.Polygon, 0)

	sinIncrement := 1. / float64(bulgeResolution)
	ringHeightIncrement := ringHeight * sinIncrement
	for b := 0.; b < float64(bulgeResolution); b += 1.0 {
		// resolution int, startingHeight, endingHeight, bottomRadius, topRadius float64
		polys = append(polys, makeRing(
			sides,
			ringHeightIncrement*b,
			ringHeightIncrement*(b+1),
			(maxRadiusBulge*math.Sin(math.Pi*sinIncrement*b))+startingRadius,
			(math.Sin(math.Pi*sinIncrement*(b+1))*maxRadiusBulge)+startingRadius)...)
	}

	medalModel, err := mesh.NewModel(polys)
	check(err)

	log.Println("completed")

	f, err := os.Create("out.obj")
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	medalModel.Save(w)
	w.Flush()

}
