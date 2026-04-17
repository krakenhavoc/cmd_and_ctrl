package deck

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// withMoxfieldStub swaps moxfieldAPIHost to a test server for the
// duration of the test. t.Cleanup restores the original host so
// parallel test files don't see a leaked override.
func withMoxfieldStub(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	orig := moxfieldAPIHost
	moxfieldAPIHost = srv.URL
	t.Cleanup(func() {
		moxfieldAPIHost = orig
		srv.Close()
	})
	return srv
}

func TestFetchFromURLRejectsNonHTTP(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"ftp://moxfield.com/decks/abc",
		"not a url",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, _, err := FetchFromURL(context.Background(), DefaultClient(), u)
			if !errors.Is(err, ErrUnknownSource) {
				t.Errorf("got %v, want ErrUnknownSource", err)
			}
		})
	}
}

func TestFetchFromURLUnknownHost(t *testing.T) {
	_, _, err := FetchFromURL(context.Background(), DefaultClient(), "https://tappedout.net/mtg-decks/abc")
	if !errors.Is(err, ErrUnknownSource) {
		t.Errorf("got %v, want ErrUnknownSource", err)
	}
}

func TestFetchFromURLMoxfieldHappyPath(t *testing.T) {
	body := `{
		"name": "Atraxa Superfriends",
		"commanders": { "Atraxa, Praetors' Voice": {"quantity": 1} },
		"mainboard":  { "Sol Ring": {"quantity": 1}, "Forest": {"quantity": 2} }
	}`
	srv := withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v3/decks/all/abc123" {
			t.Errorf("path: got %q, want /v3/decks/all/abc123", got)
		}
		if got := r.Header.Get("User-Agent"); got != userAgent {
			t.Errorf("User-Agent: got %q, want %q", got, userAgent)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	_ = srv

	for _, u := range []string{
		"https://moxfield.com/decks/abc123",
		"https://moxfield.com/decks/abc123/atraxa-superfriends",
		"https://www.moxfield.com/decks/abc123",
	} {
		t.Run(u, func(t *testing.T) {
			name, entries, err := FetchFromURL(context.Background(), DefaultClient(), u)
			if err != nil {
				t.Fatalf("FetchFromURL: %v", err)
			}
			if name != "Atraxa Superfriends" {
				t.Errorf("name: got %q", name)
			}
			if len(entries) != 3 {
				t.Errorf("entries: got %d, want 3; %+v", len(entries), entries)
			}
		})
	}
}

func TestFetchFromURLMoxfieldMalformedPath(t *testing.T) {
	cases := []string{
		"https://moxfield.com/",
		"https://moxfield.com/decks/",
		"https://moxfield.com/users/alice",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, _, err := FetchFromURL(context.Background(), DefaultClient(), u)
			if !errors.Is(err, ErrUnknownSource) {
				t.Errorf("got %v, want ErrUnknownSource", err)
			}
		})
	}
}

func TestFetchFromURLMoxfieldNotFound(t *testing.T) {
	withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	_, _, err := FetchFromURL(context.Background(), DefaultClient(), "https://moxfield.com/decks/missing")
	if !errors.Is(err, ErrDeckNotFound) {
		t.Errorf("got %v, want ErrDeckNotFound", err)
	}
}

func TestFetchFromURLMoxfieldPrivate(t *testing.T) {
	withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "auth required", http.StatusUnauthorized)
	}))
	_, _, err := FetchFromURL(context.Background(), DefaultClient(), "https://moxfield.com/decks/private")
	if !errors.Is(err, ErrDeckPrivate) {
		t.Errorf("got %v, want ErrDeckPrivate", err)
	}

	// 403 Forbidden is also treated as private, not as an
	// availability issue.
	withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	_, _, err = FetchFromURL(context.Background(), DefaultClient(), "https://moxfield.com/decks/hidden")
	if !errors.Is(err, ErrDeckPrivate) {
		t.Errorf("got %v, want ErrDeckPrivate", err)
	}
}

func TestFetchFromURLMoxfieldUpstreamDown(t *testing.T) {
	withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "sorry", http.StatusInternalServerError)
	}))
	_, _, err := FetchFromURL(context.Background(), DefaultClient(), "https://moxfield.com/decks/any")
	if !errors.Is(err, ErrExternalAPIUnavailable) {
		t.Errorf("got %v, want ErrExternalAPIUnavailable", err)
	}
}

func TestFetchFromURLContextCancelled(t *testing.T) {
	// Handler hangs. Cancelling the context must propagate through
	// http.Client.Do and surface as ErrExternalAPIUnavailable (the
	// fetch never produced a response).
	withMoxfieldStub(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancelled; request aborts immediately
	_, _, err := FetchFromURL(ctx, DefaultClient(), "https://moxfield.com/decks/slow")
	if !errors.Is(err, ErrExternalAPIUnavailable) {
		t.Errorf("got %v, want ErrExternalAPIUnavailable", err)
	}
}
