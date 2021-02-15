package grpcConst

import (
	ogcIsh "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
	"google.golang.org/grpc/metadata"
	"reflect"
	"testing"
)

func TestHeaderSetConstant(t *testing.T) {
	type args struct {
		v interface{}
	}
	var tests = []struct {
		name    string
		args    args
		want    metadata.MD
		wantErr bool
	}{
		{
			name: "a simple obj header",
			args: args{&proto.Feature{
				Name: "",
				Location: &proto.Point{
					Latitude:  11,
					Longitude: 22,
				},
			}},
			want: map[string][]string{
				XgRPCConst: {"EgQICxAW"},
			},
			wantErr: false,
		},
		{
			name: "a large obj header",
			args: args{&ogcIsh.Feature{
				Type: "Feature",
				Properties: &ogcIsh.Properties{
					Measurement: &ogcIsh.Measurement{
						Name: "John",
					},
					Station: &ogcIsh.Station{
						Name:     "Some Station Name",
						Metadata: "Some station's metadata, a short story",
					},
				},
				Geometry: &ogcIsh.Geometry{
					Type: "Lol",
					Coordinates: &ogcIsh.Point{
						Latitude:  123,
						Longitude: 321,
					},
				},
			}},
			want: map[string][]string{XgRPCConst: {"CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="}},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HeaderSetConstant(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("HeaderSetConstant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HeaderSetConstant() got = %v, want %v", got, tt.want)
			}
		})
	}
}
