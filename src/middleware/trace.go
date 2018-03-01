package middleware

import (
	"context"
	"net/http"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

type TraceMiddlewareConfig struct {
	Service        string
	JaegerEndpoint string
}

type TraceMiddleware struct {
	*middleware

	exporter *jaeger.Exporter
}

func NewTraceMiddleware(config *TraceMiddlewareConfig, ignore []string) (*TraceMiddleware, error) {
	var (
		err error
		mw  = TraceMiddleware{
			middleware: newMiddleware(ignore),
		}
	)

	if mw.exporter, err = jaeger.NewExporter(jaeger.Options{
		Endpoint:    config.JaegerEndpoint,
		ServiceName: config.Service,
	}); err != nil {

	}

	trace.RegisterExporter(mw.exporter)

	// TODO: add configurable sampling
	trace.SetDefaultSampler(trace.AlwaysSample())

	return &mw, nil
}

func (mw *TraceMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var (
		traceCtx context.Context
		span     *trace.Span
	)

	if !mw.isIgnored(r) {
		traceCtx, span = trace.StartSpan(r.Context(), r.URL.Path)
		next(rw, r.WithContext(traceCtx))
	} else {
		next(rw, r)
	}

	if !mw.isIgnored(r) {
		// TODO: configurable tags
		span.SetAttributes(
			trace.StringAttribute{Key: "method", Value: r.Method},
		)

		span.End()

		mw.exporter.Flush()
	}
}
