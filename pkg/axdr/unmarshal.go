package axdr

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"time"
)

func UnmarshalData(data DlmsData, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer")
	}

	return unify(&data, reflect.Indirect(rv))
}

func unify(data *DlmsData, rv reflect.Value) error {
	expectedKind := rv.Kind()
	gotKind := reflect.ValueOf(data.Value).Kind()

	_, isTime := rv.Interface().(time.Time)
	_, isDlmsData := rv.Interface().(DlmsData)

	switch {
	case expectedKind == reflect.Ptr:
		elem := reflect.New(rv.Type().Elem())
		err := unify(data, reflect.Indirect(elem))
		if err != nil {
			return err
		}
		rv.Set(elem)
	case isDlmsData:
		rv.Set(reflect.ValueOf(*data))
	case expectedKind == reflect.Slice && gotKind == reflect.Slice:
		return unifySlice(data, rv)
	case expectedKind == reflect.Struct && gotKind == reflect.Slice:
		return unifyStruct(data, rv)
	case expectedKind == reflect.Int && (gotKind >= reflect.Int && gotKind <= reflect.Int64):
		return unifyInt(data, rv)
	case expectedKind == reflect.Uint && (gotKind >= reflect.Uint && gotKind <= reflect.Uint64):
		return unifyUint(data, rv)
	case isTime && gotKind == reflect.String:
		return unifyDateTime(data, rv)
	case expectedKind == gotKind:
		rv.Set(reflect.ValueOf(data.Value))
	default:
		return fmt.Errorf("expected %s, got %s", expectedKind, gotKind)
	}

	return nil
}

func unifyDateTime(data *DlmsData, rv reflect.Value) error {
	v, err := hex.DecodeString(data.Value.(string))
	if err != nil {
		return fmt.Errorf("invalid date time: %w", err)
	}

	_, t, err := DecodeDateTime(&v)
	if err != nil {
		return fmt.Errorf("invalid date time: %w", err)
	}
	rv.Set(reflect.ValueOf(t))

	return nil
}

func unifySlice(data *DlmsData, rv reflect.Value) error {
	slice := data.Value.([]*DlmsData)

	n := len(slice)
	if rv.IsNil() || rv.Cap() < n {
		rv.Set(reflect.MakeSlice(rv.Type(), n, n))
	}
	rv.SetLen(n)

	for i := 0; i < n; i++ {
		sliceval := reflect.Indirect(rv.Index(i))
		if err := unify(slice[i], sliceval); err != nil {
			return fmt.Errorf("slice error in field %d: %w", i, err)
		}
	}

	return nil
}

func unifyStruct(data *DlmsData, rv reflect.Value) error {
	slice := data.Value.([]*DlmsData)
	n := len(slice)

	if rv.NumField() != n {
		return fmt.Errorf("struct has %d fields, but data has %d fields", rv.NumField(), n)
	}

	for i := 0; i < n; i++ {
		field := rv.Field(i)

		if structField := rv.Type().Field(i); rv.Type().Field(i).PkgPath != "" {
			return fmt.Errorf("cannot unmarshal into unexported field '%s'", structField.Name)
		}

		if field.Kind() == reflect.Ptr {
			if slice[i].Tag != TagNull && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}

			if slice[i].Tag == TagNull && !field.IsNil() {
				field.Set(reflect.Zero(field.Type()))
			}
		}

		if slice[i].Tag != TagNull {
			if err := unify(slice[i], reflect.Indirect(field)); err != nil {
				return fmt.Errorf("struct error in field %s: %w", rv.Type().Field(i).Name, err)
			}
		}
	}

	return nil
}

func unifyInt(data *DlmsData, rv reflect.Value) error {
	switch v := data.Value.(type) {
	case int8:
		rv.SetInt(int64(v))
	case int16:
		rv.SetInt(int64(v))
	case int32:
		rv.SetInt(int64(v))
	case int64:
		rv.SetInt(v)
	default:
		return fmt.Errorf("unexpected type %T", data.Value)
	}

	return nil
}

func unifyUint(data *DlmsData, rv reflect.Value) error {
	switch v := data.Value.(type) {
	case uint8:
		rv.SetUint(uint64(v))
	case uint16:
		rv.SetUint(uint64(v))
	case uint32:
		rv.SetUint(uint64(v))
	case uint64:
		rv.SetUint(v)
	default:
		return fmt.Errorf("unexpected type %T", data.Value)
	}

	return nil
}
