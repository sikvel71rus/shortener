package service

import (
	"errors"
	"math/rand"
	"strings"
)

type URLRepo interface {
	SaveURL(id string, url string)
	GetURL(id string) (string, bool)
}

type URLService struct {
	repo URLRepo
}

func NewURLService(repo URLRepo) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) ShortenURL(url string) string {
	id := generateID()

	s.repo.SaveURL(id, url)
	return id
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
