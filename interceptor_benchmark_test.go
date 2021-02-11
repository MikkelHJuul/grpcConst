package grpcConst

import (
	ogcIsh "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
	"github.com/MikkelHJuul/grpcConst/merge"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

type testClientStream struct {
	grpc.ClientStream
	header string
}

func (t *testClientStream) RecvMsg(_ interface{}) error {
	return nil
}

func (t *testClientStream) Header() (metadata.MD, error) {
	return map[string][]string{
		XgRPCConst: {t.header},
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
		Merger:       tt.fields.merger,
	}
	for n := 0; n < b.N; n++ {
		if err := d.RecvMsg(tt.args.m); err != nil {
			b.Errorf("RecvMsg() error = %v", err)
		}
	}
}

func BenchmarkDataAddingClientStream_RecvMsgSimple(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "EgQICxAW"},
		merger:       nil,
	},
		args{m: &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNil(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "EgQICxAW"},
		merger:       merge.NewMerger(&proto.Feature{}),
	},
		args{m: &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNeverWrite(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "EgQICxAW"},
		merger:       nil,
	},
		args{m: &proto.Feature{Name: "hey", Location: &proto.Point{Latitude: 12, Longitude: 21}}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgLarger(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="},
		merger:       nil},
		args{m: &ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}}},
	}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkInitiation(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stream := &dataAddingClientStream{
			&testClientStream{header: "CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="},
			nil}
		_ = stream.RecvMsg(&ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}})
	}
}
