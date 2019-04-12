package http

import (
	"context"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"go.uber.org/zap"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// IndagateErrorCode define the error code of indagate error
	IndagateErrorCode = "X-Indagate-Error-Code"
)

// LogEncodeError log rest request error
func LogEncodeError(logger *zap.Logger, r *http.Request, err error) {
	logger.Info("err encoding response",
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.Error(err),
	)
}

// EncodeError encodes err with the appropriate status code and format,
// sets the X-Indagate-Error-Code headers on the response.
func EncodeError(ctx context.Context, err error, rw http.ResponseWriter) {
	if err == nil {
		return
	}
	code := errors.ErrorCode(err)
	httpCode, ok := statusCodeIndagateError[code]
	if !ok {
		httpCode = http.StatusBadRequest
	}
	rw.Header().Set(IndagateErrorCode, code)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(httpCode)

	var e error
	if err, ok := err.(*errors.Error); ok {
		e = &errors.Error{
			Code: code,
			Err:  err.Err,
		}
	} else {
		e = &errors.Error{
			Code: errors.Internal,
			Err:  err.Err,
		}
	}
	b, _ := json.Marshal(e)
	rw.Write(b)
}

func UnauthorizedError(ctx context.Context, rw http.ResponseWriter) {
	EncodeError(ctx, &errors.Error{
		Code: "",
		Msg:  "unahorized access",
	}, rw)
}

var statusCodeIndagateError = map[string]int{}
