//merge is a package with support functionality for grpcConst
//this helps add default values from one object to many objects
//the code skips a lot of reflection because much can be assumed at initiation
//There may be some issues with this package, it is meant for mapping
//structs compiled via protoc/gRPC.
//
//A Merger will not "merge" fields that are private.
//
//usage:
//		var objectWithDefaultValues = ...
//		aMerger := merger.NewMerger(&objectWithDefaultValues)
//		for _, item := range myItems {
//			err := aMerger.SetFields(&item)
//			... item will receive non-zero/non-nil fields from objectWithDefaultValues
//		}
//or:
//		aReducer := merge.NewReducer(&objectWithDefaultValues)
//		for _, item := range myItems {
//			err := aReducer.aReducer.RemoveFields(&item)
//			... item will have fields removed by comparing with the default values from objectWithDefaultValues
//		}
//Limitation:
//Merging an interface{} has limitations!
//You might get unwanted behavior when reducing any reflect.[Map, Interface, Slice, Array, Func, Invalid]
//proto.Merge merges unknownFields, this does not!
//proto.Merge merges slices, this does not!
package merge

import (
	"reflect"
)

//Merger can SetFields to a receiver given values given
//by a donor at instance initiation via NewMerger
type Merger interface {
	SetFields(interface{}) error
}

//Reducer can RemoveFields of a given object given values given
//by a reference at instance initiation via NewReducer
//the reducer removes fields that are equal the reference
type Reducer interface {
	RemoveFields(interface{}) error
}

//NewMerger initiates the Merger, populating the []reflectTree
//for future merging of pointer targets
//panics if the donor is not a pointer
func NewMerger(donor interface{}) Merger {
	merger := reflectTree{}
	fieldsToSet, err := abstractSetFields(reflect.ValueOf(donor).Elem())
	if err != nil {
		//handle? log?
	}
	merger.Branches = fieldsToSet
	if len(fieldsToSet) == 0 {
		merger.Branches = nil
	}
	return merger
}

//NewReducer initiates a merger and returns its Reducer
func NewReducer(reference interface{}) Reducer {
	return NewMerger(reference).(Reducer)
}

//reflectTree is a data-structure to save the fields that should be defaulted
type reflectTree struct {
	Key      int
	Value    ValueWrapper
	Branches []reflectTree
}

type getterFunction func(reflect.Value) interface{}
type emptyCheckerFunction func(reflect.Value) bool

type ValueWrapper struct {
	Value      reflect.Value
	GetValue   getterFunction
	HasNoValue emptyCheckerFunction
}

//SetFields sets the fields, from a []reflectTree to the message.
//For each value to set. It uses the path to that field (the []int) to set this value to the message
//Only empty receiver fields have its value overridden
//panics on non-pointer values 'r'
//   checking and returning an error costs 3 ns pr. msg mapped, and it doesn't provide you with
func (m reflectTree) SetFields(r interface{}) error {
	return scanAll(m, r, setAField, false)
}

//RemoveFields removes fields from the subject that are equal to the reference
func (m reflectTree) RemoveFields(subject interface{}) error {
	return scanAll(m, subject, removeAField, true)
}

func scanAll(tree reflectTree, subject interface{}, method func(target reflect.Value, source ValueWrapper), retPtr bool) error {
	if tree.Branches == nil {
		return nil
	}
	receiverVal := reflect.ValueOf(subject).Elem()
	for _, leaf := range tree.Branches {
		if err := doWithAField(leaf, receiverVal, method, retPtr); err != nil {
			return err
		}
	}
	return nil
}

func removeAField(target reflect.Value, source ValueWrapper) {
	if source.GetValue(target) == source.GetValue(source.Value) {
		target.Set(reflect.New(source.Value.Type()).Elem())
	}
}

func setAField(target reflect.Value, source ValueWrapper) {
	if source.HasNoValue(target) {
		target.Set(source.Value)
	}
}

func doWithAField(leaf reflectTree, field reflect.Value,
	hitFunc func(target reflect.Value, source ValueWrapper),
	returnOnPtrNil bool) error {
	theField := field.Field(leaf.Key)
	if leaf.Branches == nil {
		hitFunc(theField, leaf.Value)
		return nil
	}
	if theField.Kind() == reflect.Ptr {
		if theField.IsNil() {
			if returnOnPtrNil {
				return nil
			}
			theField.Set(reflect.New(leaf.Value.Value.Type()))
		}
		theField = theField.Elem()
	}
	if len(leaf.Branches) == 0 {
		if theField.IsZero() {
			theField.Set(reflect.New(leaf.Value.Value.Type()).Elem())
		}
		return nil
	} else {
		for _, branch := range leaf.Branches {
			if err := doWithAField(branch, theField, hitFunc, returnOnPtrNil); err != nil {
				return err
			}
		}
		return nil
	}
}

//abstractSetFields is a recursive method that adds all Writable fields
//that has a nonEmpty value to the given reflectTree
//it walks the tree of structure fields of the donor. (nested tree of struct)
func abstractSetFields(donorVal reflect.Value) ([]reflectTree, error) {
	if !donorVal.IsValid() {
		return nil, nil
	}
	var tree []reflectTree
	for i := 0; i < donorVal.NumField(); i++ {
		donorField := donorVal.Field(i)
		var field = donorField
		for field.IsValid() && (field.Kind() == reflect.Ptr) {
			field = field.Elem()
		}
		if !field.CanSet() {
			//skip early if the field cannot be Set
			//This check also allow skipping the check on the method #setFields
			continue
		}
		get, check := getValueMethods(field)
		leaf := reflectTree{
			Key:      i,
			Value:    ValueWrapper{field, get, check},
			Branches: nil,
		}
		if field.Kind() == reflect.Struct {
			//nested structs
			leaf.Branches, _ = abstractSetFields(field)
			if leaf.Branches == nil {
				leaf.Branches = []reflectTree{}
			}
		}
		if !leaf.Value.HasNoValue(field) {
			tree = append(tree, leaf)
		}
	}
	return tree, nil
}

func getValueMethods(v reflect.Value) (getterFunction, emptyCheckerFunction) {
	switch v.Kind() {
	case reflect.Bool:
		return func(value reflect.Value) interface{} { return value.Bool() },
			func(value reflect.Value) bool { return !value.Bool() }
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(value reflect.Value) interface{} { return value.Int() },
			func(value reflect.Value) bool { return value.Int() == 0 }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(value reflect.Value) interface{} { return value.Uint() },
			func(value reflect.Value) bool { return value.Uint() == 0 }
	case reflect.Float32, reflect.Float64:
		return func(value reflect.Value) interface{} { return value.Float() },
			func(value reflect.Value) bool { return value.Float() == 0 }
	case reflect.String:
		return func(value reflect.Value) interface{} { return value.String() },
			func(value reflect.Value) bool { return value.Len() == 0 }
	case reflect.Array, reflect.Map, reflect.Slice:
		return func(value reflect.Value) interface{} { return nil },
			func(value reflect.Value) bool { return value.Len() == 0 }
	case reflect.Interface, reflect.Ptr:
		return func(value reflect.Value) interface{} { return nil },
			func(value reflect.Value) bool { return value.IsNil() }
	case reflect.Func:
		return func(value reflect.Value) interface{} { return nil },
			func(value reflect.Value) bool { return v.IsNil() }
	case reflect.Invalid:
		return func(value reflect.Value) interface{} { return nil },
			func(value reflect.Value) bool { return true }
	}
	return func(value reflect.Value) interface{} { return nil },
		func(value reflect.Value) bool { return false }
}
