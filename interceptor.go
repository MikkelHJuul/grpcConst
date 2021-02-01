//grpcConst is a package that allow you to communicate defaulting values of your protobuf messages
//and communicate this default and set it before your client side code interacts with the messages.
//examples:
//	server-side:
// 		func (t *myRPCServer) MyServerStreamRPC(req *proto.Request, stream proto.myRPCServer_MyServerStreamRPCServer) error {
//			header, err := grpcConst.HeaderSetConstant(
//							&proto.Feature{
//								Name: "some constant name",
//								Location: &proto.Point{Latitude: 10}
//					 	})
//			stream.SetHeader(header)
//			... your normal routine but you could
//			... fx send &proto.Feature{Location: &proto.Point{Longitude: 20}}
//			... this will yield - name: "some constant name", location: {10, 20}
//			... while sending less data in the message
//		}
//	client-side:
//		initiate your client with a grpc.StreamClientInterceptor this way:
// 		conn, err := grpc.Dial(...,  grpc.WithStreamInterceptor(grpcConst.StreamClientInterceptor()))
package grpcConst

import (
	"context"
	"encoding/base64"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"reflect"
)

//XgRPCConst is the HTTP header passed between server and client
const XgRPCConst = "x-grpc-const"

//HeaderSetConstant is a convenience method for the server side to add a metadata.MD with the correct content
// given your gRPC struct v, the user is returned the metadata to send.
//that the user can send using `grpc.ServerStream:SendHeader(metadate.MD) or :SetHeader(metadate.MD)`.
func HeaderSetConstant(v interface{}) (metadata.MD, error) {
	msg, err := marshal(v)
	return metadata.Pairs(XgRPCConst, msg), err
}

//StreamClientInterceptor is an interceptor for the client side (for unidirectional server-side streaming rpc's)
//The client side Stream interceptor intercepts the stream when it is initiated. This method decorates the actual ClientStream
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var stream, err = streamer(parentCtx, desc, cc, method, opts...)
		return &dataAddingClientStream{stream, nil, nil}, err
	}
}

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
	if err := encoding.GetCodec("proto").Unmarshal(protoMsg, receiver); err != nil {
		return err
	}
	return nil
}

//dataAddingClientStream is the decorated grpc.ClientStream
//that your code will use via the method RecvMsg
//the intermediary construct fieldToSet is used to remove to need to traverse the entire message
type dataAddingClientStream struct {
	grpc.ClientStream
	constantMessage interface{}
	fieldsToSet     map[reflect.Value][]int
}

//RecvMsg is called via your grpc.ClientStream;
//the generated code's method Recv calls this method on it's internal grpc.ClientStream
//This method initiates on first call the fields that should be default to all the messages
//on all calls the underlying grpc.ClientStream:RecvMsg message has this data added
func (d *dataAddingClientStream) RecvMsg(m interface{}) error {
	if d.constantMessage == nil {
		//prevent future initiations
		d.constantMessage = newEmpty(m)
		header, _ := d.ClientStream.Header()
		if head, ok := header[XgRPCConst]; ok && len(head) > 0 {
			if err := unmarshal(head[0], d.constantMessage); err != nil {
				return err // TODO probably not like this
			}
			d.fieldsToSet = make(map[reflect.Value][]int)
			if err := generateSetFields(d.constantMessage, d.fieldsToSet); err != nil {
				return err // TODO probably not like this
			}
			if len(d.fieldsToSet) == 0 {
				d.fieldsToSet = nil
			}
		}
	}
	if err := d.ClientStream.RecvMsg(m); err != nil {
		return err
	}
	if err := setFields(&m, d.fieldsToSet); err != nil {
		return err // TODO probably not like this
	}
	return nil
}

//setFields sets the fields, from a map[reflect.Value][]int (a value-key-ish map) to the message.
//For each value to set. It uses the path to that field (the []int) to set this value to the message
//Only empty receiver fields have its value overridden
func setFields(i *interface{}, fields map[reflect.Value][]int) error {
	if fields == nil {
		return nil
	}
	if receiverVal, ok := firstStruct(reflect.ValueOf(i)); ok {
		for k, v := range fields {
			firstField, subStructures := v[0], v[1:]
			fieldToSet := receiverVal.Field(firstField)
			for _, i := range subStructures {
				field, _ := firstStruct(fieldToSet)
				fieldToSet = field.Field(i)
			}
			if isEmptyValue(fieldToSet) {
				fieldToSet.Set(k)
			}
		}
	}
	return nil
}

//firstStruct returns the first non-reflect.Ptr, non-reflect.Interface reflect.Value
// given a reflect.Value that is still valid keep wrapping until you reach a non-Ptr/Interface
// return that value and a hint if it is a reflect.Struct.
// fx the message is a reflect.Ptr to an reflect.Interface, to a reflect.Ptr to a reflect.Struct,
// while the object given by newEmpty is a reflect.Ptr to a reflect.Struct
// the nested reflect.Struct given as reflect.Value also has to be unwrapped using this method.
func firstStruct(of reflect.Value) (reflect.Value, bool) {
	ok := false
	value := of
	for value.IsValid() && (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) {
		value = value.Elem()
	}
	if value.Kind() == reflect.Struct {
		ok = true
	}
	return value, ok
}

//newEmpty simply creates a new instance of an interface given an instance of that interface
func newEmpty(t interface{}) interface{} {
	return reflect.New(reflect.TypeOf(t).Elem()).Interface()
}

//generateSetFields an entrypoint given a target populate a
// map[reflect.Value][]int to the method abstractSetFields
func generateSetFields(target interface{}, setFields map[reflect.Value][]int) error {
	donorVal := reflect.ValueOf(target).Elem()
	return abstractSetFields(donorVal, setFields, []int(nil))
}

//abstractSetFields is a recursive method that adds all Writable fields
//that has a nonEmpty value to the given map[reflect.Value][]int
//it walks the tree of structure fields of the donor. (nested tree of struct)
func abstractSetFields(donorVal reflect.Value, aMap map[reflect.Value][]int, curPosition []int) error {
	if !donorVal.IsValid() {
		return nil
	}
	for i := 0; i < donorVal.NumField(); i++ {
		donorField := donorVal.Field(i)
		if !donorField.CanSet() {
			//skip early if the field cannot be Set
			//This check also allow skipping the check on the method #setFields
			continue
		}
		nextPosition := append(curPosition, i)
		if field, ok := firstStruct(donorField); ok {
			//nested structs
			_ = abstractSetFields(field, aMap, nextPosition)
		} else if shouldDonate(field) {
			aMap[field] = nextPosition
		}
	}
	return nil
}

//shouldDonate helps determine whether or not to donate a field to the message
func shouldDonate(field reflect.Value) bool {
	return !isEmptyValue(field)
}

// From src/pkg/encoding/json/encode.go. with changes via mergo/merge (last 7 lines)
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmptyValue(v.Elem())
	case reflect.Func:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}
