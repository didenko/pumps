package pumps

import (
	"reflect"
)

type GenericChannel chan interface{}

type target struct {
	ch  GenericChannel
	typ reflect.Type
}

type Broadcaster struct {
	Post      GenericChannel
	Subscribe chan chan interface{}
	users     []target
}

func MakeBroadcaster(bufsize int) *Broadcaster {

	bcaster := Broadcaster{
		make(chan interface{}, bufsize),
		make(chan chan interface{}, 1),
		[]target{},
	}

	go bcaster.subscriptions_listener()
	go bcaster.messages_listener()

	return &bcaster
}

func (bcaster *Broadcaster) subscriptions_listener() {

	for {
		user_channel := <-bcaster.Subscribe
		elem_type := reflect.TypeOf(user_channel).Elem()

		bcaster.users = append(bcaster.users, target{
			user_channel,
			elem_type,
		})
	}
}

func (bcaster *Broadcaster) messages_listener() {

	for {
		msg := <-bcaster.Post
		msg_type := reflect.TypeOf(msg)

		for _, tgt := range bcaster.users {
			if msg_type.AssignableTo(tgt.typ) {
				go func() {

					tgt.ch <- msg
				}()
			}
		}
	}
}
