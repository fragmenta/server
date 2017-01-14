package log

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"
)

// RequestID is but a simple token for tracing requests.
type RequestID struct {
	id []byte
}

// String returns a string formatting for the request id.
func (r *RequestID) String() string {
	return fmt.Sprintf("%X-%X", r.id[0:4], r.id[4:8])
}

// NewRequestID returns a new random request id.
func newRequestID() *RequestID {
	r := &RequestID{
		id: make([]byte, 8),
	}
	rand.Read(r.id)
	return r
}

type ctxKey struct{}

// Trace retreives the request id from a request as a string.
func Trace(r *http.Request) string {
	rid, ok := r.Context().Value(&ctxKey{}).(*RequestID)
	if ok {
		return rid.String()
	}
	return ""
}

// GetRequestID retreives the request id from a request.
func GetRequestID(r *http.Request) *RequestID {
	return r.Context().Value(&ctxKey{}).(*RequestID)
}

// SetRequestID saves the request id in the request context.
func SetRequestID(r *http.Request, rid *RequestID) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, &ctxKey{}, rid)
	return r.WithContext(ctx)
}

// Middleware adds a logging wrapper and request tracing to requests.
func Middleware(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		r = SetRequestID(r, requestID) // Sets on context for handlers
		Info(Values{"msg": "<- Request", "method": r.Method, "url": r.RequestURI, "len": r.ContentLength, "ip": r.RemoteAddr, "trace": requestID.String()})
		start := time.Now()
		h(w, r)
		Time(start, Values{"msg": "-> Response", "url": r.RequestURI, "trace": requestID.String()})
	}

}
