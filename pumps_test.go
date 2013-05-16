package pumps

import (
	"testing"
	"time"
)

func TestBroadcast(t *testing.T) {

	bcast := MakeBroadcaster(1)

	user1 := make(chan int, 1)

	bcast.Subscribe <- user1

	bcast.Post <- 5

	select {
	case got1 := <-user1:
		if got1 != 5 {
			t.Error("Got a wrong value on first channel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Did not receive expected values on channels")
	}

}
