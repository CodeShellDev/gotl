package tests

import (
	"testing"

	"github.com/codeshelldev/gotl/pkg/configutils"
	"github.com/codeshelldev/gotl/pkg/jsonutils"
)

func TestConfigUnflattening(t *testing.T) {
	flat := map[string]any{
		"data.key": "value",
		"array.0": 1,
		"array.1": 2,
		"array.2": 3,
		"dict.data.key": "value",
	}

	unflattened := configutils.Unflatten(flat)

	expected := map[string]any{
		"data": map[string]any{
			"key": "value",
		},
		"array": map[string]any{
			"0": 1,
			"1": 2,
			"2": 3,
		},
		"dict": map[string]any{
			"data": map[string]any{
				"key": "value",
			},
		},
	}

	unflattenedJson := jsonutils.Pretty(unflattened)
	expectedJson := jsonutils.Pretty(expected)

	if unflattenedJson != expectedJson {
		t.Error("Expected: ", expectedJson, "\nGot: ", unflattenedJson)
	}
}

func TestConfigFlattening(t *testing.T) {
	unflattened := map[string]any{
		"data": map[string]any{
			"key": "value",
		},
		"array": map[string]any{
			"0": 1,
			"1": 2,
			"2": 3,
		},
		"dict": map[string]any{
			"data": map[string]any{
				"key": "value",
			},
		},
	}

	flattened := map[string]any{}
	
	configutils.Flatten("", unflattened, flattened)

	expected := map[string]any{
		"data.key": "value",
		"array.0": 1,
		"array.1": 2,
		"array.2": 3,
		"dict.data.key": "value",
	}

	flattenedJson := jsonutils.Pretty(flattened)
	expectedJson := jsonutils.Pretty(expected)

	if flattenedJson != expectedJson {
		t.Error("Expected: ", expectedJson, "\nGot: ", flattenedJson)
	}
}

type Test_StructSchema struct {
	UnknownMap         	map[string]any              `koanf:"unknownmap"  transform:"default"`
	StructMap      		map[string]Test_StructType	`koanf:"structmap"   childtransform:"child"`
}

type Test_StructType struct {
	Key 				string						`koanf:"key"         aliases:".key"`
}

func TestTransformMapBuilder(t *testing.T) {
	transformTargets := configutils.BuildTransformMap("", &Test_StructSchema{})

	expected := map[string]configutils.TransformTarget{
		"unknownmap": {
			OutputKey: "unknownmap",
			Value: nil,
			ChildTransform: "",
			Transform: "default",
		},
		"structmap": {
			OutputKey: "structmap",
			Value: nil,
			ChildTransform: "child",
			Transform: "",
		},
		"structmap.key": {
			OutputKey: "structmap.key",
			Value: "",
			ChildTransform: "",
			Transform: "",
		},
		"key": {
			OutputKey: "structmap.key",
			Value: "",
			ChildTransform: "",
			Transform: "",
		},
	}

	transformTargetJson := jsonutils.Pretty(transformTargets)
	expectedJson := jsonutils.Pretty(expected)

	if transformTargetJson != expectedJson {
		t.Error("Expected: ", expectedJson, "\nGot: ", transformTargetJson)
	}
}