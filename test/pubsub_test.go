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
    sub := agent.Subscribe("foo")
    sub2 := agent.Subscribe("foo")
    defer agent.Unsubscribe(sub2)
    sub3 := agent.Subscribe("foo")
    defer agent.Unsubscribe(sub3)
	
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

    if msg,ok := gopubsub.GetChannelValue(sub.Msg);ok{
        fmt.Println("sub", msg)
    }else{
        fmt.Println("channel is closed")
    }

}
