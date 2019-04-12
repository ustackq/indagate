package http

import (
	"context"
	"net/http"
)

func encodeResponse(ctx context.Context, rw http.ResponseWriter, code int, res interface{}) error {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	return json.NewEncoder(rw).Encode(res)
}
