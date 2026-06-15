package audit

import "context"

const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}

type Observer interface {
	Notify(ctx context.Context, event Event) error
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type Broadcaster struct {
	observers []Observer
}

func NewBroadcaster(observers ...Observer) *Broadcaster {
	filtered := make([]Observer, 0, len(observers))
	for _, observer := range observers {
		if observer != nil {
			filtered = append(filtered, observer)
		}
	}

	return &Broadcaster{observers: filtered}
}

func (b *Broadcaster) Publish(ctx context.Context, event Event) error {
	var result error

	for _, observer := range b.observers {
		if err := observer.Notify(ctx, event); err != nil && result == nil {
			result = err
		}
	}

	return result
}
