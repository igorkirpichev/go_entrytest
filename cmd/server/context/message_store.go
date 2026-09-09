package context

import (
	"sync"
	"time"
)

type Message struct {
	Id        uint32
	Message   string
	CreatedAt time.Time
}

type MessageStore struct {
	data      []Message
	nextIndex uint64
	mutex     sync.RWMutex
}

func CreateMessageStore() *MessageStore {
	return &MessageStore{
		data:      make([]Message, 0),
		nextIndex: 0,
	}
}

func (store *MessageStore) Insert(message string) Message {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	msg := Message{Id: uint32(store.nextIndex), Message: message, CreatedAt: time.Now()}
	store.data = append(store.data, msg)
	store.nextIndex++

	return msg
}
