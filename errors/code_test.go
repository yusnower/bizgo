package errors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test error code initialization for nested structures
type TestErrorCodes struct {
	Common BizCode `Key:"common"`
	Auth   BizCode `Key:"auth"`

	User struct {
		NotFound BizCode `Key:"notFound"`
		Invalid  BizCode `Key:"invalid"`
	} `prefix:"user"`

	Order struct {
		NotFound BizCode `Key:"notFound"`
		Invalid  BizCode `Key:"invalid"`

		Payment struct {
			Failed   BizCode `Key:"failed"`
			Canceled BizCode `Key:"canceled"`
		} `prefix:"payment"`
	} `prefix:"order"`
}

// Test the initialization logic of the InitModule function
func TestInitModule(t *testing.T) {
	t.Run("Initialize nested error codes", func(t *testing.T) {
		errorCodes := &TestErrorCodes{}
		InitModule("app", errorCodes)

		// Test top-level error code initialization
		assert.Equal(t, "appcommon", errorCodes.Common.Key)
		assert.Equal(t, "appauth", errorCodes.Auth.Key)

		// Test nested User error code initialization
		assert.Equal(t, "appusernotFound", errorCodes.User.NotFound.Key)
		assert.Equal(t, "appuserinvalid", errorCodes.User.Invalid.Key)

		// Test multi-level nested Order error code initialization
		assert.Equal(t, "appordernotFound", errorCodes.Order.NotFound.Key)
		assert.Equal(t, "apporderinvalid", errorCodes.Order.Invalid.Key)
		assert.Equal(t, "apporderpaymentfailed", errorCodes.Order.Payment.Failed.Key)
		assert.Equal(t, "apporderpaymentcanceled", errorCodes.Order.Payment.Canceled.Key)
	})

	t.Run("Pass nil pointer", func(t *testing.T) {
		// Should not panic
		InitModule("prefix_", nil)
	})

	t.Run("Pass non-pointer", func(t *testing.T) {
		// Should not panic
		nonPtr := TestErrorCodes{}
		InitModule("prefix_", nonPtr)
	})

	t.Run("Pass nil struct pointer", func(t *testing.T) {
		// Should not panic
		var nilPtr *TestErrorCodes
		InitModule("prefix_", nilPtr)
	})

	t.Run("Pass non-struct pointer", func(t *testing.T) {
		// Should not panic
		str := "not a struct"
		InitModule("prefix_", &str)
	})
}

// Test BizCode methods
func TestBizCode(t *testing.T) {
	t.Run("Ctx method", func(t *testing.T) {
		code := BizCode{Key: "test"}
		ctx := context.Background()

		// Test if Ctx method correctly preserves Key and sets ctx
		codeWithCtx := code.Ctx(ctx)
		assert.Equal(t, code.Key, codeWithCtx.Key)
		assert.Equal(t, ctx, codeWithCtx.ctx)
	})

	t.Run("String method", func(t *testing.T) {
		code := BizCode{Key: "test_key"}
		assert.Equal(t, "test_key", code.String())
	})
}

// Test error code combination usage
func TestErrorCodeIntegration(t *testing.T) {
	// Initialize error codes
	errorCodes := &TestErrorCodes{}
	InitModule("app", errorCodes)

	t.Run("Create and wrap errors", func(t *testing.T) {
		// Create a user not found error
		originalErr := assert.AnError
		userNotFoundErr := errorCodes.User.NotFound.Wrap(originalErr, "user id", 12345)

		// Verify that BizError was created correctly
		assert.True(t, errorCodes.User.NotFound.Equal(userNotFoundErr))
		assert.False(t, errorCodes.User.Invalid.Equal(userNotFoundErr))
	})

	t.Run("Nested error wrapping", func(t *testing.T) {
		// Create payment failed error
		originalErr := assert.AnError
		paymentErr := errorCodes.Order.Payment.Failed.Wrap(originalErr, "payment id", "P12345")

		// Wrap as order error
		orderErr := errorCodes.Order.Invalid.Wrap(paymentErr, "order id", "O98765")

		// Verify error chain
		assert.True(t, errorCodes.Order.Invalid.Equal(orderErr))
		assert.True(t, errorCodes.Order.Payment.Failed.Equal(orderErr))
		assert.False(t, errorCodes.User.NotFound.Equal(orderErr))
	})
}
