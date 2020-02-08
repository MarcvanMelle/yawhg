package yawhg

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/metadata"
)

// AddToContext attaches the request ID header to context metadata as a key/value pair
func AddToContext(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, RequestIDHeader, requestID)
}

// AddToHeader is a helper function that adds a request ID to the x-request-id http header
func AddToHeader(r *http.Request, requestID string) (context.Context, *http.Request) {
	// https://godoc.org/net/http#Handler - "Except for reading the body, handlers should not modify the provided Request."
	// We create a shallow copy of the request, update the copy, and return that
	r2 := new(http.Request)
	*r2 = *r
	r2.Header.Set(RequestIDHeader, requestID) // add or replace request id header

	ctx := r2.Context()
	return AddToContext(ctx, requestID), r2
}

// FromContext retrieves the request id from the context if it exists.
// It will generate a new request id if the context has none, and append it to the context
func FromContext(ctx context.Context) (context.Context, string) {
	// check if request ID is stored in the incomingKey
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md.Get(RequestIDHeader)) > 0 {
		return ctx, md.Get(RequestIDHeader)[0]
	}

	// check if request ID is stored in the outgoingKey
	md, ok = metadata.FromOutgoingContext(ctx)
	if ok && len(md.Get(RequestIDHeader)) > 0 {
		return ctx, md.Get(RequestIDHeader)[0]
	}

	// no request ID found, append to context metadata and return the new context
	newUUID, _ := uuid.NewV4()

	requestID := newUUID.String() // generate new request ID if not present
	ctx = AddToContext(ctx, requestID)
	return ctx, requestID
}

// FromHeader retrieves the value of the x-request-id header
func FromHeader(req *http.Request) string {
	return req.Header.Get(RequestIDHeader)
}
