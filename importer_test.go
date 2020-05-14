package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstWords(t *testing.T) {
	first, rest := firstWords("v 1.030902 0.060000 0.000000", 1)

	assert.Equal(t, first, "v")
	assert.Equal(t, rest, "1.030902 0.060000 0.000000")
}

func TestParseVector(t *testing.T) {
	v, err := strToVector("1.030902 0.060000 0.000000")

	assert.NoError(t, err)
	if assert.NotNil(t, v) {
		assert.Equal(t, 1.030902, v.X())
		assert.Equal(t, 0.060000, v.Y())
		assert.Equal(t, 0.000000, v.Z())
	}
}

func TestParseFaceIndexes(t *testing.T) {
	v1, v2, v3, err := strToFaceIndexes("10//10 11//11 12//12")

	assert.NoError(t, err)
	assert.Equal(t, 10, v1)
	assert.Equal(t, 11, v2)
	assert.Equal(t, 12, v3)
}

func TestImportObj(t *testing.T) {

	obj := strings.Builder{}
	obj.WriteString("v 1.000000 0.000000 0.000000\n")
	obj.WriteString("vn 1.000000 0.000000 0.000000\n")
	obj.WriteString("vt 0.000000 0.000000 \n")
	obj.WriteString("v 1.030902 0.060000 0.000000\n")
	obj.WriteString("vn 1.030902 0.060000 0.000000\n")
	obj.WriteString("vt 0.000000 1.000000 \n")
	obj.WriteString("v 0.995185 0.000000 0.098017\n")
	obj.WriteString("vn 0.995185 0.000000 0.098017\n")
	obj.WriteString("vt 0.125000 0.000000 \n")
	obj.WriteString("f 1/1 2/2 3/3 \n")
	obj.WriteString("v 1.030902 0.060000 0.000000\n")
	obj.WriteString("vn 1.030902 0.060000 0.000000\n")
	obj.WriteString("vt 0.000000 1.000000 \n")
	obj.WriteString("v 1.025938 0.060000 0.101046\n")
	obj.WriteString("vn 1.025938 0.060000 0.101046\n")
	obj.WriteString("vt 0.125000 1.000000 \n")
	obj.WriteString("v 0.995185 0.000000 0.098017\n")
	obj.WriteString("vn 0.995185 0.000000 0.098017\n")
	obj.WriteString("vt 0.125000 0.000000 \n")
	obj.WriteString("f 4/4 5/5 6/6 \n")

	model, err := importOBJ(strings.NewReader(obj.String()))

	assert.NoError(t, err)
	if assert.NotNil(t, model) {
		// assert.Equal(t, 2, model.``)
	}
}
