package pumps

import (
	"errors"
	"os"
	"testing"
	"time"
)

const constInt int = 42
const constStr string = "ABC"

func TestPlainError(t *testing.T) {
	var resBits int8 = 0

	bcast := MakeFanOut(2)

	AllErrors := make(chan error)
	bcast.Outs <- AllErrors
	var AllErrorsBit int8 = 1

	LinkErrors := make(chan *os.LinkError)
	bcast.Outs <- LinkErrors
	var LinkErrorsBit int8 = 2

	errPlain := errors.New(constStr)

	bcast.Post <- errPlain

	select {
	case err := <-AllErrors:
		if err.Error() != constStr {
			t.Errorf("Wrong \"%s\" on generic error channel.", err.Error())
		} else {
			resBits |= AllErrorsBit
		}
	case err := <-LinkErrors:
		if err.Error() != constStr {
			t.Errorf("Wrong \"%s\" on link error channel.", err.Error())
		} else {
			resBits |= LinkErrorsBit
		}
	case <-time.After(5 * time.Millisecond):
		t.Error("Did not receive the error on generic error channel.")
	}

	bcast.Post <- nil

	if resBits != AllErrorsBit {
		t.Errorf("Incorrect result bitmap: %d", resBits)
	}
}

func TestThreeErrors(t *testing.T) {
	var resBits int8 = 0

	bcast := MakeFanOut(2)

	AllErrors := make(chan error)
	bcast.Outs <- AllErrors
	var AllErrorsBit int8 = 1

	LinkErrors := make(chan *os.LinkError)
	bcast.Outs <- LinkErrors
	var LinkErrorsBit int8 = 2

	PathErrors := make(chan *os.PathError)
	bcast.Outs <- PathErrors
	var PathErrorsBit int8 = 4

	pathError := &os.PathError{
		"Remove",
		"/mach_kernel",
		errors.New("permission denied"),
	}

	bcast.Post <- pathError

waitForResults:
	for {
		select {
		case err := <-AllErrors:
			if err.Error() != "Remove /mach_kernel: permission denied" {
				t.Errorf("Wrong \"%s\" on generic errors channel.", err.Error())
			} else {
				resBits |= AllErrorsBit
			}
		case err := <-PathErrors:
			if err.Error() != "Remove /mach_kernel: permission denied" {
				t.Errorf("Wrong \"%s\" on path errors channel.", err.Error())
			} else {
				resBits |= PathErrorsBit
			}
		case <-LinkErrors:
			resBits |= LinkErrorsBit
		case <-time.After(5 * time.Millisecond):
			break waitForResults
		}
	}

	bcast.Post <- nil

	if resBits != AllErrorsBit|PathErrorsBit {
		t.Errorf("Incorrect result bitmap: %d", resBits)
	}
}

func TestBroadcastPrimitiveAndError(t *testing.T) {

	bcast := MakeFanOut(1)

	chanInt := make(chan int)
	chanErr := make(chan error)

	bcast.Outs <- chanInt
	bcast.Outs <- chanErr

	bcast.Post <- constInt

	bcast.Outs <- nil

	select {
	case gotInt := <-chanInt:
		if gotInt != constInt {
			t.Error("Got a wrong value on the integer channel")
		}
	case <-chanErr:
		t.Error("Got value on the error channel, was not supposed to.")
	case <-time.After(5 * time.Millisecond):
		t.Error("Did not receive expected values on channels")
	}

	bcast.Post <- nil

}
