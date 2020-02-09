package yawhg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/MarcvanMelle/yawhg"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type requestIDTestCase struct {
	name              string
	request           *http.Request
	headers           map[string]string
	context           context.Context
	expectedRequestID string
}

var requestidTestCases = []requestIDTestCase{
	func() requestIDTestCase {
		requestID := "my-request-id"

		return requestIDTestCase{
			name:              "extract_request_id_from_context_if_it_exists",
			request:           httptest.NewRequest("GET", "/", nil),
			context:           yawhg.AddToContext(context.Background(), requestID),
			expectedRequestID: requestID,
		}
	}(),
	func() requestIDTestCase {
		return requestIDTestCase{
			name:    "add_new_request_id_to_context_if_it_does_not_exist",
			request: httptest.NewRequest("GET", "/", nil),
			context: context.Background(),
		}
	}(),
	func() requestIDTestCase {
		requestID := "my-request-id"

		return requestIDTestCase{
			name:              "extract_request_id_from_incoming_context_if_it_exists",
			request:           httptest.NewRequest("GET", "/", nil),
			context:           metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"x-request-id": requestID})),
			expectedRequestID: requestID,
		}
	}(),
}

func TestRequestID(t *testing.T) {
	for _, testCase := range requestidTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.context != nil && testCase.expectedRequestID != "" {
				_, reqID := yawhg.FromContext(testCase.context)
				assert.Equal(t, testCase.expectedRequestID, reqID)

				reqCtx, req := yawhg.AddToHeader(testCase.request, testCase.expectedRequestID)
				id := yawhg.FromHeader(req)
				assert.Equal(t, testCase.expectedRequestID, id)

				_, headerID := yawhg.FromContext(reqCtx)
				assert.Equal(t, testCase.expectedRequestID, headerID)
			} else {
				// generate and attach new request ID to the outgoing context
				// check that the returned request ID is well-formed
				outgoingCtx, reqID := yawhg.FromContext(context.Background())
				assert.Regexp(t, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"), reqID)

				// double-check that the request ID is actually present on the context
				_, attachedReqID := yawhg.FromContext(outgoingCtx)
				assert.Regexp(t, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"), attachedReqID)
				assert.Equal(t, reqID, attachedReqID)
			}
		})
	}
}
