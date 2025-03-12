package runjs

import (
	"time"

	"github.com/dop251/goja"
)

// createAsync defines async(fn) -> returns a function that auto-runs and returns a Promise
func createAsync(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		fn, isFunc := goja.AssertFunction(call.Argument(0))
		if !isFunc {
			return vm.NewTypeError("async() expects a function")
		}

		// Return a function that executes immediately and returns a Future
		return vm.ToValue(func(innerCall goja.FunctionCall) goja.Value {
			resultChan := make(chan goja.Value, 1)
			errorChan := make(chan goja.Value, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						errorChan <- vm.NewGoError(r.(error))
					}
				}()

				// Call function
				result, err := fn(goja.Undefined(), innerCall.Arguments...)
				if err != nil {
					errorChan <- NewPromiseError(vm, err.Error())
					return
				}

				resultChan <- result
			}()

			// Create a Future-like object
			future := vm.NewObject()
			future.Set("wait", func(call goja.FunctionCall) goja.Value {
				select {
				case result := <-resultChan:
					return result
				case err := <-errorChan:
					panic(err) // Ensure error gets propagated
				}
			})

			return future
		})
	})
}

// createAwait implements await(promise) -> waits for resolution or throws error
func createAwait(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		promiseObj := call.Argument(0).ToObject(vm)
		waitFunc, ok := goja.AssertFunction(promiseObj.Get("wait"))
		if !ok {
			return vm.NewTypeError("await() expects a valid Future object")
		}

		// Call wait() and return result
		result, err := waitFunc(goja.Undefined())
		if err != nil {
			panic(err) // Ensure error is properly thrown
		}
		return result
	})
}

// createSleep implements sleep(ms) -> returns a Future that resolves after ms milliseconds
func createSleep(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		ms := call.Argument(0).ToInteger()
		resultChan := make(chan goja.Value, 1)

		go func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			resultChan <- vm.ToValue(nil)
		}()

		// Create Future-like object
		future := vm.NewObject()
		future.Set("wait", func(call goja.FunctionCall) goja.Value {
			return <-resultChan
		})

		return future
	})
}

// Register Goja functions
func RegisterFunctions(vm *goja.Runtime) {
	vm.Set("async", createAsync(vm))
	vm.Set("await", createAwait(vm))
	vm.Set("sleep", createSleep(vm))
}
