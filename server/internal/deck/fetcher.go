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
	// 403 *with a JSON body*, i.e. the deck exists but is not publicly
	// readable. We don't support authenticated fetches at S06.5.
	ErrDeckPrivate = errors.New("deck: source requires authentication")

	// ErrUpstreamBlocked is returned when the upstream's CDN (most
	// commonly Cloudflare) returns a 403 with an HTML bot-wall page
	// rather than a JSON "deck private" response. The deck itself may
	// be public; the server's egress IP is being rate-limited or
	// reputation-scored by the CDN. Distinguished from ErrDeckPrivate
	// so the error message can steer users toward the right fix (use
	// Archidekt or paste the text export) instead of telling them to
	// change Moxfield's privacy setting when that's not the problem.
	ErrUpstreamBlocked = errors.New("deck: source CDN blocked this server's IP")

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

// FetchViolation classifies a FetchFromURL error into the Violation
// shape the HTTP layer emits on 422. Returns (violation, true) if
// the error maps to one of our structured sentinels; (zero, false)
// otherwise — callers should fall through to a generic 400/500 in
// that case.
//
// The URL is echoed in the `card` field so the client's violations-
// list renderer can show what the user actually pasted.
func FetchViolation(url string, err error) (Violation, bool) {
	switch {
	case errors.Is(err, ErrUnknownSource):
		return Violation{
			Code:    CodeUnknownSource,
			Card:    url,
			Message: err.Error(),
		}, true
	case errors.Is(err, ErrDeckNotFound):
		return Violation{
			Code:    CodeDeckNotFound,
			Card:    url,
			Message: err.Error(),
		}, true
	case errors.Is(err, ErrDeckPrivate):
		return Violation{
			Code:    CodeDeckPrivate,
			Card:    url,
			Message: err.Error(),
		}, true
	case errors.Is(err, ErrUpstreamBlocked):
		return Violation{
			Code:    CodeUpstreamBlocked,
			Card:    url,
			Message: err.Error(),
		}, true
	case errors.Is(err, ErrExternalAPIUnavailable):
		return Violation{
			Code:    CodeExternalAPIUnavailable,
			Card:    url,
			Message: err.Error(),
		}, true
	}
	return Violation{}, false
}

// looksLikeCloudflareBlock returns true when a 4xx response carries
// a Cloudflare bot-wall HTML page rather than a real upstream JSON
// body. Callers use this to re-classify a 403 from ErrDeckPrivate
// to ErrUpstreamBlocked so the error surface to the user names the
// actual problem.
//
// Detection is conservative: we only rewrite the error when the
// response both claims text/html and peeks as a Cloudflare
// challenge page. An empty body or an application/json 403 stays
// as ErrDeckPrivate so genuine private-deck responses aren't
// misreported.
//
// The peek uses a tiny buffer (1 KiB) and replaces resp.Body so
// later reads of the body still work if a caller wants the raw
// response — though none currently do on the error path.
func looksLikeCloudflareBlock(resp *http.Response) bool {
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(ct), "text/html") {
		return false
	}
	const peekBytes = 1024
	peek, err := io.ReadAll(io.LimitReader(resp.Body, peekBytes))
	if err != nil {
		return false
	}
	// Rewind so any later reader still sees the peeked bytes
	// followed by the rest of the body. Matters less on the error
	// path but keeps the contract clean.
	resp.Body = &peekedBody{peek: peek, rest: resp.Body}
	needle := strings.ToLower(string(peek))
	return strings.Contains(needle, "cloudflare") ||
		strings.Contains(needle, "attention required") ||
		strings.Contains(needle, "cf-ray")
}

// peekedBody stitches the already-read peek buffer back in front
// of the remaining stream so resp.Body stays usable after peek.
type peekedBody struct {
	peek []byte
	rest io.ReadCloser
}

func (p *peekedBody) Read(b []byte) (int, error) {
	if len(p.peek) > 0 {
		n := copy(b, p.peek)
		p.peek = p.peek[n:]
		return n, nil
	}
	return p.rest.Read(b)
}

func (p *peekedBody) Close() error { return p.rest.Close() }

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
