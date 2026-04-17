package deck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// External-fetch errors. All structured so the HTTP layer can map
// them to the 422 violations[] shape that existing deck-upload
// callers already handle, without string-matching the upstream's
// failure mode.
var (
	// ErrDeckNotFound is returned when the upstream API responds 404
	// (deck doesn't exist or was deleted).
	ErrDeckNotFound = errors.New("deck: not found at source")

	// ErrDeckPrivate is returned when the upstream API responds 401 /
	// 403 (deck exists but is not publicly readable). We don't support
	// authenticated fetches at S06.5.
	ErrDeckPrivate = errors.New("deck: source requires authentication")

	// ErrExternalAPIUnavailable is returned for upstream 5xx,
	// connection failures, or timeouts.
	ErrExternalAPIUnavailable = errors.New("deck: source API unavailable")

	// ErrUnknownSource is returned when a URL doesn't match any
	// supported host. Callers should check for it before treating a
	// fetch failure as an upstream problem.
	ErrUnknownSource = errors.New("deck: URL is not from a supported deck source")
)

// defaultHTTPTimeout caps outbound requests to the deck-builder APIs.
// Deck-builder endpoints typically respond in well under 2 s; 10 s is
// generous headroom for slow TLS handshakes on cold connections.
const defaultHTTPTimeout = 10 * time.Second

// userAgent identifies our outbound requests to upstream services.
// Descriptive so a source admin who blocks us can find the project
// and open an issue rather than just silently rate-limiting us.
const userAgent = "cmd_and_ctrl/0.1 (+https://github.com/krakenhavoc/cmd_and_ctrl)"

// DefaultClient returns an *http.Client suitable for deck-source
// requests. Reusable across fetches so the TCP / TLS connection pool
// stays warm between calls.
func DefaultClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}

// FetchFromURL resolves a deck URL to a (name, entries) pair by
// dispatching to the right upstream API based on hostname.
//
// Returns a structured error from the exported sentinels above
// (ErrDeckNotFound / ErrDeckPrivate / ErrExternalAPIUnavailable /
// ErrUnknownSource) so HTTP callers can render them as typed
// violations without parsing message text.
//
// Caller supplies the http.Client so the same pool is reused across
// requests and so tests can inject a stubbed transport.
func FetchFromURL(ctx context.Context, client *http.Client, rawURL string) (name string, entries []Entry, err error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", nil, fmt.Errorf("%w: empty URL", ErrUnknownSource)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrUnknownSource, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", nil, fmt.Errorf("%w: scheme %q", ErrUnknownSource, u.Scheme)
	}
	host := strings.ToLower(u.Host)
	// Strip a leading "www." / "api." so links copied from either the
	// site UI or the API host resolve the same way.
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "api.")
	host = strings.TrimPrefix(host, "api2.")

	switch host {
	case "moxfield.com":
		return fetchMoxfield(ctx, client, u)
	case "archidekt.com":
		return fetchArchidekt(ctx, client, u)
	}
	return "", nil, fmt.Errorf("%w: host %q", ErrUnknownSource, u.Host)
}

// readLimitedBody reads up to maxBody bytes from r and returns the
// bytes. Guards against a hostile or misbehaving upstream serving a
// gigabyte response — deck-builder JSONs are typically under 200 KiB.
func readLimitedBody(r io.Reader, maxBody int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, maxBody))
}

// maxBodyBytes caps the bytes we read from an upstream response. 2
// MiB matches the inbound deck-upload body cap — a deck that requires
// more than that is malformed on the upstream's end, not ours.
const maxBodyBytes = 2 * 1024 * 1024

// classifyHTTPStatus maps an upstream HTTP status into one of our
// structured sentinel errors. Used by every source-specific fetcher
// so the mapping stays consistent.
func classifyHTTPStatus(status int) error {
	switch {
	case status == http.StatusNotFound:
		return ErrDeckNotFound
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return ErrDeckPrivate
	case status >= 500:
		return ErrExternalAPIUnavailable
	case status >= 400:
		// 400 / 422 / 429 — treat as API-level failure; the deck URL
		// is well-formed by the time we get here, so any 4xx that
		// isn't 401/403/404 is an upstream contract change.
		return fmt.Errorf("%w: upstream returned %d", ErrExternalAPIUnavailable, status)
	}
	return nil
}
