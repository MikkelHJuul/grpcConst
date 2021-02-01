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
	msg, err := encoding.GetCodec("proto").Marshal(v)
	return metadata.Pairs(XgRPCConst, base64.URLEncoding.EncodeToString(msg)), err
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

type dataAddingClientStream struct {
	grpc.ClientStream
	constantMessage interface{}
	fieldsToSet     map[reflect.Value][]int
}

func (d *dataAddingClientStream) RecvMsg(m interface{}) error {
	if d.constantMessage == nil {
		//prevent future initiations
		d.constantMessage = newEmpty(m)
		header, _ := d.ClientStream.Header()
		if head, ok := header[XgRPCConst]; ok && len(head) > 0 {
			protoMsg, err := base64.URLEncoding.DecodeString(head[0])
			if err != nil {
				return err // TODO probably not like this
			}
			if err := encoding.GetCodec("proto").Unmarshal(protoMsg, d.constantMessage); err != nil {
				return err // TODO probably not like this
			}
			d.fieldsToSet = make(map[reflect.Value][]int)
			if err := generateSetFields(d.constantMessage, d.fieldsToSet); err != nil {
				return err // TODO probably not like this
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

func setFields(i *interface{}, fields map[reflect.Value][]int) error {
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

func newEmpty(t interface{}) interface{} {
	return reflect.New(reflect.TypeOf(t).Elem()).Interface()
}

func generateSetFields(target interface{}, setFields map[reflect.Value][]int) error {
	donorVal := reflect.ValueOf(target).Elem()
	return abstractSetFields(donorVal, setFields, []int(nil))
}

func abstractSetFields(donorVal reflect.Value, aMap map[reflect.Value][]int, curPosition []int) error {
	if !donorVal.IsValid() {
		return nil
	}
	for i := 0; i < donorVal.NumField(); i++ {
		donorField := donorVal.Field(i)
		if !donorField.CanSet() {
			continue
		}
		nextPosition := append(curPosition, i)
		if field, ok := firstStruct(donorField); ok {
			_ = abstractSetFields(field, aMap, nextPosition)
		} else if shouldDonate(field) {
			aMap[field] = nextPosition
		}
	}
	return nil
}

func shouldDonate(field reflect.Value) bool {
	return !isEmptyValue(field)
}

// From src/pkg/encoding/json/encode.go.
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
