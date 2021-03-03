package merge

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

type protoMerger struct {
	donor proto.Message
}

//SetFields implements the interface Merger
//proxy for proto.Merge using a base donor
func (pm protoMerger) SetFields(receiver interface{}) error {
	if m, ok := receiver.(proto.Message); ok {
		proto.Merge(m, pm.donor)
		return nil
	}
	return fmt.Errorf("receiver is not a proto.Message")
}

//NewProtoMerger returns an instance of a Merger using proto.Merge to merge Messages
func NewProtoMerger(donor interface{}) Merger {
	if m, ok := donor.(proto.Message); ok {
		return protoMerger{donor: m}
	}
	return nil
}
