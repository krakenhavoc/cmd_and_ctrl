package cards

import (
	"context"
	"net/http"
	"time"
)

// contextWithTimeout returns a context derived from r.Context() with
// the given maximum duration. Extracted so handler code can stay
// short; also makes the timeout duration discoverable in one place.
func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
