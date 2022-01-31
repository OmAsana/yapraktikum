package server

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

var compressor = middleware.NewCompressor(
	5,
)

func compressorHandler(next http.Handler) http.Handler {
	return compressor.Handler(next)
}
