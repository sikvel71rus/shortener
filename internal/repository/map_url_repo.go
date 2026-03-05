package repository

import "sync"

type MapURLRepo struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMapURLRepo() *MapURLRepo {
	return &MapURLRepo{urls: make(map[string]string)}
}

func (r *MapURLRepo) SaveURL(id string, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.urls[id] = url
}

func (r *MapURLRepo) GetURL(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[id]
	return url, ok
}

func (r *MapURLRepo) CheckIfURLExist(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.urls[id]
	return ok
}
