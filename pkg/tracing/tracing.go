package tracing

import (
	"context"
	"runtime"
	"strings"

	"go.opencensus.io/trace"
)

func StartSpanFromContext(ctx context.Context) (*trace.Span, context.Context) {
	if ctx == nil {
		panic("StartSpanFromContext called with nil context")
	}
	// learn runtime about callers
	var pcs [1]uintptr
	n := runtime.Callers(2, pcs[:])
	if n < 1 {
		ctx, span := trace.StartSpan(ctx, "unknown")
		return span, ctx
	}
	fn := runtime.FuncForPC(pcs[0])
	name := fn.Name()
	if last := strings.LastIndexByte(name, '/'); last > 0 {
		name = name[last+1:]
	}
	c, span := trace.StartSpan(ctx, name)
	ctx = trace.NewContext(c, span)

	return span, ctx
}
