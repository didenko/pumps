package cherr

import (
	"errors"
	_ "reflect"
	"sync"
	"testing"
)

var h ErrHandler

func TestAddAndTriggerHandler(t *testing.T) {

	var wg sync.WaitGroup

	failed := true

	h = func() ErrHandler {
		return func(e error) {
			failed = false
			wg.Done()
		}
	}()

	AddHandler(h)

	wg.Add(1)
	PostAndGo(errors.New("Add handler test"))
	wg.Wait()

	if failed {
		t.Error("Failed adding a handler")
	}
}

func TestHandlerPanic(t *testing.T) {

	var wg sync.WaitGroup

	h = func() ErrHandler {
		return func(e error) {
			panic(errors.New("Panicked!"))
			wg.Done()
		}
	}()

	AddHandler(h)

	wg.Add(1)
	PostAndGo(errors.New("Handler panic test"))
	wg.Wait()

	t.Error("Failed halting on panic")
}
