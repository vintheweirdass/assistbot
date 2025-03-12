package runjs

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// Future (Custom async handling)
func createFuture(vm *goja.Runtime) goja.Value {
	constructor := func(call goja.ConstructorCall) *goja.Object {
		executor := call.Argument(0)
		if executor == nil || goja.IsUndefined(executor) || goja.IsNull(executor) {
			return vm.NewTypeError("Future requires an executor function")
		}

		resultChan := make(chan goja.Value, 1)
		errorChan := make(chan goja.Value, 1)
		hasCatch := false

		go func() {
			defer func() {
				if r := recover(); r != nil {
					errorChan <- vm.NewTypeError("Error Exception: %v", r)
				}
			}()

			execFunc, ok := executor.Export().(func() goja.Value)
			if !ok {
				errorChan <- vm.NewTypeError("Executor must be a function")
				return
			}

			res := execFunc()
			resultChan <- res
		}()

		obj := vm.NewObject()
		obj.Set("then", func(call goja.FunctionCall) goja.Value {
			successCallback := call.Argument(0)
			errorCallback := call.Argument(1)

			go func() {
				select {
				case res := <-resultChan:
					if successCallback != nil && !goja.IsUndefined(successCallback) {
						vm.RunString(fmt.Sprintf("(%s)(%v)", successCallback.String(), res))
					}
				case err := <-errorChan:
					if errorCallback != nil && !goja.IsUndefined(errorCallback) {
						hasCatch = true
						vm.RunString(fmt.Sprintf("(%s)(%v)", errorCallback.String(), err))
					}
				}
			}()
			return obj
		})

		obj.Set("catch", func(call goja.FunctionCall) goja.Value {
			hasCatch = true
			errorCallback := call.Argument(0)

			go func() {
				err := <-errorChan
				if errorCallback != nil && !goja.IsUndefined(errorCallback) {
					vm.RunString(fmt.Sprintf("(%s)(%v)", errorCallback.String(), err))
				}
			}()
			return obj
		})

		go func() {
			err := <-errorChan
			if !hasCatch {
				fmt.Println("Uncaught Promise Exception:", err)
			}
		}()

		return obj
	}

	return vm.ToValue(constructor)
}

// Promise (JS-style async handling)
func createPromise(vm *goja.Runtime) goja.Value {
	constructor := func(call goja.ConstructorCall) *goja.Object {
		executor := call.Argument(0)
		if executor == nil || goja.IsUndefined(executor) || goja.IsNull(executor) {
			return vm.NewTypeError("Promise requires an executor function")
		}

		resultChan := make(chan goja.Value, 1)
		errorChan := make(chan goja.Value, 1)
		hasCatch := false

		go func() {
			defer func() {
				if r := recover(); r != nil {
					errorChan <- vm.NewTypeError("Promise Exception: %v", r)
				}
			}()

			resolver := func(value goja.Value) { resultChan <- value }
			rejector := func(reason goja.Value) { errorChan <- reason }

			executor.Export().(func(goja.Value, goja.Value))(vm.ToValue(resolver), vm.ToValue(rejector))
		}()

		obj := vm.NewObject()
		obj.Set("then", func(call goja.FunctionCall) goja.Value {
			successCallback := call.Argument(0)
			errorCallback := call.Argument(1)

			go func() {
				select {
				case res := <-resultChan:
					if successCallback != nil && !goja.IsUndefined(successCallback) {
						vm.RunString(fmt.Sprintf("(%s)(%v)", successCallback.String(), res))
					}
				case err := <-errorChan:
					if errorCallback != nil && !goja.IsUndefined(errorCallback) {
						hasCatch = true
						vm.RunString(fmt.Sprintf("(%s)(%v)", errorCallback.String(), err))
					}
				}
			}()
			return obj
		})

		obj.Set("catch", func(call goja.FunctionCall) goja.Value {
			hasCatch = true
			errorCallback := call.Argument(0)

			go func() {
				err := <-errorChan
				if errorCallback != nil && !goja.IsUndefined(errorCallback) {
					vm.RunString(fmt.Sprintf("(%s)(%v)", errorCallback.String(), err))
				}
			}()
			return obj
		})

		go func() {
			err := <-errorChan
			if !hasCatch {
				fmt.Println("Uncaught Promise Exception:", err)
			}
		}()

		return obj
	}

	return vm.ToValue(constructor)
}

// Sleep (safe version)
func createSleep(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		ms := call.Argument(0).ToInteger()

		// Create a Future that resolves after `ms` milliseconds
		constructorCall := goja.ConstructorCall{
			This: vm.NewObject(),
			Arguments: []goja.Value{
				vm.ToValue(func() goja.Value {
					time.Sleep(time.Duration(ms) * time.Millisecond)
					return goja.Undefined() // Resolve with `undefined`
				}),
			},
		}

		// Call Future() constructor
		return createFuture(vm).Export().(func(goja.ConstructorCall) *goja.Object)(constructorCall)
	})
}

// setTimeout (safe async execution)
func createSetTimeout(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		callback := call.Argument(0)
		delay := call.Argument(1).ToInteger()

		if callback == nil || goja.IsUndefined(callback) {
			return vm.NewTypeError("setTimeout requires a function")
		}

		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			vm.RunString(fmt.Sprintf("(%s)()", callback.String()))
		}()

		return goja.Undefined()
	})
}

// async() shorthand for Future
func createAsync(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		constructorCall := goja.ConstructorCall{
			This:      vm.NewObject(), // New empty object as `this`
			Arguments: call.Arguments, // Forward all arguments
		}
		return createFuture(vm).Export().(func(goja.ConstructorCall) *goja.Object)(constructorCall)
	})
}

// await() (blocking wait for Promise/Future)
// await() (blocking wait for Future/Promise)
func createAwait(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		promise := call.Argument(0)

		// Ensure it's an object
		obj := promise.ToObject(vm)
		if obj == nil {
			return vm.NewTypeError("await() requires a Future or Promise object")
		}

		// Get the "then" method
		thenMethod := obj.Get("then")
		thenCallable, ok := goja.AssertFunction(thenMethod)
		if !ok {
			return vm.NewTypeError("await() requires an object with a 'then' method")
		}

		// Channel to receive the resolved value
		resultChan := make(chan goja.Value, 1)

		// Call then() with success and error handlers
		_, err := thenCallable(obj, obj, vm.ToValue(func(value goja.Value) {
			resultChan <- value
		}))
		if err != nil {
			return vm.NewTypeError(err)
		}

		// Block and wait for result
		return <-resultChan
	})
}

// Register everything in Goja
func RegisterFunctions(vm *goja.Runtime) {
	vm.Set("Future", createFuture(vm))
	vm.Set("Promise", createPromise(vm))
	vm.Set("sleep", createSleep(vm))
	vm.Set("setTimeout", createSetTimeout(vm))
	vm.Set("async", createAsync(vm))
	vm.Set("await", createAwait(vm))
}
