package runjs

import (
	"fmt"

	"github.com/dop251/goja"
)

// Custom PromiseError type
type AssertionError struct {
	message string
}

// Implements Goja's Error interface
func (e *AssertionError) Error() string {
	return fmt.Sprintf("AssertionError: %s", e.message)
}

// Creates a new PromiseError in Goja
func NewAssertError(vm *goja.Runtime, msg string) *goja.Object {
	errObj := vm.NewObject()
	errObj.Set("name", "AssertionError")
	errObj.Set("message", msg)
	errObj.Set("toString", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(fmt.Sprintf("AssertionError: %s", msg))
	})
	return errObj
}
