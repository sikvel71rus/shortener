package repository

type MapURLRepo struct {
	urls map[string]string
}

func NewMapURLRepo() *MapURLRepo {
	return &MapURLRepo{urls: make(map[string]string)}
}

func (r *MapURLRepo) SaveURL(id string, url string) {
	r.urls[id] = url
}

func (r *MapURLRepo) GetURL(id string) (string, bool) {
	url, ok := r.urls[id]
	return url, ok
}

func (r *MapURLRepo) CheckIfURLExist(id string) bool {
	_, ok := r.urls[id]
	return ok
}
