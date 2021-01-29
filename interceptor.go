package grpcConst

import (
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"reflect"
	"strconv"
	"strings"
)

const XgRPCConst = "x-grpc-const"

func HeaderSetConstant(v interface{}) (metadata.MD, error) {
	msg, err := encoding.GetCodec("proto").Marshal(v)
	return metadata.Pairs(XgRPCConst, base64.URLEncoding.EncodeToString(msg)), err
}

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
	setFields map[string]reflect.Value
}

func (d *dataAddingClientStream) RecvMsg(m interface{}) error {
	if d.constantMessage == nil {
		header, _ := d.ClientStream.Header()
		head := header[XgRPCConst]
		protoMsg, err := base64.URLEncoding.DecodeString(head[0])
		if err != nil {
			return err
		}
		d.constantMessage = newEmpty(m)
		if err := encoding.GetCodec("proto").Unmarshal(protoMsg, d.constantMessage); err != nil {
			return err
		}
		d.setFields = make(map[string]reflect.Value)
		if err := generateSetFields(d.constantMessage, d.setFields); err != nil {
			return err
		}
	}
	if err := d.ClientStream.RecvMsg(m); err != nil {
		return err
	}
	if err := setFields(&m, d.setFields); err != nil {
		return err
	}
	return nil
}

func setFields(i *interface{}, fields map[string]reflect.Value) error {
	if receiverVal, ok := firstStruct(reflect.ValueOf(i)); ok {
		for k, v := range fields {
			fieldPath := strings.Split(k, ".")
			var fieldToSet reflect.Value
			for _, a := range fieldPath {
				i, _ := strconv.Atoi(a)
				if !fieldToSet.IsValid() {
					fieldToSet = receiverVal.Field(i)
					continue
				}
				field, _ := firstStruct(fieldToSet)
				fieldToSet = field.Field(i)
			}
			if isEmptyValue(fieldToSet) {
				fieldToSet.Set(v)
			}
		}
	}
	return nil
}

func firstStruct(of reflect.Value) (reflect.Value, bool) {
	ok := false
	value := of
	for value.IsValid() && ( value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface ) {
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


func generateSetFields(target interface{}, setFields map[string]reflect.Value) error {
	donorVal := reflect.ValueOf(target).Elem()
	return abstractSetFields(donorVal, setFields, "")
}

func abstractSetFields(donorVal reflect.Value, aMap map[string]reflect.Value, baseString string) error {
	if !donorVal.IsValid() {
		return nil
	}
	for i := 0; i < donorVal.NumField(); i++ {
		donorField := donorVal.Field(i)
		if !donorField.CanSet() {
			continue
		}
		if field, ok := firstStruct(donorField); ok {
			_ = abstractSetFields(field, aMap, fmt.Sprintf("%s%d.", baseString, i))
		} else if shouldDonate(field) {
				aMap[fmt.Sprintf("%s%d", baseString, i)] = field
		}
	}
	return nil
}

func shouldDonate(field reflect.Value) bool {
	return ! isEmptyValue(field)
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
