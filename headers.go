package server

import (
	"fmt"
	"net/http"
	"time"
)

const daysToSeconds = 86400

// AddCacheHeaders adds Cache-Control, Expires and Etag headers
// with a default age of 30 days
func AddCacheHeaders(w http.ResponseWriter, days int, hash string) {
	// Cache for 30 days
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d", days*daysToSeconds))

	// Set an expires header Mon Jan 2 15:04:05 -0700 MST 2006
	w.Header().Set("Expires", time.Now().AddDate(0, 0, days).UTC().Format("Mon, 2 Jan 2006 15:04:05 MST"))

	// For etag Just hash the path - static resources are assumed to have a fingerprint
	w.Header().Set("ETag", fmt.Sprintf("\"%s\"", hash))
}
