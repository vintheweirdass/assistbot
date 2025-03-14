package run

import (
	"fmt"

	"github.com/dop251/goja"
)

// Custom PromiseError type
type PromiseError struct {
	message string
}

// Implements Goja's Error interface
func (e *PromiseError) Error() string {
	return fmt.Sprintf("(Promise) %s", e.message)
}

// Creates a new PromiseError in Goja
func NewPromiseError(vm *goja.Runtime, msg string) *goja.Object {
	errObj := vm.NewObject()
	errObj.Set("name", "PromiseError")
	errObj.Set("message", msg)
	errObj.Set("toString", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(fmt.Sprintf("(Promise) %s", msg))
	})
	return errObj
}
