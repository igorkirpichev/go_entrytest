package context

import (
	"slices"
	"sync"
	"time"
)

type Message struct {
	Id        uint64
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

	msg := Message{Id: store.nextIndex, Message: message, CreatedAt: time.Now()}
	store.data = append(store.data, msg)
	store.nextIndex++

	return msg
}

func (store *MessageStore) List() []Message {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	copied := make([]Message, 0, len(store.data))
	for _, message := range slices.Backward(store.data) {
		copied = append(copied, message)
	}

	return copied
}
