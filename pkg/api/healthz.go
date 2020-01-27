package api

import (
	"io"
	"net/http"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
