package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type FileObserver struct {
	path string
	mu   sync.Mutex
}

// NewFileObserver creates an observer that appends audit events to a file.
func NewFileObserver(path string) *FileObserver {
	if path == "" {
		return nil
	}

	return &FileObserver{path: path}
}

// Notify writes the audit event as one JSON line to the configured file.
func (o *FileObserver) Notify(_ context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	file, err := os.OpenFile(o.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(payload, '\n')); err != nil {
		return err
	}

	return nil
}
