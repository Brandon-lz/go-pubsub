package gopubsub

import (
    "sync"
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


func (a *Agent) Publish(topic string, msg MsgT) {
    a.mu.Lock()
    defer a.mu.Unlock()

    if a.closed {
        return
    }

    for _, ch := range a.subs[topic] {
        ch.Msg <- msg
    }
}


type Subscriber struct {
    id  int
    Msg chan MsgT
}

var subscriberIdSet = make(map[int]bool)

func NewSubscriber(bufferSize int) *Subscriber {
    for i := 0; i < 1000000; i++ {
        if _, ok := subscriberIdSet[i]; !ok {
            subscriberIdSet[i] = true
            return &Subscriber{
                id:  i,
                Msg: make(chan MsgT, bufferSize),
            }
        }
    }
    panic("too many subscribers")
}

func (a *Agent) Subscribe(topic string) *Subscriber {
    a.mu.Lock()
    defer a.mu.Unlock()

    if a.closed {
        return nil
    }

    suber := NewSubscriber(10)
    a.subs[topic] = append(a.subs[topic], suber)
    return suber
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

