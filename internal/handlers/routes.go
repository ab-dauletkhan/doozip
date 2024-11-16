package handlers

import "net/http"

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}
