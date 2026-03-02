package handler

import "net/http"

func (h *URLHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, err := h.srv.GetOriginalURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
