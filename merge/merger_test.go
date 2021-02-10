package merge

import (
	"reflect"
	"testing"

	ogcish "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
)

type nested struct {
	Some int
}
type testStruct struct {
	Obj string
	Sub nested
}

func TestMerger_SetFields(t *testing.T) {
	tests := []struct {
		name     string
		receiver interface{}
		donor    interface{}
		result   interface{}
		wantErr  bool
	}{
		{
			name: "",
			donor: &ogcish.Feature{Geometry: &ogcish.Geometry{
				Type: "origin",
				Coordinates: &ogcish.Point{
					Latitude:  0,
					Longitude: 0,
				},
			}},
			receiver: &ogcish.Feature{
				Type:     "Top",
				Id:       "uuid",
				Geometry: nil,
			},
			result: &ogcish.Feature{
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
			name: "",
			donor: &testStruct{
				Sub: nested{1},
			},
			receiver: &testStruct{
				Obj: "asd",
			},
			result: &testStruct{
				Obj: "asd",
				Sub: nested{1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMerger(tt.donor)
			if err := m.SetFields(tt.receiver); (err != nil) != tt.wantErr {
				t.Errorf("SetFields() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.receiver, tt.result) {
				t.Error("The objects are not equal")
			}
		})
	}
}

func TestDataIsDeepCopied(t *testing.T) {
	donor, receiver, result :=
		&testStruct{
			Sub: nested{1},
		},
		&testStruct{
			Obj: "hello",
		},
		&testStruct{
			Obj: "hello",
			Sub: nested{1},
		}
	m := NewMerger(donor)
	_ = m.SetFields(receiver)
	if receiver.Obj != result.Obj || receiver.Sub.Some != result.Sub.Some {
		t.Error("The objects are not equal")
	}
	donor.Sub.Some = 3
	if receiver.Obj != result.Obj || receiver.Sub.Some != result.Sub.Some {
		t.Error("The objects are not equal")
	}
}

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

func TestMergeMap(t *testing.T) {
	type objWithMap struct {
		Name string
		Obj  map[string]string
	}
	donor, receiver, result :=
		&objWithMap{
			Obj: map[string]string{"hell": "world"},
		},
		&objWithMap{
			Name: "hello",
		},
		&objWithMap{
			Name: "hello",
			Obj:  map[string]string{"hell": "world"},
		}
	m := NewMerger(donor)
	_ = m.SetFields(receiver)
	if receiver.Obj["hell"] != result.Obj["hell"] || receiver.Name != result.Name {
		t.Error("The objects are not equal")
	}
}
