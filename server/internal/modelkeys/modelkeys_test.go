package modelkeys

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/netguard"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// fakeKey is the secret every test hunts for. It must never appear in a
// View, an error, a log line, a notice or the stored row.
const fakeKey = "sk-test-FAKE-KEY-do-not-leak-0123456789abcdef"

// leakMarkers are the substrings that would count as a leak: the whole
// key and a long-enough prefix that a provider's "sk-test-FAKE...cdef"
// echo would still be caught.
var leakMarkers = []string{fakeKey, fakeKey[:16], "do-not-leak"}

func assertNoLeak(t *testing.T, where, s string) {
	t.Helper()
	for _, m := range leakMarkers {
		if strings.Contains(s, m) {
			t.Errorf("%s leaks the key (%q): %s", where, m, s)
		}
	}
}

type harness struct {
	svc  *Service
	d    *db.DB
	user uuid.UUID
	logs *syncBuffer
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func newSealer(t *testing.T, secret string) *users.Sealer {
	t.Helper()
	s, err := users.NewSealer(strings.Repeat(secret, 40)[:40])
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newHarness(t *testing.T, guard netguard.Options) *harness {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	h := &harness{d: d, user: uuid.New(), logs: &syncBuffer{}}
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Ian', 1, 1)`, h.user.String()); err != nil {
		t.Fatal(err)
	}
	h.svc = New(d, newSealer(t, "a"), Options{
		Log:   slog.New(slog.NewTextHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
		Guard: guard,
	})
	return h
}

// provider is a fake OpenAI-compatible endpoint on an httptest TLS
// server. reply decides each answer; hits counts the requests that
// reached it.
type provider struct {
	srv   *httptest.Server
	hits  atomic.Int32
	mu    sync.Mutex
	auths []string
	model []string
	reply func(w http.ResponseWriter)
}

func newProvider(t *testing.T) *provider {
	t.Helper()
	p := &provider{reply: okReply}
	p.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.hits.Add(1)
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		p.mu.Lock()
		p.auths = append(p.auths, r.Header.Get("Authorization"))
		p.model = append(p.model, body.Model)
		reply := p.reply
		p.mu.Unlock()
		reply(w)
	}))
	t.Cleanup(p.srv.Close)
	return p
}

func okReply(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`)
}

// echoing answers status with an error that quotes the key back, the
// way a provider's "Incorrect API key provided: sk-..." does.
func echoing(status int, kind string) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, `{"error":{"type":%q,"message":"Incorrect API key provided: %s"}}`, kind, fakeKey)
	}
}

func (p *provider) set(reply func(w http.ResponseWriter)) {
	p.mu.Lock()
	p.reply = reply
	p.mu.Unlock()
}

// guardFor reaches p as https://example.com/v1 (the httptest
// certificate's name) while every other address keeps the production
// policy.
func guardFor(p *provider) netguard.Options {
	tlsCfg := p.srv.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	tlsCfg.MinVersion = tls.VersionTLS12
	port := p.srv.URL[strings.LastIndex(p.srv.URL, ":")+1:]
	return netguard.Options{
		Timeout:   5 * time.Second,
		TLSConfig: tlsCfg,
		Port:      port,
		Allow: func(ip netip.Addr) bool {
			return ip == netip.MustParseAddr("127.0.0.1") || netguard.AddrAllowed(ip)
		},
		Lookup: func(_ context.Context, host string) ([]netip.Addr, error) {
			if host == "example.com" {
				return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
			}
			return nil, errors.New("no such host")
		},
	}
}

func saveCustom(t *testing.T, h *harness, key string, games, calls int) View {
	t.Helper()
	v, err := h.svc.Save(context.Background(), h.user, Input{
		Provider: CustomProvider, BaseURL: "https://example.com/v1", Model: "my-model",
		Key: key, GamesPerDay: games, CallsPerGame: calls,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	return v
}

func seat(h *harness, game uuid.UUID, modelID string) *seatClient {
	return h.svc.SeatClient(aiseat.SeatSpec{KeyOwner: h.user, GameID: game, KeyModel: modelID}).(*seatClient)
}

func call(c model.Client) (model.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Complete(ctx, model.Request{Model: "server-default", User: "pick", MaxTokens: 16, Effort: "low", Thinking: "disabled"})
}

func TestNewIsOffWithoutASealerOrADatabase(t *testing.T) {
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	if s := New(d, nil, Options{}); s != nil {
		t.Error("no sealer: got a Service, want the feature off")
	}
	if s := New(nil, newSealer(t, "a"), Options{}); s != nil {
		t.Error("no database: got a Service, want the feature off")
	}
}

func TestSaveSealsTheKeyAndNeverReturnsIt(t *testing.T) {
	h := newHarness(t, netguard.Options{})
	v, err := h.svc.Save(context.Background(), h.user, Input{
		Provider: "anthropic", Model: "claude-haiku-4-5", Key: "  " + fakeKey + " ",
		GamesPerDay: 3, CallsPerGame: 50,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	raw, _ := json.Marshal(v)
	assertNoLeak(t, "the saved View", string(raw))
	if v.Last4 != "cdef" || v.Provider != "anthropic" || v.ProviderLabel != "Anthropic" || v.Model != "claude-haiku-4-5" {
		t.Errorf("View = %+v", v)
	}
	if v.BaseURL != "" {
		t.Errorf("a listed provider stores no address: %q", v.BaseURL)
	}
	var sealed []byte
	if err := h.d.QueryRow(`SELECT sealed_key FROM model_keys WHERE user_id = ?`, h.user.String()).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed, []byte(fakeKey)) || bytes.Contains(sealed, []byte("do-not-leak")) {
		t.Error("the stored key is not sealed")
	}
	if sealed[0] != 1 {
		t.Errorf("sealed value version byte = %d, want 1 (ADR 0051)", sealed[0])
	}
	// The row's whole text, as a backup would hold it.
	var dump string
	if err := h.d.QueryRow(`SELECT provider || base_url || model || last4 || COALESCE(last_error, '') FROM model_keys`).Scan(&dump); err != nil {
		t.Fatal(err)
	}
	assertNoLeak(t, "the stored row", dump)
	assertNoLeak(t, "the logs", h.logs.String())
}

func TestSaveWithoutAKeyKeepsTheSavedOneOnlyForTheSameProvider(t *testing.T) {
	h := newHarness(t, netguard.Options{})
	ctx := context.Background()
	if _, err := h.svc.Save(ctx, h.user, Input{Provider: "openai", Model: "gpt-x", GamesPerDay: 1, CallsPerGame: 1}); !errors.Is(err, ErrInvalid) {
		t.Errorf("no key and none saved: err = %v, want ErrInvalid", err)
	}
	first, err := h.svc.Save(ctx, h.user, Input{Provider: "openai", Model: "gpt-x", Key: fakeKey, GamesPerDay: 1, CallsPerGame: 1})
	if err != nil {
		t.Fatal(err)
	}
	v, err := h.svc.Save(ctx, h.user, Input{Provider: "openai", Model: "gpt-y", GamesPerDay: 4, CallsPerGame: 9})
	if err != nil {
		t.Fatalf("same provider, no key: %v", err)
	}
	if v.Model != "gpt-y" || v.GamesPerDay != 4 || v.CallsPerGame != 9 || v.Last4 != first.Last4 || !v.CreatedAt.Equal(first.CreatedAt) {
		t.Errorf("after an edit: %+v", v)
	}
	if _, err := h.svc.Save(ctx, h.user, Input{Provider: "groq", Model: "g", GamesPerDay: 1, CallsPerGame: 1}); !errors.Is(err, ErrInvalid) {
		t.Errorf("a new provider with no key: err = %v, want ErrInvalid", err)
	}
}

func TestSaveRefusesBadInput(t *testing.T) {
	h := newHarness(t, netguard.Options{})
	ok := Input{Provider: "openai", Model: "gpt-x", Key: fakeKey, GamesPerDay: 1, CallsPerGame: 1}
	for name, mut := range map[string]func(*Input){
		"unknown provider": func(in *Input) { in.Provider = "skynet" },
		"short key":        func(in *Input) { in.Key = "abc" },
		"key with space":   func(in *Input) { in.Key = "sk-abc def-ghij" },
		"key with newline": func(in *Input) { in.Key = "sk-abcdef\r\nX-Evil: 1" },
		"no model":         func(in *Input) { in.Model = " " },
		"odd model":        func(in *Input) { in.Model = "gpt x<script>" },
		"no games":         func(in *Input) { in.GamesPerDay = 0 },
		"many calls":       func(in *Input) { in.CallsPerGame = 5001 },
		"custom no url":    func(in *Input) { in.Provider = CustomProvider },
	} {
		in := ok
		mut(&in)
		_, err := h.svc.Save(context.Background(), h.user, in)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
			continue
		}
		assertNoLeak(t, name+" error", err.Error())
	}
}

// The save-time URL checks. They are for the message; the guard is the
// dial (TestACustomAddressThatResolvesPrivatelyIsRefusedAtDial).
func TestValidateCustomURL(t *testing.T) {
	for _, raw := range []string{
		"", "http://api.example.com/v1", "ftp://api.example.com", "https://", "api.example.com/v1",
		"https://user:pw@api.example.com/v1", "https://api.example.com:8443/v1", "https://api.example.com/v1?x=1",
		"https://api.example.com/v1#frag", "https://127.0.0.1/v1", "https://192.168.1.10/v1", "https://[::1]/v1",
		"https://10.0.0.5/v1", "https://169.254.169.254/latest", "https://100.64.0.1/v1", "https://[fd00::1]/v1",
		"https://8.8.8.8/v1", "https://localhost/v1", "https://ollama.local/v1", "https://box.internal/v1",
		"https://intranet/v1", "https://" + strings.Repeat("a", MaxURLLen) + ".com",
	} {
		if got, err := ValidateCustomURL(raw); !errors.Is(err, ErrInvalid) {
			t.Errorf("ValidateCustomURL(%q) = %q, %v; want ErrInvalid", raw, got, err)
		}
	}
	for raw, want := range map[string]string{
		"https://api.example.com/v1/":     "https://api.example.com/v1",
		" https://API.Example.com:443/v1": "https://api.example.com/v1",
		"https://llm.example.org":         "https://llm.example.org",
	} {
		if got, err := ValidateCustomURL(raw); err != nil || got != want {
			t.Errorf("ValidateCustomURL(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
}

func TestASeatCallsTheProviderWithTheOwnersKeyAndTheSeatsModel(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 5, 50)

	c := seat(h, uuid.New(), "seat-model")
	resp, err := call(c)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if resp.Text != "ok" {
		t.Errorf("reply = %q", resp.Text)
	}
	p.mu.Lock()
	auth, modelID := p.auths[0], p.model[0]
	p.mu.Unlock()
	if auth != "Bearer "+fakeKey {
		t.Error("the provider did not receive the owner's key")
	}
	if modelID != "seat-model" {
		t.Errorf("model sent = %q, want the seat's", modelID)
	}
	// A seat with no model of its own asks for the key's.
	if _, err := call(seat(h, uuid.New(), "")); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	if got := p.model[1]; got != "my-model" {
		t.Errorf("model sent = %q, want the key's default", got)
	}
	p.mu.Unlock()
	v, _, _ := h.svc.Get(context.Background(), h.user)
	if v.LastUsedAt == nil || v.GamesToday != 2 {
		t.Errorf("after two games: last used %v, games today %d", v.LastUsedAt, v.GamesToday)
	}
	assertNoLeak(t, "the logs", h.logs.String())
}

// The issue's hard rule, end to end: a refused key whose error quotes
// the key back ends the seat's model use, tells the owner, and the key
// appears nowhere this server writes.
func TestARefusedKeyStopsTheSeatTellsTheOwnerAndLeaksNothing(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reply  func(http.ResponseWriter)
		reason string
	}{
		{"401", echoing(http.StatusUnauthorized, "invalid_request_error"), ReasonRefused},
		{"403", echoing(http.StatusForbidden, "permission_error"), ReasonRefused},
		{"402", echoing(http.StatusPaymentRequired, "payment_required"), ReasonQuota},
		{"429 quota", echoing(http.StatusTooManyRequests, "insufficient_quota"), ReasonQuota},
		{"404 model", echoing(http.StatusNotFound, "model_not_found"), ReasonModel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProvider(t)
			p.set(tc.reply)
			h := newHarness(t, guardFor(p))
			saveCustom(t, h, fakeKey, 5, 50)
			c := seat(h, uuid.New(), "")

			_, err := call(c)
			if err == nil {
				t.Fatal("a refused call returned no error")
			}
			if errors.Is(err, model.ErrBudgetExceeded) {
				t.Error("the failing call itself must be reported as a failure, not a quiet skip")
			}
			assertNoLeak(t, "the error", err.Error())
			assertNoLeak(t, "the error (%+v)", fmt.Sprintf("%+v", err))

			notice := c.TakeNotice()
			if !strings.Contains(notice, tc.reason) || !strings.Contains(notice, "CMD & CTRL") {
				t.Errorf("notice = %q, want one naming %q", notice, tc.reason)
			}
			assertNoLeak(t, "the notice", notice)
			if again := c.TakeNotice(); again != "" {
				t.Errorf("the notice was posted twice: %q", again)
			}

			before := p.hits.Load()
			if _, err := call(c); !errors.Is(err, model.ErrBudgetExceeded) {
				t.Errorf("after the failure: err = %v, want ErrBudgetExceeded (the heuristic, quietly)", err)
			}
			if p.hits.Load() != before {
				t.Error("the seat called the provider again after its key failed")
			}

			v, _, _ := h.svc.Get(context.Background(), h.user)
			if v.LastError != tc.reason || v.LastErrorAt == nil {
				t.Errorf("the owner's Settings shows %q, want %q", v.LastError, tc.reason)
			}
			raw, _ := json.Marshal(v)
			assertNoLeak(t, "the View", string(raw))
			assertNoLeak(t, "the logs", h.logs.String())
		})
	}
}

// A passing failure (rate limit, a 500, a timeout) costs one window and
// the next call tries again.
func TestATransientFailureDoesNotStopTheSeat(t *testing.T) {
	p := newProvider(t)
	p.set(echoing(http.StatusTooManyRequests, "rate_limit_error"))
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 5, 50)
	c := seat(h, uuid.New(), "")
	_, err := call(c)
	var api *model.APIError
	if !errors.As(err, &api) || api.Status != http.StatusTooManyRequests {
		t.Fatalf("err = %v, want a 429 APIError", err)
	}
	assertNoLeak(t, "the error", err.Error())
	if n := c.TakeNotice(); n != "" {
		t.Errorf("a rate limit posted a notice: %q", n)
	}
	p.set(okReply)
	if _, err := call(c); err != nil {
		t.Errorf("the next call: %v", err)
	}
}

func TestTheCallsPerGameLimitStopsTheSeat(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 5, 2)
	c := seat(h, uuid.New(), "")
	for i := 0; i < 2; i++ {
		if _, err := call(c); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	if _, err := call(c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Errorf("third call: err = %v, want ErrBudgetExceeded", err)
	}
	if p.hits.Load() != 2 {
		t.Errorf("provider reached %d times, want 2", p.hits.Load())
	}
	if n := c.TakeNotice(); !strings.Contains(n, ReasonCallLimit) {
		t.Errorf("notice = %q", n)
	}
	if v, _, _ := h.svc.Get(context.Background(), h.user); v.LastError != "" {
		t.Errorf("a limit the owner set is not an error with the key: %q", v.LastError)
	}
}

func TestTheGamesPerDayLimitCountsEachGameOnce(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 1, 50)
	ctx := context.Background()
	if ok, err := h.svc.GamesLeftToday(ctx, h.user); err != nil || !ok {
		t.Fatalf("a fresh key: left %v, err %v", ok, err)
	}
	game := uuid.New()
	if _, err := call(seat(h, game, "")); err != nil {
		t.Fatal(err)
	}
	// The same game again — a second seat, or the seat after a deploy.
	if _, err := call(seat(h, game, "")); err != nil {
		t.Errorf("the same game again: %v", err)
	}
	if ok, _ := h.svc.GamesLeftToday(ctx, h.user); ok {
		t.Error("GamesLeftToday says yes after the day's only game")
	}
	other := seat(h, uuid.New(), "")
	if _, err := call(other); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Errorf("a second game: err = %v, want ErrBudgetExceeded", err)
	}
	if n := other.TakeNotice(); !strings.Contains(n, ReasonDailyLimit) {
		t.Errorf("notice = %q", n)
	}
	// Tomorrow the count starts again.
	h.svc.now = func() time.Time { return time.Now().UTC().Add(24 * time.Hour) }
	if _, err := call(seat(h, uuid.New(), "")); err != nil {
		t.Errorf("the next day: %v", err)
	}
}

func TestReplacingOrRemovingTheKeyReachesARunningSeat(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 5, 50)
	c := seat(h, uuid.New(), "")
	if _, err := call(c); err != nil {
		t.Fatal(err)
	}
	const second = "sk-second-key-0000000000000000"
	saveCustom(t, h, second, 5, 50)
	if _, err := call(c); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	got := p.auths[len(p.auths)-1]
	p.mu.Unlock()
	if got != "Bearer "+second {
		t.Error("the running seat kept using the replaced key")
	}
	if ok, err := h.svc.Remove(context.Background(), h.user); err != nil || !ok {
		t.Fatalf("Remove: %v %v", ok, err)
	}
	before := p.hits.Load()
	if _, err := call(c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Errorf("after removal: err = %v, want ErrBudgetExceeded", err)
	}
	if p.hits.Load() != before {
		t.Error("a removed key was used")
	}
	if n := c.TakeNotice(); !strings.Contains(n, ReasonRemoved) {
		t.Errorf("notice = %q", n)
	}
}

func TestAKeySealedUnderAnotherServerKeyIsUnreadable(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	saveCustom(t, h, fakeKey, 5, 50)
	h.svc.sealer = newSealer(t, "b")
	c := seat(h, uuid.New(), "")
	if _, err := call(c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Errorf("err = %v, want ErrBudgetExceeded", err)
	}
	if p.hits.Load() != 0 {
		t.Error("the provider was called without a key")
	}
	if v, _, _ := h.svc.Get(context.Background(), h.user); v.LastError != ReasonUnreadable {
		t.Errorf("last error = %q", v.LastError)
	}
	// A sealed key moved onto another person's row does not open either.
	other := uuid.New()
	if _, err := h.d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Bea', 1, 1)`, other.String()); err != nil {
		t.Fatal(err)
	}
	h.svc.sealer = newSealer(t, "a")
	if _, err := h.d.Exec(`INSERT INTO model_keys (user_id, provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game, created_at, updated_at)
		SELECT ?, provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game, created_at, updated_at FROM model_keys WHERE user_id = ?`,
		other.String(), h.user.String()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.svc.openKey(context.Background(), other); !errors.Is(err, errUnreadable) {
		t.Errorf("a copied sealed key opened for another person: %v", err)
	}
}

// SSRF through the feature: a custom address whose name resolves to a
// private or loopback address is refused at the dial, by the
// production policy, and the seat stops rather than retrying.
func TestACustomAddressThatResolvesPrivatelyIsRefusedAtDial(t *testing.T) {
	for _, addr := range []string{"127.0.0.1", "192.168.1.10", "10.1.2.3", "169.254.169.254", "100.64.1.1", "fd00::1", "::1", "0.0.0.0"} {
		t.Run(addr, func(t *testing.T) {
			p := newProvider(t)
			g := guardFor(p)
			g.Allow = nil // the production policy, no test exemption
			ip := netip.MustParseAddr(addr)
			g.Lookup = func(context.Context, string) ([]netip.Addr, error) { return []netip.Addr{ip}, nil }
			h := newHarness(t, g)
			saveCustom(t, h, fakeKey, 5, 50)

			res, err := h.svc.Test(context.Background(), h.user)
			if err != nil {
				t.Fatal(err)
			}
			if res.OK || res.Message != ReasonPrivate {
				t.Errorf("Test = %+v, want the private-address refusal", res)
			}
			c := seat(h, uuid.New(), "")
			if _, err := call(c); err == nil {
				t.Fatal("a private address answered")
			}
			if n := c.TakeNotice(); !strings.Contains(n, ReasonPrivate) {
				t.Errorf("notice = %q", n)
			}
			if p.hits.Load() != 0 {
				t.Errorf("the server behind the private address was reached %d times", p.hits.Load())
			}
			assertNoLeak(t, "the logs", h.logs.String())
			if strings.Contains(h.logs.String(), addr) {
				t.Errorf("the logs name the refused address %s", addr)
			}
		})
	}
}

func TestTestRecordsSuccessAndFailure(t *testing.T) {
	p := newProvider(t)
	h := newHarness(t, guardFor(p))
	ctx := context.Background()
	if _, err := h.svc.Test(ctx, h.user); !errors.Is(err, ErrNoKey) {
		t.Errorf("no key: err = %v", err)
	}
	saveCustom(t, h, fakeKey, 5, 50)
	p.set(echoing(http.StatusUnauthorized, "invalid_request_error"))
	res, err := h.svc.Test(ctx, h.user)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK || res.Message != ReasonRefused || res.View.LastError != ReasonRefused {
		t.Errorf("a refused key: %+v", res)
	}
	raw, _ := json.Marshal(res)
	assertNoLeak(t, "the test result", string(raw))
	p.set(okReply)
	res, err = h.svc.Test(ctx, h.user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.View.LastError != "" || res.View.LastUsedAt == nil {
		t.Errorf("a working key: %+v", res)
	}
	assertNoLeak(t, "the logs", h.logs.String())
}

func TestClassify(t *testing.T) {
	for _, tc := range []struct {
		err    error
		fatal  bool
		reason string
	}{
		{&model.APIError{Status: 401}, true, ReasonRefused},
		{&model.APIError{Status: 403}, true, ReasonRefused},
		{&model.APIError{Status: 402}, true, ReasonQuota},
		{&model.APIError{Status: 429, Kind: "insufficient_quota"}, true, ReasonQuota},
		{&model.APIError{Status: 400, Message: "Your credit balance is too low"}, true, ReasonQuota},
		{&model.APIError{Status: 404}, true, ReasonModel},
		{&model.APIError{Status: 302}, true, ReasonUnreachable},
		{&model.APIError{Status: 429, Kind: "rate_limit_error"}, false, ReasonRateLimited},
		{&model.APIError{Status: 500}, false, "The provider answered with an error (status 500)."},
		{context.DeadlineExceeded, false, ReasonTimeout},
		{fmt.Errorf("%w: dial: %v", model.ErrOutage, netguard.ErrAddressRefused), true, ReasonPrivate},
		{fmt.Errorf("%w: dns", model.ErrOutage), false, ReasonUnreachable},
		{errors.New("decode"), false, ReasonBadReply},
	} {
		fatal, reason := classify(tc.err)
		if fatal != tc.fatal || reason != tc.reason {
			t.Errorf("classify(%v) = %v %q, want %v %q", tc.err, fatal, reason, tc.fatal, tc.reason)
		}
	}
	if err := sanitize(&model.APIError{Status: 401, Kind: fakeKey, Message: fakeKey}, ReasonRefused); strings.Contains(err.Error(), "FAKE") {
		t.Errorf("sanitize kept the provider's text: %v", err)
	}
	if err := sanitize(context.DeadlineExceeded, ReasonTimeout); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("sanitize lost the deadline: %v", err)
	}
}

func TestEveryListedProviderIsHTTPSOn443(t *testing.T) {
	for _, p := range Providers() {
		if p.ID == CustomProvider {
			if p.BaseURL != "" {
				t.Errorf("custom has a fixed address %q", p.BaseURL)
			}
			continue
		}
		if _, err := ValidateCustomURL(p.BaseURL); err != nil {
			t.Errorf("%s: %s fails the custom-address checks: %v", p.ID, p.BaseURL, err)
		}
	}
	if _, ok := LookupProvider(" OpenAI "); !ok {
		t.Error("LookupProvider is not forgiving about case and spaces")
	}
}
