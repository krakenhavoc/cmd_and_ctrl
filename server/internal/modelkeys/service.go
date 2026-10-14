package modelkeys

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/netguard"
)

// Sealer is ADR 0051's AES-256-GCM sealer. *users.Sealer satisfies it.
type Sealer interface {
	Seal(plaintext, ad []byte) ([]byte, error)
	Open(sealed, ad []byte) ([]byte, error)
}

// ErrNoKey is a request for a key the person has not saved.
var ErrNoKey = errors.New("modelkeys: no key saved")

// Request limits for the provider calls.
const (
	// RequestTimeout bounds one call end to end. A seat's own deadline
	// (the tier's think time, a few seconds) is shorter and wins; this
	// is the ceiling for the Settings test.
	RequestTimeout = 15 * time.Second
	// MaxResponseBytes caps a provider's reply. A move pick is a few
	// hundred bytes and an improvisation a few KiB.
	MaxResponseBytes = 256 << 10
)

// Service is the feature. Nil means it is off on this server (no
// sealing key, or no database), and every caller treats it so.
type Service struct {
	store  *sqlStore
	sealer Sealer
	http   *http.Client
	log    *slog.Logger
	now    func() time.Time

	mu sync.Mutex
	// gen counts changes to a person's key (save, remove, account
	// deletion). A seat's client compares it before each call, so a
	// replaced key is used from the next call and a removed one is not
	// used at all, without restarting the seat.
	gen map[uuid.UUID]uint64
	// counted is the games already counted today, keyed day|user|game,
	// so a seat restarted by a deploy does not count its game twice.
	counted map[string]bool
}

// Options configure New.
type Options struct {
	Log *slog.Logger
	// Guard replaces the netguard options. TEST SEAM: production leaves
	// it zero, which is the strict policy with this package's timeout
	// and response cap.
	Guard netguard.Options
}

// New builds the Service, or returns nil when it cannot be offered
// safely: no database to keep keys in, or no sealer to seal them with.
// ADR 0182 §4: no sealing key means the feature is off, never plaintext.
func New(d *db.DB, sealer Sealer, opt Options) *Service {
	if d == nil || sealer == nil {
		return nil
	}
	g := opt.Guard
	if g.Timeout == 0 {
		g.Timeout = RequestTimeout
	}
	if g.MaxResponseBytes == 0 {
		g.MaxResponseBytes = MaxResponseBytes
	}
	log := opt.Log
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		store:   newSQLStore(d),
		sealer:  sealer,
		http:    netguard.NewClient(g),
		log:     log,
		now:     func() time.Time { return time.Now().UTC() },
		gen:     map[uuid.UUID]uint64{},
		counted: map[string]bool{},
	}
}

// View is a saved key as anyone outside this package may see it: never
// the key, only enough to say which key it is.
type View struct {
	Provider      string     `json:"provider"`
	ProviderLabel string     `json:"provider_label"`
	BaseURL       string     `json:"base_url,omitempty"`
	Model         string     `json:"model"`
	Last4         string     `json:"last4"`
	GamesPerDay   int        `json:"games_per_day"`
	CallsPerGame  int        `json:"calls_per_game"`
	GamesToday    int        `json:"games_today"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	LastErrorAt   *time.Time `json:"last_error_at,omitempty"`
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (s *Service) day() string { return s.now().Format(time.DateOnly) }

func (s *Service) viewOf(r record) View {
	p, _ := LookupProvider(r.Provider)
	label := p.Label
	if label == "" {
		label = r.Provider
	}
	v := View{
		Provider: r.Provider, ProviderLabel: label, BaseURL: r.BaseURL, Model: r.Model, Last4: r.Last4,
		GamesPerDay: r.GamesPerDay, CallsPerGame: r.CallsPerGame,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		LastUsedAt: timePtr(r.LastUsedAt), LastError: r.LastError, LastErrorAt: timePtr(r.LastErrorAt),
	}
	if r.Day == s.day() {
		v.GamesToday = r.DayGames
	}
	return v
}

// Get returns the person's saved key as a View.
func (s *Service) Get(ctx context.Context, user uuid.UUID) (View, bool, error) {
	r, ok, err := s.store.get(ctx, user)
	if err != nil || !ok {
		return View{}, ok, err
	}
	return s.viewOf(r), true, nil
}

// Input is a save from the Settings page.
type Input struct {
	Provider string
	// BaseURL is the "custom" provider's address; ignored for the rest.
	BaseURL string
	Model   string
	// Key is the API key. Empty keeps the saved one, which is only
	// allowed when the provider and address are unchanged: a key is
	// sealed to its provider, and a key for one provider is no use to
	// another.
	Key          string
	GamesPerDay  int
	CallsPerGame int
}

// keyAD binds a sealed key to its owner and provider, so a sealed value
// copied onto another person's row, or kept across a provider change,
// does not open. Distinct from the refresh tokens' associated data.
func keyAD(user uuid.UUID, provider string) []byte {
	return []byte("cmd_and_ctrl model key v1\x00" + user.String() + "\x00" + provider)
}

// Save validates in, seals the key and stores it.
func (s *Service) Save(ctx context.Context, user uuid.UUID, in Input) (View, error) {
	p, ok := LookupProvider(in.Provider)
	if !ok {
		return View{}, invalid("choose a provider from the list")
	}
	base := ""
	if p.ID == CustomProvider {
		var err error
		if base, err = ValidateCustomURL(in.BaseURL); err != nil {
			return View{}, err
		}
	}
	modelID, err := ValidateModel(in.Model)
	if err != nil {
		return View{}, err
	}
	if err := ValidateLimits(in.GamesPerDay, in.CallsPerGame); err != nil {
		return View{}, err
	}
	old, had, err := s.store.get(ctx, user)
	if err != nil {
		return View{}, err
	}
	now := s.now()
	rec := record{
		UserID: user, Provider: p.ID, BaseURL: base, Model: modelID,
		GamesPerDay: in.GamesPerDay, CallsPerGame: in.CallsPerGame,
		CreatedAt: now, UpdatedAt: now,
	}
	if had {
		rec.CreatedAt = old.CreatedAt
	}
	if in.Key == "" {
		if !had || old.Provider != p.ID || old.BaseURL != base {
			return View{}, invalid("paste the API key for this provider")
		}
		rec.Sealed, rec.Last4 = old.Sealed, old.Last4
	} else {
		key, err := ValidateKey(in.Key)
		if err != nil {
			return View{}, err
		}
		sealed, err := s.sealer.Seal([]byte(key), keyAD(user, p.ID))
		if err != nil {
			return View{}, errors.New("modelkeys: sealing the key failed")
		}
		rec.Sealed, rec.Last4 = sealed, last4(key)
	}
	if err := s.store.put(ctx, rec); err != nil {
		return View{}, err
	}
	s.bump(user)
	saved, _, err := s.store.get(ctx, user)
	if err != nil {
		return View{}, err
	}
	return s.viewOf(saved), nil
}

// Remove deletes the person's key. Seats already running on it stop
// using it from their next call.
func (s *Service) Remove(ctx context.Context, user uuid.UUID) (bool, error) {
	ok, err := s.store.remove(ctx, user)
	s.bump(user)
	return ok, err
}

// Forget is told when a person's account is deleted (the row went in
// the deletion's own transaction): their running seats stop using the
// key they had already opened.
func (s *Service) Forget(user uuid.UUID) { s.bump(user) }

func (s *Service) bump(user uuid.UUID) {
	s.mu.Lock()
	s.gen[user]++
	s.mu.Unlock()
}

func (s *Service) generation(user uuid.UUID) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gen[user]
}

// HasKey reports whether the person has a saved key, and its model.
func (s *Service) HasKey(ctx context.Context, user uuid.UUID) (View, bool) {
	v, ok, err := s.Get(ctx, user)
	if err != nil {
		s.log.Warn("reading a model key failed", "err", err)
		return View{}, false
	}
	return v, ok
}

// GamesLeftToday reports whether the key's games-a-day limit has room
// for one more game today. The lobby asks before seating a bot on the
// key, so a full day is a clear refusal at the picker rather than a
// seat that falls back at its first move.
func (s *Service) GamesLeftToday(ctx context.Context, user uuid.UUID) (bool, error) {
	used, limit, ok, err := s.store.gamesToday(ctx, user, s.day())
	if err != nil {
		return false, err
	}
	if !ok {
		return false, ErrNoKey
	}
	return used < limit, nil
}

// claimGame counts game against the owner's daily limit once per day.
func (s *Service) claimGame(ctx context.Context, user, game uuid.UUID) (bool, error) {
	day := s.day()
	k := day + "|" + user.String() + "|" + game.String()
	s.mu.Lock()
	done := s.counted[k]
	s.mu.Unlock()
	if done && game != uuid.Nil {
		return true, nil
	}
	ok, err := s.store.claimGame(ctx, user, day)
	if err != nil || !ok {
		return ok, err
	}
	s.mu.Lock()
	// Yesterday's entries are no use; keep the map to today.
	for old := range s.counted {
		if len(old) < len(day) || old[:len(day)] != day {
			delete(s.counted, old)
		}
	}
	s.counted[k] = true
	s.mu.Unlock()
	return true, nil
}

// openKey loads and opens the person's key. The plaintext goes straight
// into a client and nowhere else.
func (s *Service) openKey(ctx context.Context, user uuid.UUID) (record, string, error) {
	r, ok, err := s.store.get(ctx, user)
	if err != nil {
		return record{}, "", err
	}
	if !ok {
		return record{}, "", ErrNoKey
	}
	pt, err := s.sealer.Open(r.Sealed, keyAD(user, r.Provider))
	if err != nil {
		return record{}, "", errUnreadable
	}
	return r, string(pt), nil
}

var errUnreadable = errors.New("modelkeys: the saved key could not be opened")

// clientFor builds the provider client for a record and its opened key.
func (s *Service) clientFor(r record, key string) (model.Client, error) {
	p, ok := LookupProvider(r.Provider)
	if !ok {
		return nil, errUnreadable
	}
	switch p.Transport {
	case TransportAnthropic:
		return &model.AnthropicClient{APIKey: key, Endpoint: p.BaseURL, HTTP: s.http}, nil
	default:
		endpoint := p.BaseURL
		if p.ID == CustomProvider {
			// Checked again, in case a row predates a tighter rule.
			u, err := ValidateCustomURL(r.BaseURL)
			if err != nil {
				return nil, errUnreadable
			}
			endpoint = u
		}
		return &model.OpenAIClient{
			Endpoint: endpoint, APIKey: key, Label: "own-key",
			// A hosted API refuses the `think` and `reasoning_effort`
			// fields a local Ollama endpoint takes.
			OmitThink: true,
			HTTP:      s.http,
		}, nil
	}
}

// TestResult is what the Settings page's Test button shows.
type TestResult struct {
	OK bool `json:"ok"`
	// Message is one of this package's fixed sentences.
	Message string `json:"message"`
	View    View   `json:"key"`
}

// Test makes one small call with the saved key and records the result
// on the row (last used, or the last error). It is a real request to the
// provider the person chose, under the same guard as a seat's.
func (s *Service) Test(ctx context.Context, user uuid.UUID) (TestResult, error) {
	r, key, err := s.openKey(ctx, user)
	if errors.Is(err, ErrNoKey) {
		return TestResult{}, ErrNoKey
	}
	now := s.now()
	if err != nil {
		if errors.Is(err, errUnreadable) {
			_ = s.store.setError(ctx, user, ReasonUnreadable, now)
			v, _, _ := s.Get(ctx, user)
			return TestResult{Message: ReasonUnreadable, View: v}, nil
		}
		return TestResult{}, err
	}
	c, err := s.clientFor(r, key)
	if err != nil {
		_ = s.store.setError(ctx, user, ReasonUnreadable, now)
		v, _, _ := s.Get(ctx, user)
		return TestResult{Message: ReasonUnreadable, View: v}, nil
	}
	callCtx, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()
	_, cerr := c.Complete(callCtx, model.Request{
		Model:     r.Model,
		User:      "Reply with the single word: ok",
		MaxTokens: 16,
	})
	res := TestResult{OK: cerr == nil, Message: "The key works: " + r.Model + " answered."}
	if cerr != nil {
		res.Message = testFailure(cerr)
		_ = s.store.setError(ctx, user, res.Message, s.now())
	} else {
		_ = s.store.clearError(ctx, user, s.now())
	}
	res.View, _, _ = s.Get(ctx, user)
	return res, nil
}
