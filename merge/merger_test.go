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
