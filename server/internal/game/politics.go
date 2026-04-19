package game

import (
	"errors"

	"github.com/google/uuid"
)

// politics.go holds the S10 "politics scaffold" types and mutations:
// per-pair promise tokens and a single active vote at a time. Both
// are sandbox-only — the rules graft track will hook keyword effects
// (Council's Dilemma, Will of the Council, vote-from-among, etc.)
// onto these primitives if the playgroup wants automation later.

// PromiseKey is the directed (from, to) key for a promise-token entry.
// Stored as a struct rather than a "uuid|uuid" string so the type
// system catches any accidental swap and so JSON marshaling is
// explicit on the wire.
type PromiseKey struct {
	From uuid.UUID
	To   uuid.UUID
}

// Vote is the state of the currently open council's-dilemma /
// politics vote, if any. There is at most one open vote per game
// (no nested votes — if a card needs that, end the outer vote first).
//
// The Initiator is recorded so the UI can show "Alice called this"
// and so EndVote can surface a permission gate later if needed
// (today any seated player may end any vote, mirroring the
// "clear_combat is loose" sandbox posture).
type Vote struct {
	ID        uuid.UUID
	Topic     string
	Options   []string
	Initiator uuid.UUID
	// Ballots maps voter player ID → option index (0-based into
	// Options). A player who hasn't voted is absent. Re-voting
	// overwrites their previous ballot.
	Ballots map[uuid.UUID]int
}

// ErrVoteAlreadyOpen is returned by StartVote when a vote is already
// in progress and a new one is attempted. The client must call
// EndVote first.
var ErrVoteAlreadyOpen = errors.New("game: a vote is already open")

// ErrNoVote is returned by CastVote / EndVote when no vote is
// currently open.
var ErrNoVote = errors.New("game: no vote in progress")

// ErrInvalidOption is returned by CastVote when the option index is
// outside the bounds of the current Vote.Options slice.
var ErrInvalidOption = errors.New("game: vote option index out of range")

// SetPromise sets the promise-token count for a directed pair (from
// → to). Set semantics so two stale tabs don't double-count. count
// is clamped at 0 from below; entries explicitly set to zero are
// left in the map for rendering simplicity (the wire view drops
// them).
//
// Both `from` and `to` must be seated; the call is rejected with
// ErrPlayerNotFound otherwise.
func (g *Game) SetPromise(from, to uuid.UUID, count int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(from) == nil {
		return ErrPlayerNotFound
	}
	if g.playerByIDLocked(to) == nil {
		return ErrPlayerNotFound
	}
	if count < 0 {
		count = 0
	}
	if g.Promises == nil {
		g.Promises = make(map[PromiseKey]int)
	}
	g.Promises[PromiseKey{From: from, To: to}] = count
	return nil
}

// StartVote opens a new vote. Topic is an arbitrary string ("who
// gets the monarchy?", "Brago's representative gives draw or
// discard"). Options must contain at least two entries; an empty
// option list is rejected with ErrInvalidParam.
//
// Returns ErrVoteAlreadyOpen if a vote is currently in progress;
// callers must EndVote first.
func (g *Game) StartVote(initiator uuid.UUID, topic string, options []string) (uuid.UUID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return uuid.Nil, ErrGameNotActive
	}
	if g.playerByIDLocked(initiator) == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	if len(options) < 2 {
		return uuid.Nil, ErrInvalidParam
	}
	if g.Vote != nil {
		return uuid.Nil, ErrVoteAlreadyOpen
	}
	v := &Vote{
		ID:        uuid.New(),
		Topic:     topic,
		Options:   append([]string(nil), options...),
		Initiator: initiator,
		Ballots:   make(map[uuid.UUID]int),
	}
	g.Vote = v
	return v.ID, nil
}

// CastVote records or overwrites `voter`'s ballot at the given option
// index. Idempotent if the voter casts the same ballot again.
//
// Returns ErrNoVote if there is no open vote, ErrInvalidOption if
// the index is out of range, ErrPlayerNotFound if the voter isn't
// seated.
func (g *Game) CastVote(voter uuid.UUID, option int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Vote == nil {
		return ErrNoVote
	}
	if option < 0 || option >= len(g.Vote.Options) {
		return ErrInvalidOption
	}
	if g.playerByIDLocked(voter) == nil {
		return ErrPlayerNotFound
	}
	g.Vote.Ballots[voter] = option
	return nil
}

// EndVote closes the currently open vote and clears g.Vote. Returns
// ErrNoVote if there is no open vote. Tally is preserved on the
// returned Vote pointer so callers (the action handler) can log it
// or surface a "winner" to the client; the next snapshot drops the
// Vote field entirely so the modal closes for everyone.
func (g *Game) EndVote() (*Vote, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return nil, ErrGameNotActive
	}
	if g.Vote == nil {
		return nil, ErrNoVote
	}
	finished := g.Vote
	g.Vote = nil
	return finished, nil
}
