package grpcConst

import (
	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
	"github.com/MikkelHJuul/grpcConst/merge"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"reflect"
	"testing"
)

type testClientStream struct {
	grpc.ClientStream
}

func (t *testClientStream) RecvMsg(v interface{}) error {
	return nil
}

func (t *testClientStream) Header() (metadata.MD, error) {
	return map[string][]string{
		XgRPCConst: []string{"EgQICxAW"},
	}, nil
}

type fields struct {
	ClientStream grpc.ClientStream
	merger       merge.Merger
}
type args struct {
	m interface{}
}

type testType struct {
	fields fields
	args   args
}

func benchmarkDataaddingclientstreamRecvmsgTest(tt testType, b *testing.B) {
	d := &dataAddingClientStream{
		ClientStream: tt.fields.ClientStream,
		merger:       tt.fields.merger,
	}
	for n := 0; n < b.N; n++ {
		if err := d.RecvMsg(tt.args.m); err != nil {
			b.Errorf("RecvMsg() error = %v", err)
		}
	}
}

func BenchmarkDataAddingClientStream_RecvMsgSimple(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{},
		merger:       merge.Merger{Initiated: false},
	},
		args{m: &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNil(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{},
		merger:       merge.Merger{Initiated: true},
	},
		args{m: &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNeverWrite(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{},
		merger:       merge.Merger{Initiated: false},
	},
		args{m: &proto.Feature{Name: "hey", Location: &proto.Point{Latitude: 12, Longitude: 21}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

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
