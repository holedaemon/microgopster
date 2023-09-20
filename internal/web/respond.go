package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

type Error struct {
	Message string `json:"message"`
}

func respondErrorf(w http.ResponseWriter, r *http.Request, status int, msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	respondError(w, r, status, msg)
}

func respondError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	render.Status(r, status)
	render.JSON(w, r, &Error{
		Message: msg,
	})
}

func respond(w http.ResponseWriter, r *http.Request, status int, v any) {
	render.Status(r, status)
	render.JSON(w, r, v)
}
