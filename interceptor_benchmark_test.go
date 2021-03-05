package grpcConst

import (
	"testing"

	ogcIsh "github.com/MikkelHJuul/grpcConst/examples/ogc_ish/proto"
	"github.com/MikkelHJuul/grpcConst/examples/route_guide/proto"
	"github.com/MikkelHJuul/grpcConst/merge"

	gogoProto "github.com/gogo/protobuf/proto"
	pb "github.com/gogo/protobuf/proto/test_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	goProto "google.golang.org/protobuf/proto"
)

//It seem like there is an overhead of about 100-150 ns just from getting and assigning objects every loop
//
//BenchmarkDataAddingClientStream_RecvMsgSimple
//BenchmarkDataAddingClientStream_RecvMsgSimple-8       	 5831353	       189 ns/op
//BenchmarkDataAddingClientStream_RecvMsgNil
//BenchmarkDataAddingClientStream_RecvMsgNil-8          	 5951642	       190 ns/op
//BenchmarkDataAddingClientStream_RecvMsgNeverWrite
//BenchmarkDataAddingClientStream_RecvMsgNeverWrite-8   	 6063172	       193 ns/op
//BenchmarkDataAddingClientStream_RecvMsgLarger
//BenchmarkDataAddingClientStream_RecvMsgLarger-8       	 2064686	       568 ns/op  (170 ns for preCompiled)
//BenchmarkProtoMergeMerger
//BenchmarkProtoMergeMerger-8                           	 2121768	       566 ns/op
//BenchmarkInitiation
//BenchmarkInitiation-8                                 	  263386	      4624 ns/op
//BenchmarkPreCompiled
//BenchmarkPreCompiled-8                                	56600329	        19.5 ns/op
//BenchmarkProtoMerge
//BenchmarkProtoMerge-8                                 	 2237928	       529 ns/op
//BenchmarkGoGoProtoMerge
//BenchmarkGoGoProtoMerge-8                             	 1000000	      1150 ns/op
//
type testClientStream struct {
	grpc.ClientStream
	header string
}

func (t *testClientStream) RecvMsg(m interface{}) error {
	return nil
}

func (t *testClientStream) Header() (metadata.MD, error) {
	return map[string][]string{
		XgRPCConst: {t.header},
	}, nil
}

type fields struct {
	ClientStream grpc.ClientStream
	creator      MergerCreator
}
type args struct {
	m func() interface{}
}

type testType struct {
	fields fields
	args   args
}

func benchmarkDataaddingclientstreamRecvmsgTest(tt testType, b *testing.B) {
	d := &dataAddingClientStream{
		ClientStream:  tt.fields.ClientStream,
		mergerCreator: mergerCreatorDefaulting(tt.fields.creator),
	}
	for n := 0; n < b.N; n++ {
		if err := d.RecvMsg(tt.args.m()); err != nil {
			b.Errorf("RecvMsg() error = %v", err)
		}
	}
}

func BenchmarkDataAddingClientStream_RecvMsgSimple(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "CgdEZUZBVUxUIgYSBAgMEBU="},
	},
		args{m: func() interface{} { return &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}} }},
	}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNil(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: ""},
	},
		args{m: func() interface{} { return &proto.Feature{Name: "", Location: &proto.Point{Latitude: 0, Longitude: 0}} }}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgNeverWrite(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "CgdEZUZBVUxUIgYSBAgMEBU="},
	},
		args{m: func() interface{} {
			return &ogcIsh.Feature{Id: "hey", Geometry: &ogcIsh.Geometry{Coordinates: &ogcIsh.Point{Latitude: 12, Longitude: 21}}}
		}}}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkDataAddingClientStream_RecvMsgLarger(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="}},
		args{m: func() interface{} {
			return &ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}}
		}},
	}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkProtoMergeMerger(b *testing.B) {
	tests := testType{fields{
		ClientStream: &testClientStream{header: "CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="},
		creator:      merge.NewProtoMerger},
		args{m: func() interface{} {
			return &ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}}
		}},
	}
	benchmarkDataaddingclientstreamRecvmsgTest(tests, b)
}

func BenchmarkInitiation(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stream := &dataAddingClientStream{
			&testClientStream{header: "CgdGZWF0dXJlGkUKBgoESm9oblI7ChFTb21lIFN0YXRpb24gTmFtZRImU29tZSBzdGF0aW9uJ3MgbWV0YWRhdGEsIGEgc2hvcnQgc3RvcnkiDAoDTG9sEgUIexDBAg=="},
			nil, merge.NewMerger}
		_ = stream.RecvMsg(&ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}})
	}
}

func BenchmarkPreCompiled(b *testing.B) {
	f := &proto.Feature{
		Name: "Hello world",
		Location: &proto.Point{
			Latitude:  123,
			Longitude: 321,
		},
	}
	for n := 0; n < b.N; n++ {
		r := &proto.Feature{Location: &proto.Point{
			Latitude: 11,
		}}
		r.Merge(*f)
	}
}

func BenchmarkProtoMerge(b *testing.B) {
	r := &ogcIsh.Feature{
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
	}
	for n := 0; n < b.N; n++ {
		f := &ogcIsh.Feature{Properties: &ogcIsh.Properties{Measurement: &ogcIsh.Measurement{Value: 666}}}
		goProto.Merge(f, r)
	}
}

func BenchmarkGoGoProtoMerge(b *testing.B) {
	f := &pb.MyMessage{
		Inner: &pb.InnerMessage{
			Host:      gogoProto.String("hey"),
			Connected: gogoProto.Bool(true),
		},
		Pet: []string{"horsey"},
		Others: []*pb.OtherMessage{
			{
				Value: []byte("some bytes"),
			},
		},
	}
	for n := 0; n < b.N; n++ {
		gogoProto.Merge(&pb.MyMessage{
			Inner: &pb.InnerMessage{
				Host: gogoProto.String("niles"),
				Port: gogoProto.Int32(9099),
			},
			Pet: []string{"bunny", "kitty"},
			Others: []*pb.OtherMessage{
				{
					Key: gogoProto.Int64(31415926535),
				},
				{
					// Explicitly test a src=nil field
					Inner: nil,
				},
			},
		}, f)
	}
}
