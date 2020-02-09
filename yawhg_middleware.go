package yawhg

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
)

// AddMiddleware adds middleware to a Handler
func AddMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

// GRPCLogInterceptor logs server side incoming requests and responses
func GRPCLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	t := time.Now()

	_, requestID := FromContext(ctx)

	InfoWithTracing(
		ctx,
		Fields{
			"Method":    info.FullMethod,
			"Request":   req,
			"RequestID": requestID,
		},
	)

	resp, err := handler(ctx, req)

	payload := Fields{
		"Method":       info.FullMethod,
		"Response":     resp,
		"RequestID":    requestID,
		"ResponseTime": time.Since(t).Seconds(),
	}

	if err != nil {
		payload["Error"] = err.Error()
	}

	InfoWithTracing(ctx, payload)

	return resp, err
}

// GRPCTraceInterceptor adds x-request-id to incoming context if not present
func GRPCTraceInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestIDCtx, _ := FromContext(ctx)
	resp, err := handler(requestIDCtx, req)
	return resp, err
}

// HTTPLogMiddleware logs the incoming request to the http server
func HTTPLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Buffer{}
		buf.ReadFrom(r.Body)

		InfoWithTracing(r.Context(), Fields{
			"Method":       r.Method,
			"RequestPath":  r.URL.Path,
			"RequestQuery": r.URL.RawQuery,
			"RequestBody":  buf.String(),
		})

		next.ServeHTTP(w, r)
	})
}

// HTTPTraceMiddleware adds a x-request-id to the http header, and request context if not present
func HTTPTraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var newUUID uuid.UUID

		reqID := r.Header.Get(RequestIDHeader)
		if l := len(reqID); l == 0 {
			newUUID, err = uuid.NewV4()
			if err != nil {
				WithTracing(r.Context(), Fields{}, err).Error("failed to generate new request ID")
			}

			reqID = newUUID.String() // generate new request ID if not present
		} else if l > 36 {
			reqID = reqID[:36] // truncate if longer than UUID
		}

		if err != nil {
			next.ServeHTTP(w, r)
		} else {
			ctx, updatedReq := AddToHeader(r, reqID)
			next.ServeHTTP(w, updatedReq.WithContext(ctx))
		}
	})
}
