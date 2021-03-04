package grpcConst

import "fmt"

//Merger is the interface of a type that can merge into itself
type Merger interface {
	Merge(interface{})
}

type MessageMergerReducer struct {
	ConstantMessage interface{}
}

func (m MessageMergerReducer) SetFields(msg interface{}) error {
	if merger, ok := msg.(Merger); ok {
		merger.Merge(m.ConstantMessage)
		return nil
	}
	return fmt.Errorf("message %v is not a Merger", msg)
}

//Reducer is the interface of a type that can reduce itself from a reference
type Reducer interface {
	Reduce(interface{})
}

func (m MessageMergerReducer) RemoveFields(msg interface{}) error {
	if reducer, ok := msg.(Reducer); ok {
		reducer.Reduce(m.ConstantMessage)
		return nil
	}
	return fmt.Errorf("message %v is not a Reducer", msg)
}
