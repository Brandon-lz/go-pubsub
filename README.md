# go-pubsub
golang publish-subscribe lib

Are you wondering why golang, as a language that natively supports concurrency, only implements the message queue channel in its syntax, but does not implement the publish-subscribe pattern? This library leverages channels to encapsulate the publish-subscribe pattern, providing convenience for concurrent programming.

## installation
```shell
go get  "github.com/Brandon-lz/go-pubsub"
```

## usage
main.go
```go
package main

import (
    "fmt"
    "sync"

    "github.com/Brandon-lz/go-pubsub"
)

func main() {
    // channel to publish messages to
	// Create a new agent
	agent := gopubsub.NewAgent()
	defer agent.Close()

	// Subscribe to a topic
	sub, _ := agent.Subscribe("foo")
	sub2, cancel := agent.Subscribe("foo")
	defer cancel(agent, sub2) // remember to unsubscribe !
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
}
```
