package yawhg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MarcvanMelle/yawhg"
)

const testRequestID string = "3ad746c8-45b8-4b27-ac81-2f6d24b83076"

type middlewareTestCase struct {
	name      string
	request   *http.Request
	requestID string
	context   context.Context
}

var middlewareTestCases = []middlewareTestCase{
	middlewareTestCase{
		name:    "generate_and_attach_x_request_id_header_when_it_does_not_exist",
		request: httptest.NewRequest("GET", "/", nil),
	},
	middlewareTestCase{
		name:      "forward_x_request_id_header_when_it_does_exist",
		request:   httptest.NewRequest("GET", "/", nil),
		requestID: testRequestID,
	},
}

func TestHTTPTraceMiddleware(t *testing.T) {
	for _, testCase := range middlewareTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if testCase.requestID != "" { // request headers match the headers passed in even after processing by middleware
					assert.Equal(t, testCase.requestID, r.Header.Get(yawhg.RequestIDHeader))
					return
				}

				// if no request ID passed in, expect that new request ID has been generated
				assert.Regexp(t, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"), r.Header.Get(yawhg.RequestIDHeader))
			})

			handler := yawhg.AddMiddleware(
				testHandler,               // executed last, perform assertions
				yawhg.HTTPTraceMiddleware, // set or generate request ID before logging
			)

			req := testCase.request

			if testCase.requestID != "" {
				req.Header.Set(yawhg.RequestIDHeader, testCase.requestID)
			}

			handler.ServeHTTP(w, req)
		})
	}
}
