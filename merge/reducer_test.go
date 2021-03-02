package merge

import (
	ogcish "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"reflect"
	"testing"
)

func TestReducerReduces(t *testing.T) {
	reference, subject, result :=
		&testStruct{
			Obj: "hello",
		},
		&testStruct{
			Obj: "hello",
			Sub: nested{1},
		},
		&testStruct{
			Sub: nested{1},
		}
	r := NewReducer(reference)
	_ = r.RemoveFields(subject)
	if subject.Obj != result.Obj || subject.Sub.Some != result.Sub.Some {
		t.Error("The objects are not equal")
	}
}

func TestReducer_RemoveFields(t *testing.T) {
	tests := []struct {
		name     string
		receiver interface{}
		donor    interface{}
		result   interface{}
	}{
		{
			name: "Test ogcIsh reduce - inefficient reduce",
			donor: &ogcish.Feature{Geometry: &ogcish.Geometry{
				Type: "origin",
				Coordinates: &ogcish.Point{
					Latitude:  0,
					Longitude: 0,
				},
			}},
			result: &ogcish.Feature{ // reduces to empty objects, not Geometry: nil
				Type:     "Top",
				Id:       "uuid",
				Geometry: &ogcish.Geometry{Coordinates: &ogcish.Point{}},
			},
			receiver: &ogcish.Feature{
				Type: "Top",
				Id:   "uuid",
				Geometry: &ogcish.Geometry{
					Type: "origin",
					Coordinates: &ogcish.Point{
						Latitude:  0,
						Longitude: 0,
					},
				},
			},
		},
		{
			name: "Test nested interface{} (simple type)", // false positive! nested{1} is reduced as any would
			donor: &testStruct{
				Sub: nested{1},
			},
			result: &testStruct{
				Obj: "asd",
			},
			receiver: &testStruct{
				Obj: "asd",
				Sub: nested{1},
			},
		},
		{
			name: "Reducer mangles interfaces", // true negative nested obj incorrectly reduced
			result: &testStruct{
				Sub: nested{nil},
			},
			donor: &testStruct{
				Obj: "hello",
				Sub: nested{"Hello"},
			},
			receiver: &testStruct{
				Obj: "hello",
				Sub: nested{1},
			},
		},
		{
			name: "insert intermediate interface object", // false positive again
			donor: &testWithInterface{
				NonInterface: "hello",
				InterfaceObj: testWithInterface{"HelloNested", 1},
			},
			result: &testWithInterface{
				NonInterface: "he11o",
			},
			receiver: &testWithInterface{
				NonInterface: "he11o",
				InterfaceObj: testWithInterface{"HelloNested", 1},
			},
		},
		{
			name: "Reducer Breaks interfaces", // true negative
			donor: &testWithInterface{
				NonInterface: "hello",
				InterfaceObj: testWithInterface{"HelloNested", nil},
			},
			result: &testWithInterface{
				NonInterface: "he11o",
				InterfaceObj: nil,
			},
			receiver: &testWithInterface{ //helloNested is not merged!
				NonInterface: "he11o",
				InterfaceObj: testWithInterface{"", 1},
			},
		},
		{
			name: "Test with a Map", // true negative!
			donor: &objWithMap{
				Obj: map[string]string{"hell": "world"},
			},
			result: &objWithMap{
				Name: "hello",
			},
			receiver: &objWithMap{
				Name: "hello",
				Obj:  map[string]string{"hell1": "world"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewReducer(tt.donor)
			if err := m.RemoveFields(tt.receiver); err != nil {
				t.Errorf("SetFields() error = %v", err)
			}
			if !reflect.DeepEqual(tt.receiver, tt.result) {
				t.Error("The objects are not equal")
			}
		})
	}
}
