package gopubsub

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Agent struct {
	mu     sync.Mutex
	subs   map[string][]*Subscriber
	quit   chan struct{}
	closed bool
}

func NewAgent() *Agent {
	return &Agent{
		subs: make(map[string][]*Subscriber),
		quit: make(chan struct{}),
	}
}

type MsgT string

func (a *Agent) Publish(topic string, msg MsgT) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return fmt.Errorf("channel has been closed")
	}

	for _, ch := range a.subs[topic] {
		if len(ch.Msg) < cap(ch.Msg) {
			ch.Msg <- msg
			continue
		}
		if len(ch.Msg) == cap(ch.Msg) {
			log.Printf("gopubsub Warning: suber:[%s] from topic: [%s]  channel is full, please check if suber not to cancel, or increase suber's buffsize\n", ch.name, topic)
			// return fmt.Errorf("suber:[%s] from topic: [%s]  channel is full, please check if suber not to cancel, or increase suber's buffsize", ch.name, topic)
			go a.pushWithTimeout(topic, ch, msg, time.Minute)
			continue
		}
	}
	return nil
}

func (a *Agent) pushWithTimeout(topic string, suber *Subscriber, msg MsgT, timeout time.Duration) {
	finish := make(chan bool)
	go func() {
		defer func() {
			recover()
		}()
		suber.Msg <- msg
		finish <- true
	}()

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-finish:
			return
		default:
			if time.Now().After(deadline) {
				log.Println("suber", suber.name, "from topic:", topic, "perhaps go wrong, stop to publish to this suber")
				a.Unsubscribe(suber)
				close(finish)
				return
			}
		}
	}
}

type Subscriber struct {
	id   int
	name string
	Msg  chan MsgT
}

func (s *Subscriber) GetSuberName() string {
	return s.name
}

var subscriberIdSet = make(map[int]bool)

func NewSubscriber(bufferSize int, name ...string) *Subscriber {
	for i := 0; i < 1000000; i++ {
		if _, ok := subscriberIdSet[i]; !ok {
			subscriberIdSet[i] = true
			subname := "noname"
			if len(name) > 0 {
				subname = name[0]
			}
			return &Subscriber{
				id:   i,
				name: subname,
				Msg:  make(chan MsgT, bufferSize),
			}
		}
	}
	panic("too many subscribers")
}

type SubscribeConfig struct {
	BufferSize int
}

type SubscribeOption func(config *SubscribeConfig) error

func SetBufferSizeOpt(buffersize int) SubscribeOption {
	return func(config *SubscribeConfig) error {
		if buffersize <= 0 {
			config.BufferSize = 10
		} else {
			config.BufferSize = buffersize
		}
		return nil
	}
}

func (a *Agent) Subscribe(topic string, options ...SubscribeOption) (suber *Subscriber, cancel func(a *Agent, suber *Subscriber)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil, func(a *Agent, suber *Subscriber) {}
	}

	subconfig := SubscribeConfig{BufferSize: 10}

	for _, op := range options {
		op(&subconfig)
	}

	suber = NewSubscriber(subconfig.BufferSize)
	a.subs[topic] = append(a.subs[topic], suber)
	return suber, func(a *Agent, suber *Subscriber) { a.Unsubscribe(suber) }
}

func (a *Agent) Unsubscribe(suber *Subscriber) {
	defer func() {
		recover()
	}()
	defer close(suber.Msg)
	a.mu.Lock()
	defer a.mu.Unlock()
	for topic, subs := range a.subs {
		for i, sub := range subs {
			if sub.id == suber.id {
				subs = append(subs[:i], subs[i+1:]...)
				a.subs[topic] = subs
				return
			}
		}
	}
}

func (a *Agent) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}

	a.closed = true
	close(a.quit)

	for _, ch := range a.subs {
		for _, sub := range ch {
			func() {
				defer func() {
					recover()
				}()
				close(sub.Msg)
			}()
		}
	}
}

// get channel value, if channel is closed, return false
// if channnel is not closed, wait until get one value and return it
func GetChannelValue[T interface{}](ch chan T) (res T, IsNotClosed bool) {
	IsNotClosed = true
	for res = range ch {
		return
	}
	return *new(T), false
}
