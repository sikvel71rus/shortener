package repository

type URLRepo interface {
	SaveURL(id string, url string)
	GetURL(id string) (string, bool)
}

type MemoryRepo struct {
	urls map[string]string
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{urls: make(map[string]string)}
}

func (r *MemoryRepo) SaveURL(id string, url string) {
	r.urls[id] = url
}

func (r *MemoryRepo) GetURL(id string) (string, bool) {
	url, ok := r.urls[id]
	return url, ok
}
