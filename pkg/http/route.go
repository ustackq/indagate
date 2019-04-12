package http

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/ustackq/indagate/pkg/logger"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
)

// NewRouter return a new router with other handler.
func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowed = http.HandlerFunc(methodNotAllowedHandler)
	router.PanicHandler = panicHandler
	return router
}

func notFoundHandler(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := &errors.Error{
		Code: errors.NotFound,
		Msg:  "Resource not found",
	}
	EncodeError(ctx, err, rw)
}

func methodNotAllowedHandler(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allow := rw.Header().Get("allodw")
	err := &errors.Error{
		Code: errors.MethodNotAllowed,
		Msg:  fmt.Sprintf("allow: %s", allow),
	}
	EncodeError(ctx, err, rw)

}
func panicHandler(rw http.ResponseWriter, r *http.Request, rcv interface{}) {
	ctx := r.Context()
	err := &errors.Error{
		Code: errors.Internal,
		Msg:  "a panic has occurred.",
		Err:  fmt.Errorf("%v", rcv),
	}
	logger := getPanicLogger()
	logger.Error(
		err.Msg,
		zap.String("err", err.Err.Error()),
		zap.String("stack", fmt.Sprintf("%s", debug.Stack())),
	)
	EncodeError(ctx, err, rw)
}

var (
	panicLogger     *zap.Logger
	panicLoggerOnce sync.Once
)

func getPanicLogger() *zap.Logger {
	panicLoggerOnce.Do(func() {
		panicLogger = logger.New(os.Stderr)
		panicLogger = panicLogger.With(zap.String("handler", "panic"))
	})
	return panicLogger
}
