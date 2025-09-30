package broadcast

import (
	"errors"
	"sync"
)

type Broadcaster struct {
	mutex    sync.RWMutex
	channels map[string]*channel
}

type channel struct {
	mutex     sync.RWMutex
	lastID    int
	listeners map[int]*Listener
}

func New() *Broadcaster {
	return &Broadcaster{
		channels: make(map[string]*channel),
	}
}

func (b *Broadcaster) Send(chanName string) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	c := b.channels[chanName]
	if c == nil {
		return nil
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	allSent := true
	for _, l := range c.listeners {
		select {
		case l.C <- struct{}{}:
		default:
			allSent = false
		}
	}

	if !allSent {
		return errors.New("not broadcast to all listeners")
	}
	return nil
}

type Listener struct {
	C  chan struct{}
	id int

	channel *channel
}

func (b *Broadcaster) Listen(chanName string) *Listener {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	c := b.channels[chanName]
	if c == nil {
		c = &channel{
			listeners: make(map[int]*Listener),
		}
		b.channels[chanName] = c
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lastID++

	l := &Listener{
		id:      c.lastID,
		C:       make(chan struct{}),
		channel: c,
	}
	c.listeners[l.id] = l
	return l
}

func (l *Listener) Close() {
	l.channel.mutex.Lock()
	defer l.channel.mutex.Unlock()
	delete(l.channel.listeners, l.id)
	close(l.C)
}
