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
		"array": []any{
			1,
			2,
			3,
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
		"array": []any{
			1,
			2,
			3,
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
	UnknownMap         	map[string]any                  `koanf:"unknownmap"   transform:"normal"`
	UnknownArray		[]any						    `koanf:"unknownarray" childtransform:"child"`
	StructMap      		map[string]Test_StructMapType	`koanf:"structmap"    childtransform:"child"`
	Struct				Test_StructType				    `koanf:"struct"`
}

type Test_StructMapType struct {
	Key 				string						    `koanf:"key"          transform:"normal"`
}

type Test_StructType struct {
	Key2 				string						    `koanf:"key2"         aliases:".key2"         transform:"normal"`
}

func TestTransformMapBuilder(t *testing.T) {
	transformTargets := configutils.BuildTransformMap("", &Test_StructSchema{})

	expected := map[string]configutils.TransformTarget{
		"unknownmap": {
			OutputKey: "unknownmap",
			Value: nil,
			ChildTransform: "",
			Transform: "normal",
		},
		"unknownarray": {
			OutputKey: "unknownarray",
			Value: nil,
			ChildTransform: "child",
			Transform: "",
		},
		"structmap": {
			OutputKey: "structmap",
			Value: nil,
			ChildTransform: "child",
			Transform: "",
		},
		"structmap.*.key": {
			OutputKey: "structmap.*.key",
			Value: "",
			ChildTransform: "",
			Transform: "normal",
		},
		"struct": {
			OutputKey: "struct",
			Value: Test_StructType{
				Key2: "",
			},
			ChildTransform: "",
			Transform: "",
		},
		"struct.key2": {
			OutputKey: "struct.key2",
			Value: "",
			ChildTransform: "",
			Transform: "normal",
		},
		"key2": {
			OutputKey: "struct.key2",
			Value: "",
			ChildTransform: "",
			Transform: "normal",
		},
	}

	transformTargetJson := jsonutils.Pretty(transformTargets)
	expectedJson := jsonutils.Pretty(expected)

	if transformTargetJson != expectedJson {
		t.Error("Expected: ", expectedJson, "\nGot: ", transformTargetJson)
	}
}

func TestTransform(t *testing.T) {
	data := map[string]any{
		"key2": "value2",
		"unknownmap": map[string]any{
			"key": "value",
		},
		"unknownarray": []any {
			1, 2, 3,
		},
		"structmap": map[string]any{
			"mapKey": map[string]any{
				"key": "value",
			},
		},
	}

	flattened := map[string]any{}

	configutils.Flatten("", data, flattened)

	transformTargets := configutils.BuildTransformMap("", &Test_StructSchema{})

	funcs := map[string]func(string, any) (string, any){
		"normal": func(s string, a any) (string, any) {
			return "normal:" + s, a
		},
		"child": func(s string, a any) (string, any) {
			return "child:" + s, a
		},
	}

	transformed := configutils.ApplyTransforms(flattened, transformTargets, funcs)

	unflattened := configutils.Unflatten(transformed)

	expected := map[string]any{
		"structmap": map[string]any{
            "child:mapkey": map[string]any{
              "normal:key": "value",
            },
		},
		"normal:unknownmap": map[string]any{
			"key": "value",
		},
		"struct": map[string]any{
			"normal:key2": "value2",
		},
		"unknownarray": map[string]any {
			"child:0": 1,
			"child:1": 2,
			"child:2": 3,
		},
	}

	transformedJson := jsonutils.Pretty(unflattened)
	expectedJson := jsonutils.Pretty(expected)

	if transformedJson != expectedJson {
		t.Error("Expected: ", expectedJson, "\nGot: ", transformedJson)
	}
}