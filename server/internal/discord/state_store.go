package discord

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StateEntry binds an in-flight OAuth round-trip to the invite it
// started from. The client hits /auth/discord/start?game=&t=, we
// stamp a fresh state + PKCE verifier, park the (game_id, invite,
// verifier) under that state, then send the user off to Discord.
// When Discord bounces them back to /callback?state=&code=, we
// look up the entry to recover which game/invite the code is for
// and to verify the PKCE hash.
//
// This matters for two reasons: (1) a user who clicks two invite
// links in quick succession shouldn't have the second overwrite
// the first, so each start gets its own state; (2) a forged
// callback without the matching state is rejected, which is the
// CSRF guarantee PKCE + state are for.
type StateEntry struct {
	GameID       uuid.UUID
	InviteToken  string
	CodeVerifier string
	CreatedAt    time.Time
}

// StateStore is a tiny in-memory TTL map. Sized for at most a
// few dozen in-flight OAuths at once (hobby scale); a real
// service would put this in Redis or similar, but the rest of
// the server keeps state in RAM too and this matches.
//
// 5-minute TTL matches Discord's recommended state lifetime —
// enough time for a user to alt-tab, authorise, and come back,
// short enough that a lost mobile click doesn't clog the map
// forever.
type StateStore struct {
	mu      sync.Mutex
	entries map[string]StateEntry
	ttl     time.Duration
	nowFn   func() time.Time // overridable for tests
}

// NewStateStore returns a store with the conventional 5-minute
// TTL. Callers that want a different TTL (tests, mostly) can
// construct the zero value and set ttl + nowFn before first use.
func NewStateStore() *StateStore {
	return &StateStore{
		entries: make(map[string]StateEntry),
		ttl:     5 * time.Minute,
		nowFn:   time.Now,
	}
}

// ErrStateNotFound is returned by Consume when the state isn't in
// the store — either it's expired, already consumed, or forged.
// The callback handler maps it to a 400 (bad_request) rather than
// a 401 because the client can reasonably retry by re-starting
// the OAuth flow.
var ErrStateNotFound = errors.New("discord: oauth state not found or expired")

// Start mints a state + PKCE verifier, parks them with the caller-
// supplied invite identity, and returns (state, codeChallenge).
// Caller feeds both into the Discord authorize URL: state becomes
// the ?state= param, codeChallenge becomes ?code_challenge=.
func (s *StateStore) Start(game uuid.UUID, inviteToken string) (state, codeChallenge string, err error) {
	// Both values are 32 random bytes → base64url. state doubles as
	// the map key; verifier is the PKCE S256 source. Separate
	// randomness so compromising one doesn't compromise the other.
	stateBytes := make([]byte, 32)
	if _, err = rand.Read(stateBytes); err != nil {
		return "", "", err
	}
	verifierBytes := make([]byte, 32)
	if _, err = rand.Read(verifierBytes); err != nil {
		return "", "", err
	}
	state = base64.RawURLEncoding.EncodeToString(stateBytes)
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)
	challengeHash := sha256.Sum256([]byte(verifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(challengeHash[:])

	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	s.entries[state] = StateEntry{
		GameID:       game,
		InviteToken:  inviteToken,
		CodeVerifier: verifier,
		CreatedAt:    s.nowFn(),
	}
	return state, codeChallenge, nil
}

// Consume removes and returns the entry for the given state.
// Single-use: a successful lookup deletes the map entry, so a
// replayed callback (e.g. user hitting back and forward) gets
// ErrStateNotFound on the second shot. That's the desired
// behaviour — OAuth codes are also single-use on Discord's side,
// so the second exchange would fail anyway.
func (s *StateStore) Consume(state string) (StateEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	entry, ok := s.entries[state]
	if !ok {
		return StateEntry{}, ErrStateNotFound
	}
	delete(s.entries, state)
	return entry, nil
}

// gcLocked drops expired entries. Called opportunistically on
// every Start / Consume so the map doesn't grow unbounded if
// users start OAuths and never complete them. Caller holds mu.
func (s *StateStore) gcLocked() {
	cutoff := s.nowFn().Add(-s.ttl)
	for k, e := range s.entries {
		if e.CreatedAt.Before(cutoff) {
			delete(s.entries, k)
		}
	}
}
