package errors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
)

var logger defaultLogger

type ErrorInfo struct {
	Ctx      context.Context `json:"-"`
	Err      error           `json:"err"`
	Location string          `json:"location"`
	Value    interface{}     `json:"value"`
	Uuid     string          `json:"uuid"`
}

type Logger interface {
	PrintBizError(ErrorInfo)
}

type defaultLogger struct{}

func (l *defaultLogger) PrintBizError(obj *ErrorInfo) {
	log.Println(obj.Uuid, obj.Value, obj.Err)
}

const (
	bizCodeKey    = "Key"
	bizCodePrefix = "prefix"
)

// BizCode represents a business error code that can be used to create
// and identify specific types of errors in an application.
type BizCode struct {
	ctx context.Context

	Key string // Unique identifier for the error code
}

func (r BizCode) Ctx(ctx context.Context) BizCode {
	return BizCode{
		Key: r.Key,
		ctx: ctx,
	}
}

// Wrap creates a new BizError with the given error message and captures
// the current code location (file, line, function).
// This allows attaching a specific error code to any error.
func (r BizCode) Wrap(err error, obj ...interface{}) error {
	if err == nil {
		return nil
	}

	newBizErr := &BizError{
		key: r.Key,
		err: err,
	}

	var bizErr *BizError
	if errors.As(err, &bizErr) {
		newBizErr.stack = bizErr.stack
		newBizErr.uuid = bizErr.uuid
	} else {
		newBizErr.stack = callers()
		newBizErr.uuid = "123"
	}

	logger.PrintBizError(&ErrorInfo{
		Ctx:      r.ctx,
		Err:      err,
		Location: captureLocation(1),
		Value:    obj,
		Uuid:     newBizErr.uuid,
	})

	return newBizErr
}

// Equal checks if the provided error has the same error code as this BizCode.
// This enables type-safe error comparison.
func (r BizCode) Equal(err error) bool {
	if err == nil {
		return false
	}

	var bizErr *BizError
	if errors.As(err, &bizErr) {
		if r.Key == bizErr.key {
			return true
		}

		if bizErr.err != nil {
			return r.Equal(bizErr.err)
		}

	}

	return false
}

// String returns the string representation of the BizCode.
func (r BizCode) String() string {
	return r.Key
}

// InitModule recursively initializes all BizCode fields in the provided struct.
// The prefix is prepended to all error codes to create namespaced error codes.
//
// Parameters:
//   - prefix: A string prefix to prepend to all error codes
//   - obj: A pointer to a struct containing BizCode fields
func InitModule(prefix string, obj interface{}) {
	if obj == nil {
		return
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return
	}

	v = v.Elem()
	bizCodeType := reflect.TypeOf(BizCode{})

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)

		// Initialize BizCode fields
		if field.Type == bizCodeType && fieldValue.CanSet() {
			key := field.Tag.Get(bizCodeKey)
			if key != "" {
				fieldValue.FieldByName(bizCodeKey).SetString(fmt.Sprintf("%s%s", prefix, key))
			}
		}

		// Recursively process nested structs
		if field.Type.Kind() == reflect.Struct {
			nestedPrefix := prefix
			if prefixTag := field.Tag.Get(bizCodePrefix); prefixTag != "" {
				nestedPrefix = fmt.Sprintf("%s%s", prefix, prefixTag)
			}

			if fieldValue.CanAddr() {
				InitModule(nestedPrefix, fieldValue.Addr().Interface())
			}
		}
	}
}
