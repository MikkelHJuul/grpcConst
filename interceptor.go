//grpcConst is a package that allow you to communicate defaulting values of your protobuf messages
//and communicate this default and set it before your client side code interacts with the messages.
//example server-side:
//		header, err := grpcConst.HeaderSetConstant(
//				&proto.Feature{
//					Name: "some constant name",
//					Location: &proto.Point{Latitude: 10}
//		})
//		stream.SetHeader(header)
//			... your normal routine but you could
//			... fx send &proto.Feature{Location: &proto.Point{Longitude: 20}}
//			... this will yield - name: "some constant name", location: {10, 20}
//			... while sending less data in the message
//or:
//      stream = grpcConst.ServerStreMWrapper(
//				&proto.Feature{
//					Name: "some constant name",
//					Location: &proto.Point{Latitude: 10}
//		})
//			... using stream.Send() now removes the default values from your objects; sending less data
//example client-side:
//initiate your client with a grpc.StreamClientInterceptor this way:
// 		conn, err := grpc.Dial(...,  grpc.WithStreamInterceptor(grpcConst.StreamClientInterceptor()))
package grpcConst

import (
	"context"
	"encoding/base64"
	"log"
	"reflect"

	"github.com/MikkelHJuul/grpcConst/merge"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
)

//XgRPCConst is the HTTP header passed between server and client
const XgRPCConst = "x-grpc-const"

//HeaderSetConstant is a convenience method for the server side to add a metadata.MD with the correct content
// given your gRPC struct v, the user is returned the metadata to send.
//that the user can send using `grpc.ServerStream:SendHeader(metadata.MD) or :SetHeader(metadata.MD)`.
// v must be passed by reference.
func HeaderSetConstant(v interface{}) (metadata.MD, error) {
	msg, err := marshal(v)
	return metadata.Pairs(XgRPCConst, msg), err
}

//ServerStreamWrapper wraps your stream object and returns the decorated stream with a SendMsg method,
//that removes items that are equal a reference object.
//The stream remains untouched if the client did not send an XgRPCConst header
func ServerStreamWrapper(reference interface{}) (stream grpc.ServerStream, err error) {
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok && len(md.Get(XgRPCConst)) > 0 {
		md, err = HeaderSetConstant(reference)
		if err != nil {
			return
		}
		stream.SetHeader(md)
		var reducer merge.Reducer
		if _, ok := reference.(Reducer); ok {
			reducer = MessageMergerReducer{ConstantMessage: reference}
		} else {
			reducer = merge.NewReducer(reference)
		}
		stream = &dataRemovingServerStream{stream, reducer}
	}
	return
}

//StreamClientInterceptor is an interceptor for the client side (for unidirectional server-side streaming rpc's)
//The client side Stream interceptor intercepts the stream when it is initiated.
//This method decorates the actual ClientStream adding data to each message where applicable
//this variadic function accepts none or one argument. defaulting the method for constructing
//the merge.Merger to use merge.NewMerger.
//for a more safe alternative
func StreamClientInterceptor(mergerCreator ...MergerCreator) grpc.StreamClientInterceptor {
	mergeCreator := mergerCreatorDefaulting(mergerCreator...)
	return func(
		parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx := metadata.AppendToOutgoingContext(parentCtx, XgRPCConst, "")
		var stream, err = streamer(ctx, desc, cc, method, opts...)
		return &dataAddingClientStream{stream, nil, mergeCreator}, err
	}
}

func mergerCreatorDefaulting(creator ...MergerCreator) MergerCreator {
	if creator == nil || creator[0] == nil {
		return merge.NewMerger
	} else {
		return creator[0]
	}
}

//MergerCreator is a proxy for a function returning a merge.Merger from an interface{}
type MergerCreator func(interface{}) merge.Merger

//marshal implements the server side marshalling of a protobuf message into the specification header value
func marshal(v interface{}) (string, error) {
	msg, err := encoding.GetCodec("proto").Marshal(v)
	return base64.URLEncoding.EncodeToString(msg), err
}

//unmarshal implements the client side handling/unmarshalling of the specification header
func unmarshal(header string, receiver interface{}) error {
	protoMsg, err := base64.URLEncoding.DecodeString(header)
	if err != nil {
		return err
	}
	return encoding.GetCodec("proto").Unmarshal(protoMsg, receiver)
}

//dataAddingClientStream is the decorated grpc.ClientStream
//that your code will use via the method RecvMsg
//the intermediary construct fieldToSet is used to remove to need to traverse the entire message
type dataAddingClientStream struct {
	grpc.ClientStream
	Merger        merge.Merger
	mergerCreator MergerCreator
}

type dataRemovingServerStream struct {
	grpc.ServerStream
	Reducer merge.Reducer
}

//RecvMsg is called via your grpc.ClientStream;
//the generated code's method Recv calls this method on it's internal grpc.ClientStream
//This method initiates on first call the fields that should be default to all the messages
//on all calls the underlying grpc.ClientStream:RecvMsg message has this data added
func (dc *dataAddingClientStream) RecvMsg(m interface{}) error {
	if dc.Merger == nil {
		donor := newEmpty(m)
		header, _ := dc.ClientStream.Header()
		if head, ok := header[XgRPCConst]; ok && len(head) > 0 {
			if err := unmarshal(head[0], donor); err != nil {
				log.Printf("ERROR: an %s-header could not be unmarshalled correctly: %v", XgRPCConst, head)
			}
		}
		if _, ok := donor.(Merger); ok {
			dc.Merger = MessageMergerReducer{ConstantMessage: donor}
		} else {
			dc.Merger = dc.mergerCreator(donor)
		}
	}
	if err := dc.ClientStream.RecvMsg(m); err != nil {
		return err
	}
	return dc.Merger.SetFields(m)
}

//newEmpty simply creates a new instance of an interface given an instance of that interface
func newEmpty(t interface{}) interface{} {
	return reflect.New(reflect.TypeOf(t).Elem()).Interface()
}

//SendMsg reduces the message using the reference before sending it using the underlying ServerStream
func (ds *dataRemovingServerStream) SendMsg(m interface{}) error {
	if err := ds.Reducer.RemoveFields(m); err != nil {
		log.Printf("ERROR: could not remove fields from %v", m)
	}
	return ds.ServerStream.SendMsg(m)
}
