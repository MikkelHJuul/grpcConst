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
package merge

import (
	"fmt"
	"reflect"
)

type merger struct {
	mergeValues []reflectTree
	Initiated   bool
}

type Merger interface {
	SetFields(interface{}) error
}

//NewMerger initiates the Merger, populating the []reflectTree
//for future merging of pointer targets
//panics if the donor is not a pointer
func NewMerger(donor interface{}) Merger {
	mrgr := merger{Initiated: true}
	fieldsToSet, err := abstractSetFields(reflect.ValueOf(donor).Elem())
	if err != nil {
		//handle? log?
	}
	mrgr.mergeValues = fieldsToSet
	if len(fieldsToSet) == 0 {
		mrgr.mergeValues = nil
	}
	return mrgr
}

//reflectTree is a data-structure to save the fields that should be defaulted
type reflectTree struct {
	Key      int
	Value    reflect.Value
	Branches []reflectTree
}

//SetFields sets the fields, from a []reflectTree to the message.
//For each value to set. It uses the path to that field (the []int) to set this value to the message
//Only empty receiver fields have its value overridden
//panics on non-pointer values 'r'
//   checking and returning an error costs 3 ns pr. msg mapped, and it doesn't provide you with
func (m merger) SetFields(r interface{}) error {
	if m.mergeValues == nil {
		return nil
	}
	receiverVal := reflect.ValueOf(r).Elem()
	for _, leaf := range m.mergeValues {
		if err := setAField(leaf, receiverVal); err != nil {
			return err
		}
	}
	return nil
}

func setAField(leaf reflectTree, field reflect.Value) error {
	theField := field.Field(leaf.Key)
	if len(leaf.Branches) == 0 {
		if isEmptyValue(theField) {
			theField.Set(leaf.Value)
		}
		return nil
	}
	if theField.Kind() == reflect.Ptr {
		if theField.IsNil() {
			theField.Set(leaf.Value)
			return nil
		}
		theField = theField.Elem()
	}
	for _, branch := range leaf.Branches {
		err := setAField(branch, theField)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("you shouldn't be able to hit this")
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
		for field.IsValid() && (field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface) {
			field = field.Elem()
		}
		if !field.CanSet() {
			//skip early if the field cannot be Set
			//This check also allow skipping the check on the method #setFields
			continue
		}
		leaf := reflectTree{
			Key:      i,
			Value:    donorField,
			Branches: []reflectTree{},
		}
		if field.Kind() == reflect.Struct {
			//nested structs
			branches, _ := abstractSetFields(field)
			leaf.Branches = branches
		}
		if shouldDonate(field) {
			tree = append(tree, leaf)
		}
	}
	return tree, nil
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
