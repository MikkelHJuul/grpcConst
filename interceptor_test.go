package grpcConst

import (
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
	}{{
		name: "a simple obj header",
		args: args{&proto.Feature{
			Name: "",
			Location: &proto.Point{
				Latitude:  11,
				Longitude: 22,
			},
		}},
		want: map[string][]string{
			XgRPCConst: []string{"EgQICxAW"},
		},
		wantErr: false,
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
