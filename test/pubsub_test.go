package test

import (
	"fmt"
	"sync"
	"testing"

	gopubsub "github.com/Brandon-lz/go-pubsub"
)

func TestChanSub(t *testing.T) {
	// channel to publish messages to
	// Create a new agent
	agent := gopubsub.NewAgent()
	defer agent.Close()

	// Subscribe to a topic
	sub, _ := agent.Subscribe("foo")
	sub2, cancel := agent.Subscribe("foo")
	defer cancel(agent, sub2) // recommend using defer to ensure unsubscribe run finally
	sub3, cancel := agent.Subscribe("foo")
	defer cancel(agent, sub3)

	// Publish a message to the topic
	var wg = &sync.WaitGroup{}

	wg.Add(3)
	go func() {
		defer wg.Done()
		fmt.Println(<-sub.Msg)
	}()

	go func() {
		defer wg.Done()
		fmt.Println(<-sub2.Msg)
	}()

	go func() {
		defer wg.Done()
		fmt.Println(<-sub3.Msg)
	}()

	go agent.Publish("foo", "hello world")

	wg.Wait()

	agent.Unsubscribe(sub)
	fmt.Println("subscrip closed suber")

	if msg, ok := gopubsub.GetChannelValue(sub.Msg); ok {
		fmt.Println("sub", msg)
	} else {
		fmt.Println("channel is closed")
	}

	suber4, cancel := agent.Subscribe("foo2")
	defer cancel(agent, suber4)
	for _ = range 11 {
		agent.Publish("foo2", "hello") // will block on 11, but will auto wait for 1 minite, when timeout, force unsubscribe the blocked suber
	}

	// wait ctrl c
	quit := make(chan bool)
	<-quit
}
