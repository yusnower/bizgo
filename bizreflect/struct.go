package bizreflect

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	tagValue  = "biz"
	prefixTag = "prefix" // New constant for the prefix tag
)

func InitStructG[T any]() *T {
	obj := new(T)
	err := InitStruct(obj)
	if err != nil {
		panic(err)
	}

	return obj
}

// InitStruct default struct with tag
func InitStruct(obj interface{}) error {
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer to struct, got %v", typ.Kind())
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct, got pointer to %v", typ.Kind())
	}

	val := reflect.ValueOf(obj).Elem()

	return initValue(val, "")
}

func initValue(val reflect.Value, parentPrefix string) error {
	typ := val.Type()
	// Check if struct has a prefix tag
	prefix := parentPrefix

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get the prefix from the current struct if it exists
		if fieldType.Name == prefixTag {
			if field.Kind() == reflect.String && field.String() != "" {
				prefix = field.String()
			}
		}

		if err := setFieldValue(field, fieldType, prefix); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, fieldType reflect.StructField, prefix string) error {
	// Check if this field itself has a prefix tag that overrides the struct prefix
	fieldPrefix := prefix
	if prefixValue := fieldType.Tag.Get(prefixTag); prefixValue != "" {
		fieldPrefix = prefixValue
	}

	switch fieldType.Type.Kind() {
	case reflect.Struct:
		return handleStructField(field, fieldType, fieldPrefix)
	case reflect.Map:
		return handleMapField(field)
	case reflect.String:
		value := fieldType.Tag.Get(tagValue)
		// Apply prefix to string fields if prefix exists and value is not empty
		if fieldPrefix != "" && value != "" {
			value = fieldPrefix + value
		}
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := fieldType.Tag.Get(tagValue)
		return handleIntField(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := fieldType.Tag.Get(tagValue)
		return handleUintField(field, value)
	case reflect.Float32, reflect.Float64:
		value := fieldType.Tag.Get(tagValue)
		return handleFloatField(field, value)
	case reflect.Bool:
		value := fieldType.Tag.Get(tagValue)
		return handleBoolField(field, value)
	default:

	}

	return nil
}

func handleStructField(field reflect.Value, fieldType reflect.StructField, prefix string) error {
	fieldValue := reflect.New(fieldType.Type).Elem()

	if err := initValue(fieldValue, prefix); err != nil {
		return err
	}

	for i := 0; i < fieldValue.NumField(); i++ {
		innerField := fieldValue.Field(i)
		innerFieldType := fieldType.Type.Field(i)

		if !innerField.CanSet() {
			continue
		}

		innerBizTag := innerFieldType.Tag.Get(tagValue)
		if innerBizTag != "" {
			continue
		}

		parentTagValue := fieldType.Tag.Get(innerFieldType.Name)
		if parentTagValue != "" {
			// When setting parent-defined values, also apply prefix if it's a string field
			if innerField.Kind() == reflect.String && prefix != "" {
				parentTagValue = prefix + parentTagValue
			}
			if err := setFieldByType(innerField, parentTagValue); err != nil {
				return fmt.Errorf("failed to set nested field %s from parent tag: %w", innerFieldType.Name, err)
			}
		}
	}

	field.Set(fieldValue)

	return nil
}

func setFieldByType(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return handleIntField(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return handleUintField(field, value)
	case reflect.Float32, reflect.Float64:
		return handleFloatField(field, value)
	case reflect.Bool:
		return handleBoolField(field, value)
	case reflect.Struct:
		return fmt.Errorf("cannot set struct field directly from tag value")
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

func handleMapField(field reflect.Value) error {
	field.Set(reflect.MakeMap(field.Type()))

	return nil
}

func handleIntField(field reflect.Value, fieldDefault string) error {
	if fieldDefault == "" {
		return nil
	}
	parseInt, err := strconv.ParseInt(fieldDefault, 10, 64)
	if err != nil {
		return fmt.Errorf("parse int error: %w", err)
	}
	field.SetInt(parseInt)

	return nil
}

func handleUintField(field reflect.Value, fieldDefault string) error {
	if fieldDefault == "" {
		return nil
	}
	parseUint, err := strconv.ParseUint(fieldDefault, 10, 64)
	if err != nil {
		return fmt.Errorf("parse uint error: %w", err)
	}
	field.SetUint(parseUint)

	return nil
}

func handleFloatField(field reflect.Value, fieldDefault string) error {
	if fieldDefault == "" {
		return nil
	}
	parseFloat, err := strconv.ParseFloat(fieldDefault, 64)
	if err != nil {

		return fmt.Errorf("parse float error: %w", err)
	}
	field.SetFloat(parseFloat)

	return nil
}

func handleBoolField(field reflect.Value, fieldDefault string) error {
	if fieldDefault == "" {
		return nil
	}
	parseBool, err := strconv.ParseBool(fieldDefault)
	if err != nil {
		return fmt.Errorf("parse bool error: %w", err)
	}
	field.SetBool(parseBool)

	return nil
}
