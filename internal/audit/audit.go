package audit

import "context"

const (
	// ActionShorten marks an audit event for short URL creation.
	ActionShorten = "shorten"
	// ActionFollow marks an audit event for following a short URL.
	ActionFollow = "follow"
)

// Event describes one audit record emitted by the server.
type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}

// Observer receives audit events from a publisher.
type Observer interface {
	Notify(ctx context.Context, event Event) error
}

// Publisher broadcasts audit events to one or more observers.
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// Broadcaster fan-outs one audit event to all configured observers.
type Broadcaster struct {
	observers []Observer
}

// NewBroadcaster creates a publisher for the provided observers.
func NewBroadcaster(observers ...Observer) *Broadcaster {
	filtered := make([]Observer, 0, len(observers))
	for _, observer := range observers {
		if observer != nil {
			filtered = append(filtered, observer)
		}
	}

	return &Broadcaster{observers: filtered}
}

// Publish sends an event to every configured observer.
func (b *Broadcaster) Publish(ctx context.Context, event Event) error {
	var result error

	for _, observer := range b.observers {
		if err := observer.Notify(ctx, event); err != nil && result == nil {
			result = err
		}
	}

	return result
}
