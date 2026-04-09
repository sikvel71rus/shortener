package service

import (
	"errors"
	"math/rand"
	"strings"
)

type URLRepo interface {
	SaveURL(id string, originalURL string) error
	GetURL(id string) (string, bool)
	CheckIfURLExist(id string) bool
}

type URLService struct {
	repo    URLRepo
	baseURL string
}

func NewURLService(repo URLRepo, baseURL string) *URLService {
	return &URLService{repo: repo, baseURL: baseURL}
}

func (s *URLService) ShortenURL(url string) string {
	id := ""
	for {
		id = generateID()
		if !s.repo.CheckIfURLExist(id) {
			break
		}
	}

	s.repo.SaveURL(id, url)
	return s.baseURL + "/" + id
}

func (s *URLService) GetOriginalURL(id string) (string, error) {
	url, ok := s.repo.GetURL(id)
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}

func generateID() string {
	length := 6
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
