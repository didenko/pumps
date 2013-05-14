package cherr

import (
	"log"
	"os"
	"runtime"
	"sync"
)

type ErrHandler func(err error)

type callback func()

type fault struct {
	err  error
	stop bool
}

var def int
var faults = make(chan *fault, 1)
var handlers = make(map[int]ErrHandler)
var hwm = 0 // high water mark

func PostAndGo(err error) {
	Post(err, false)
}

func PostAndStop(err error) {
	Post(err, true)
}

func Post(err error, stop bool) {
	if err != nil {
		faults <- &fault{err, stop}
	}
}

func Close() {
	close(faults)
}

func next() {
	hwm++
}

func AddHandler(h ErrHandler) int {
	defer next()
	handlers[hwm] = h
	return hwm
}

func DelHandler(idx int) bool {
	if _, present := handlers[idx]; present {
		delete(handlers, idx)
		return true
	}

	return false
}

func RemoveDefault() {
	DelHandler(def)
}

func AddDefault() {
	def = AddHandler(defaultHandler)
}

func defaultHandler(err error) {
	log.Println(err)
}

func process(f *fault) {
	var waiter sync.WaitGroup
	onDone := func() {
		waiter.Done()
		return
	}

	for _, handler := range handlers {
		waiter.Add(1)
		go wrap_handler(handler, f, onDone)
	}

	waiter.Wait()

	if f.stop {
		os.Exit(1)
	}
}

func wrap_handler(handler ErrHandler, f *fault, onDone callback) {

	defer func() {

		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			log.Fatal(r.(error))
		}

		onDone()
	}()

	handler(f.err)
}

func init() {
	AddDefault()

	go func() {
		for f := range faults {
			process(f)
		}
	}()
}
