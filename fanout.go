// Package pumps is a collection of channel manipulators. Contained objects expected to communicate via channels, including meta-communication.
package pumps

import (
	"reflect"
)

// target struct stores channels which subscribed to a specific fan. With each consumer channel target keeps pre-extracted it's reflect.Type to speed up message processing.
type target struct {
	ch  reflect.Value
	typ reflect.Type
}

// FanOut stores the in- and out-flow channel information. FanOut should not be instantiated directly, but rather via the MakeFanOut function. As this object expected to be used to route errors it itself does not emit errors to avoid ambiguity. For example, if a no-channel value fed into the Outs intake channel the code will panic while attempting to reflect on its element's type or later when it tries to post into it.
type FanOut struct {

	// Post channel accepts messages to be fanned out to relevant downstream channels provided by users. The Post channel buffer size provided as a parameter to MakeFanOut. Sending nil to the Post channel stops the FanOut goroutines, closes all channels and allows the object to be garbage-collected.
	Post chan interface{}

	// Outs channel accepts "subscriber" channels. Messages received on the Post channel will be sent to each of channels received earlier on the via Outs, where the meggase value is assignable to the ouf-flow channel value. Sending nil to the Outs channel stops subscription goroutine and closes the Outs channel - preventing future subscriptions.
	Outs  chan interface{}
	users []target
}

// The MakeFanOut function is the main entry point. The only parameter to it is the buffer size of the intake channel.
func MakeFanOut(bufsize int) *FanOut {

	fan := &FanOut{
		make(chan interface{}, bufsize),
		make(chan interface{}, 1),
		[]target{},
	}

	go fan.subscrLoop()
	go fan.messageLoop()

	return fan
}

func (fan *FanOut) subscrLoop() {

	for userCh := range fan.Outs {

		if userCh == nil {
			close(fan.Outs)
			fan.Outs = nil
			break
		}

		elemType := reflect.TypeOf(userCh).Elem()

		fan.users = append(fan.users, target{
			reflect.ValueOf(userCh),
			elemType,
		})
	}
}

func (fan *FanOut) messageLoop() {

	for msg := range fan.Post {

		if msg == nil {
			fan.closeAll()
			break
		}

		msgType := reflect.TypeOf(msg)

		for _, tgt := range fan.users {
			if msgType.AssignableTo(tgt.typ) {
				go func(chValue reflect.Value) {
					chValue.Send(reflect.ValueOf(msg))
				}(tgt.ch)
			}
		}
	}
}

func (fan *FanOut) closeAll() {
	if fan.Outs != nil {
		fan.Outs <- nil
	}
	close(fan.Post)
}
