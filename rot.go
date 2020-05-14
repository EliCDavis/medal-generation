package main

import (
	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

func Rotate(m mesh.Model, pivot vector.Vector3, rot mesh.Quaternion) mesh.Model {

	polysRotated := make([]mesh.Polygon, 0)

	for _, f := range m.GetFaces() {
		vertRotated := make([]vector.Vector3, 0)
		for _, v := range f.GetVertices() {
			vertRotated = append(vertRotated, rot.Rotate(v.Sub(pivot)).Add(pivot))
		}
		p, _ := mesh.NewPolygon(vertRotated, vertRotated)
		polysRotated = append(polysRotated, p)
	}

	// for i := 0; i < len(ls); i++ {
	// 	results[i] = rot.Rotate(ls[i].Sub(pivot)).Add(pivot)
	// }

	mRot, _ := mesh.NewModel(polysRotated)

	return mRot
}
