// Copyright (c) 2013 Vlad Didenko. All rights reserved

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Vlad Didenko nor the names of other
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package pumps is a collection of channel manipulators. Contained objects expected to communicate via channels, including meta-communication.
package pumps

import (
	"reflect"
)

// target struct stores channels which subscribed to a specific fan. With each consumer channel the target keeps it's reflect.Type pre-extracted to speed up message processing.
type target struct {
	ch  reflect.Value
	typ reflect.Type
}

// FanOut stores the in- and out-flow channel information. FanOut should not be instantiated directly, but rather via the MakeFanOut function. As this object expected to be used to route errors it itself does not emit errors to avoid ambiguity. For example, if a non-channel value fed into the Outs intake channel the code will panic while attempting to reflect on its element's type or later when it tries to post into it.
type FanOut struct {

	// Post channel accepts messages to be fanned out to relevant downstream channels provided by users. The Post channel's buffer size provided as a parameter to MakeFanOut. Sending nil to the Post channel stops the FanOut goroutines, closes all channels (including user-provided channels!) and allows the object to be garbage-collected.
	Post chan interface{}

	// Outs channel accepts "subscriber" channels. Messages received on the Post channel will be sent to each of the channels received earlier on the via Outs, where the message value is assignable to the out-flow channel value. Sending nil to the Outs channel stops subscription goroutine and closes the Outs channel - preventing future subscriptions.
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
	for _, tgt := range fan.users {
		tgt.ch.Close()
	}
}
