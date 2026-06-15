package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

// Record describes one persisted file-storage event.
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
	IsDeleted   bool   `json:"is_deleted,omitempty"`
}

// Consumer reads storage events from a file.
type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

// NewConsumer opens a file-backed event reader.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

// ReadEvent reads and decodes the next storage event from the file.
func (c *Consumer) ReadEvent() (*Record, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	record := Record{}
	err := json.Unmarshal(data, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// Close closes the underlying consumer file.
func (c *Consumer) Close() error {
	if c.file == nil {
		return nil
	}
	return c.file.Close()
}

// Producer appends storage events to a file.
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// NewProducer opens a file-backed event writer.
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// WriteEvent writes one storage event to the file.
func (p *Producer) WriteEvent(record *Record) error {
	return p.encoder.Encode(record)
}

// Close closes the underlying producer file.
func (p *Producer) Close() error {
	if p.file == nil {
		return nil
	}
	return p.file.Close()
}
