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

func (b *Agent) Publish(topic string, msg string) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.closed {
        return
    }

    for _, ch := range b.subs[topic] {
        ch.Msg <- msg
    }
}

type Subscriber struct {
    id  int
    Msg chan string
}

var subscriberIdSet = make(map[int]bool)

func NewSubscriber() *Subscriber {
    for i := 0; i < 1000000; i++ {
        if _, ok := subscriberIdSet[i]; !ok {
            subscriberIdSet[i] = true
            return &Subscriber{
                id:  i,
                Msg: make(chan string),
            }
        }
    }
    panic("too many subscribers")
}

func (b *Agent) Subscribe(topic string) *Subscriber {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.closed {
        return nil
    }

    suber := NewSubscriber()
    b.subs[topic] = append(b.subs[topic], suber)
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

func (b *Agent) Close() {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.closed {
        return
    }

    b.closed = true
    close(b.quit)

    for _, ch := range b.subs {
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

