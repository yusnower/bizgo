package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Create a standard error
func a1() error {
	bizError := BizCode{Key: "a1"}
	return bizError.Wrap(stderrors.New("a1 error"), "a1")
}

// Wrap an existing BizError
func a2() error {
	bizError := BizCode{Key: "a2"}
	err := a1()
	return bizError.Wrap(err, "a2")
}

// Wrap a BizError again
func a3() error {
	bizError := BizCode{Key: "a3"}
	err := a2()
	return bizError.Wrap(err, "a3")
}

// Return nil error
func nilError() error {
	return nil
}

// Return a non-BizError type error
func standardError() error {
	return stderrors.New("standard error")
}

// Test error formatting
func TestErrorFormat(t *testing.T) {
	err := a3()

	// Test formatting of the error chain
	t.Logf("Standard format: %s", err)
	t.Logf("Verbose format: %+v", err)
	t.Logf("Quote format: %q", err)
}

// Test various scenarios for the Equal method
func TestEqual(t *testing.T) {
	t.Run("Deep nested error comparison", func(t *testing.T) {
		err := a3()

		a1Code := BizCode{Key: "a1"}
		a2Code := BizCode{Key: "a2"}
		a3Code := BizCode{Key: "a3"}
		nonExistCode := BizCode{Key: "nonexist"}

		// Verify that each error in the chain can be found
		assert.True(t, a1Code.Equal(err), "Should find a1 error")
		assert.True(t, a2Code.Equal(err), "Should find a2 error")
		assert.True(t, a3Code.Equal(err), "Should find a3 error")
		assert.False(t, nonExistCode.Equal(err), "Should not find non-existent error")
	})

	t.Run("Comparison with nil", func(t *testing.T) {
		code := BizCode{Key: "test"}
		err := nilError()
		assert.False(t, code.Equal(err), "Comparison with nil should return false")
		assert.False(t, code.Equal(nil), "Comparison with nil should return false")
	})

	t.Run("Comparison with standard error", func(t *testing.T) {
		code := BizCode{Key: "test"}
		err := standardError()
		assert.False(t, code.Equal(err), "Comparison with non-BizError should return false")
	})

	t.Run("Direct comparison with same Key", func(t *testing.T) {
		code := BizCode{Key: "same"}
		err := code.Wrap(stderrors.New("test"), "value")
		assert.True(t, code.Equal(err), "Same key should return true")
	})
}

// Test Wrap method
func TestWrap(t *testing.T) {
	t.Run("Wrapping nil error", func(t *testing.T) {
		code := BizCode{Key: "test"}
		err := code.Wrap(nil)
		assert.Nil(t, err, "Wrapping nil should return nil")
	})

	t.Run("Wrapping standard error", func(t *testing.T) {
		code := BizCode{Key: "test"}
		stdErr := stderrors.New("standard error")
		err := code.Wrap(stdErr)

		var bizErr *BizError
		assert.True(t, stderrors.As(err, &bizErr), "Should be convertible to BizError")
		assert.Equal(t, code.Key, bizErr.key, "Keys should be the same")
		assert.Equal(t, stdErr, bizErr.Unwrap(), "Inner error should be the original error")
	})

	t.Run("Wrapping BizError", func(t *testing.T) {
		code1 := BizCode{Key: "inner"}
		code2 := BizCode{Key: "outer"}

		inner := code1.Wrap(stderrors.New("inner error"))
		outer := code2.Wrap(inner)

		assert.True(t, code2.Equal(outer), "Outer key should match")
		assert.True(t, code1.Equal(outer), "Inner key should match")
	})
}

// Test error judgment using errors.Is and errors.As
func TestStandardErrorFunctions(t *testing.T) {
	// Test BizError's Is method using errors.Is
	t.Run("Using errors.Is", func(t *testing.T) {
		code := BizCode{Key: "test"}
		err := code.Wrap(stderrors.New("test error"))

		target := &BizError{key: "test"}
		assert.True(t, stderrors.Is(err, target), "errors.Is should recognize BizError with the same key")

		// Test error chain
		nestedCode1 := BizCode{Key: "inner"}
		nestedCode2 := BizCode{Key: "middle"}
		nestedCode3 := BizCode{Key: "outer"}

		inner := nestedCode1.Wrap(stderrors.New("inner error"))
		middle := nestedCode2.Wrap(inner)
		outer := nestedCode3.Wrap(middle)

		// All levels should be recognizable
		assert.True(t, stderrors.Is(outer, &BizError{key: "outer"}), "should find outer error")
		assert.True(t, stderrors.Is(outer, &BizError{key: "middle"}), "should find middle error")
		assert.True(t, stderrors.Is(outer, &BizError{key: "inner"}), "should find inner error")
		assert.False(t, stderrors.Is(outer, &BizError{key: "nonexist"}), "should not find nonexistent error")
	})

	t.Run("Using errors.As", func(t *testing.T) {
		code := BizCode{Key: "test"}
		err := code.Wrap(stderrors.New("test error"))

		var bizErr *BizError
		assert.True(t, stderrors.As(err, &bizErr), "errors.As should be able to convert the error to BizError")
		assert.Equal(t, code.Key, bizErr.key, "The converted error should retain the original key")
	})
}
