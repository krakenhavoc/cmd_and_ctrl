package lobby

import (
	"crypto/subtle"
	"errors"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/token"
)

// Booster drafts (ADR 0222 §5, #5438).
//
// A draft is its own lobby object, not a game: it has no room in the
// ws.RoomManager, no games row, no snapshot and no engine. It lives in
// l.drafts for as long as it runs and is gone on a restart. Nothing
// about a person is written anywhere: persisting a finished draft is
// slice 6 (#5440), which is the slice that owes ADR 0149's privacy
// work.
//
// What it shares with a table, rather than forking: the name filter
// and the generated names (CreateWith), ADR 0148's load cap and
// reservation (capacity.go), the idle sweep (idle.go), and the
// authenticator, whose sessions its seats are.
//
// Lock order: l.mu, then a draft.Room's own lock. A Room never calls
// the lobby.

// Draft errors the HTTP layer maps to statuses.
var (
	// ErrDraftNotFound is an unknown draft, a closed one, and a wrong
	// invite: one answer for all three, so an invite cannot be probed.
	ErrDraftNotFound = errors.New("lobby: draft not found")
	// ErrDraftNoMatch is a seat asking for its match's table when it
	// has none: the round has not begun, its match is over, or it sits
	// this round out.
	ErrDraftNoMatch = errors.New("lobby: this seat has no match to play right now")
)

// draftEntry is one live draft. Its fields are fixed at creation.
type draftEntry struct {
	room *draft.Room
	// invite is the code a joiner presents, "" for a solo draft. It is
	// held in memory only, like the draft.
	invite string
	// owner is who the draft belongs to for the one-at-a-time rule:
	// "user:<id>" for a signed-in creator, "" for a visitor.
	owner string
	// set is the set drafted, for the basic lands a deck adds.
	set string

	// done is closed when the draft closes, which ends watchDraft.
	done     chan struct{}
	doneOnce sync.Once
	// sync serialises the steps that open a round's tables against a
	// seat leaving, so a table is never opened for a seat on its way
	// out (draft_matches.go).
	sync sync.Mutex
}

// finish marks the entry closed.
func (e *draftEntry) finish() { e.doneOnce.Do(func() { close(e.done) }) }

// DraftSpec is what POST /draft asks for, after the handler has
// resolved the set.
type DraftSpec struct {
	Name string
	// Set, SetName and Booster label the draft; Source makes its packs.
	// The handler builds Source only from a set the draftable check
	// passes.
	Set     string
	SetName string
	Booster string
	// Cube marks a draft of a saved card list (ADR 0222 §8): Set is
	// empty, SetName is the cube's name and Source deals it.
	Cube   bool
	Source draft.Source
	Mode   draft.Mode
	Seats  int
	// PickTimeout is the pick timer; zero is the default.
	PickTimeout time.Duration
	// BuildTimeout is the deckbuilding timer and BestOf the length of
	// the draft's matches (1 or 3); zero is the default of each.
	BuildTimeout time.Duration
	BestOf       int
	// BuildUntimed has the build run with no timer (#5700); it
	// overrides BuildTimeout.
	BuildUntimed bool

	// HostName is the creator's seat name, HostUser their account
	// (uuid.Nil for a visitor).
	HostName string
	HostUser uuid.UUID
	// Solo makes a visitor's draft: no invite, bots in every other
	// seat.
	Solo bool
	// Seed fixes the packs, for a test. Zero draws one.
	Seed uint64
}

// CreateDraft opens a draft in its lobby with the creator in the first
// seat, and returns the draft's id and that seat's.
//
// It is held to ADR 0148's load cap as a table is: the cost of the
// draft with its creator is reserved first, and a refused create costs
// nothing. A signed-in creator holds one unfinished draft at a time; a
// new one closes the old.
func (l *Lobby) CreateDraft(spec DraftSpec) (draftID, seatID uuid.UUID, err error) {
	name := trimToLimit(spec.Name, maxTableNameLen)
	if name != "" {
		if err := moderation.Name(name); err != nil {
			return uuid.Nil, uuid.Nil, err
		}
	} else {
		name = SuggestTableName(spec.HostName)
	}
	hostName := trimToLimit(spec.HostName, 40)
	if hostName == "" {
		hostName = "Player"
	}
	if err := moderation.Name(hostName); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	seed := spec.Seed
	if seed == 0 {
		seed = rand.Uint64()
	}
	d, err := draft.New(draft.Config{
		Name:         name,
		Set:          spec.Set,
		SetName:      spec.SetName,
		Booster:      spec.Booster,
		Cube:         spec.Cube,
		Mode:         spec.Mode,
		Seats:        spec.Seats,
		Seed:         seed,
		PickTimeout:  spec.PickTimeout,
		Solo:         spec.Solo,
		BuildTimeout: spec.BuildTimeout,
		BuildUntimed: spec.BuildUntimed,
		BestOf:       spec.BestOf,
	}, time.Now())
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	entry := &draftEntry{room: draft.NewRoom(d, spec.Source), set: spec.Set, done: make(chan struct{})}
	if !spec.Solo {
		entry.invite, err = token.Random(16)
		if err != nil {
			return uuid.Nil, uuid.Nil, err
		}
	}
	if spec.HostUser != uuid.Nil {
		entry.owner = "user:" + spec.HostUser.String()
	}
	invite := entry.invite
	entry.room.SetDecorator(func(_ uuid.UUID, v *draft.View) {
		// The invite goes to the host alone, and only while there is a
		// lobby to invite someone to.
		if invite == "" || v.State != draft.StateLobby || v.You < 0 || v.You >= len(v.Players) {
			return
		}
		if v.Players[v.You].Host {
			v.Invite = invite
		}
	})

	// The creator's earlier unfinished draft goes first, so its cost
	// and its tables' do not count against the new one. A draft whose
	// rounds are over is left for its people to look at.
	if entry.owner != "" {
		var replaced []uuid.UUID
		l.mu.Lock()
		for id, e := range l.drafts {
			if e.owner == entry.owner && e.room.Status().Phase != draft.PhaseFinished {
				replaced = append(replaced, id)
			}
		}
		l.mu.Unlock()
		for _, id := range replaced {
			l.closeDraft(id)
		}
	}
	l.mu.Lock()
	if !l.limits.Fits(l.loadLocked(), l.limits.TableCost(1, 0)) {
		l.mu.Unlock()
		metrics.CapacityRefused(metrics.RefuseTableLoad)
		return uuid.Nil, uuid.Nil, ErrServerFull
	}
	seatID, err = entry.room.Join(hostName, spec.HostUser)
	if err == nil {
		if l.drafts == nil {
			l.drafts = map[uuid.UUID]*draftEntry{}
		}
		l.drafts[d.ID] = entry
	}
	l.mu.Unlock()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	go l.watchDraft(d.ID, entry)
	return d.ID, seatID, nil
}

// draftEntryOf finds a live draft.
func (l *Lobby) draftEntryOf(id uuid.UUID) (*draftEntry, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.drafts[id]
	return e, ok
}

// DraftRoom is a live draft's room, or nil. It is how the draft socket
// and slice 6 (#5440, which reads the pools with Room.Pools) reach a
// draft.
func (l *Lobby) DraftRoom(id uuid.UUID) *draft.Room {
	e, ok := l.draftEntryOf(id)
	if !ok {
		return nil
	}
	return e.room
}

// inviteMatches compares a submitted invite in constant time. A solo
// draft has none and matches nothing.
func (e *draftEntry) inviteMatches(invite string) bool {
	invite = strings.TrimSpace(invite)
	return e.invite != "" && invite != "" &&
		subtle.ConstantTimeCompare([]byte(e.invite), []byte(invite)) == 1
}

// PreviewDraft is what the invite page shows. seat, when it is a seat
// of this draft, stands in for the invite.
func (l *Lobby) PreviewDraft(id uuid.UUID, invite string, seat uuid.UUID) (draft.Preview, error) {
	e, ok := l.draftEntryOf(id)
	if !ok {
		return draft.Preview{}, ErrDraftNotFound
	}
	if !e.inviteMatches(invite) && (seat == uuid.Nil || !e.room.HasSeat(seat)) {
		return draft.Preview{}, ErrDraftNotFound
	}
	return e.room.Preview(), nil
}

// JoinDraft takes a seat in a draft's lobby for whoever holds the
// invite. A signed-in person who already holds a seat gets that seat
// back, in any state: it is how they return from another device.
func (l *Lobby) JoinDraft(id uuid.UUID, invite, name string, userID uuid.UUID) (uuid.UUID, error) {
	e, ok := l.draftEntryOf(id)
	if !ok {
		return uuid.Nil, ErrDraftNotFound
	}
	if !e.inviteMatches(invite) {
		return uuid.Nil, ErrDraftNotFound
	}
	if seat := e.room.SeatOfUser(userID); seat != uuid.Nil {
		return seat, nil
	}
	name = trimToLimit(name, 40)
	if name == "" {
		name = "Player"
	}
	if err := moderation.Name(name); err != nil {
		return uuid.Nil, err
	}
	// A new person costs one human seat against the load cap.
	l.mu.Lock()
	fits := l.limits.Fits(l.loadLocked(), l.limits.HumanSeatWeight)
	l.mu.Unlock()
	if !fits {
		metrics.CapacityRefused(metrics.RefuseTableLoad)
		return uuid.Nil, ErrServerFull
	}
	return e.room.Join(name, userID)
}

// StartDraft starts a draft for its host.
func (l *Lobby) StartDraft(id, seat uuid.UUID) error {
	e, ok := l.draftEntryOf(id)
	if !ok {
		return ErrDraftNotFound
	}
	return e.room.Start(seat)
}

// LeaveDraft takes a person out of a draft. A draft nobody is left in
// closes.
func (l *Lobby) LeaveDraft(id, seat uuid.UUID) error {
	e, ok := l.draftEntryOf(id)
	if !ok {
		return ErrDraftNotFound
	}
	// A person who leaves in the middle of a match concedes it, and
	// what they drafted is kept before the seat forgets who they were
	// (draft_matches.go). Both under the draft's sync lock, so no table
	// is being opened for the seat as it goes.
	e.sync.Lock()
	l.concedeDraftMatch(id, seat)
	l.recordDraft(e, seat)
	humans, err := e.room.Leave(seat)
	e.sync.Unlock()
	if err != nil {
		return err
	}
	if humans == 0 {
		l.closeDraft(id)
	}
	return nil
}

// closeDraft ends a draft and forgets it: what its signed-in people
// drafted is kept first (a draft the sweep closes is as finished as it
// will get), and its match tables go with it.
func (l *Lobby) closeDraft(id uuid.UUID) {
	l.mu.Lock()
	e, ok := l.drafts[id]
	l.mu.Unlock()
	if !ok {
		return
	}
	l.recordDraft(e, uuid.Nil)
	l.mu.Lock()
	delete(l.drafts, id)
	l.mu.Unlock()
	e.finish()
	e.room.Close()
	l.dropDraftMatches(id)
}

// draftHasOpenMatch reports whether a draft has a match table still in
// play. Callers hold l.mu.
func (l *Lobby) draftHasOpenMatchLocked(id uuid.UUID) bool {
	for _, g := range l.games {
		if g.draftMatch != nil && g.draftMatch.draft == id && g.meta.State != string(game.StateEnded) {
			return true
		}
	}
	return false
}

// DraftCount is how many drafts are live.
func (l *Lobby) DraftCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.drafts)
}

// draftCostLocked is what a draft counts against the load cap (ADR
// 0222 §9, ADR 0148): one table, plus a human seat for each person,
// while it is in its lobby or people are picking. Its bots add nothing:
// a draft bot picks a card and runs no game, which is the work a bot
// seat's weight stands for. Once the rounds begin the draft counts
// nothing itself: each match it opens is a table and counts as one
// (#5440). A finished draft counts nothing, like an ended table.
func (l *Lobby) draftCostLocked(e *draftEntry) int {
	st := e.room.Status()
	if st.Closed {
		return 0
	}
	// Through the picks and the build the draft is one table with its
	// people. In the rounds its people are counted at the match tables
	// they play at, each a table of its own (draft_matches.go).
	if st.State == draft.StateDone && st.Phase != draft.PhaseBuilding {
		return 0
	}
	return l.limits.TableCost(st.Humans, 0)
}

// draftsLoadLocked is every live draft's cost.
func (l *Lobby) draftsLoadLocked() int {
	n := 0
	for _, e := range l.drafts {
		n += l.draftCostLocked(e)
	}
	return n
}

// SoloDraftAbandonedAfter is how long a visitor's draft may sit in its
// lobby with no socket before the sweep closes it. A visitor needs no
// account to make one, and an unstarted draft holds a table's weight
// against the load cap, so one that was opened and walked away from is
// not kept for the day an unstarted table is. Nobody else can be on
// their way to it: a solo draft has no invite.
const SoloDraftAbandonedAfter = 15 * time.Minute

// tickDraftsIdle closes idle drafts, the way the sweep closes idle
// tables (ADR 0211), with the same limits:
//
//   - a draft still in its lobby, after IdleLobbyAfter with no person
//     doing anything;
//   - a draft that is picking or finished, after IdleBotsAfter with no
//     person connected (its timer drafts on without them, so nobody is
//     left waiting on it), or after IdleGameAfter whatever is
//     connected;
//   - a visitor's draft still in its lobby, after
//     SoloDraftAbandonedAfter with its page closed;
//   - a draft with no person left in it, at once.
//
// A limit of zero is that close switched off. A draft has no warning
// line: it has no chat to post one in. It is not counted in the idle-
// table metric, which counts tables.
func (l *Lobby) tickDraftsIdle(now time.Time) {
	l.mu.Lock()
	lim := l.limits
	type cand struct {
		id   uuid.UUID
		room *draft.Room
		// playing says a match of the draft is still being played. Its
		// people are at that table, not on the draft's socket, and the
		// table has its own idle limit (draftMatchIdle).
		playing bool
	}
	cands := make([]cand, 0, len(l.drafts))
	for id, e := range l.drafts {
		cands = append(cands, cand{id, e.room, l.draftHasOpenMatchLocked(id)})
	}
	l.mu.Unlock()

	over := func(since time.Time, after time.Duration) bool {
		return after > 0 && !now.Before(since.Add(after))
	}
	for _, c := range cands {
		st := c.room.Status()
		due := st.Closed || st.Humans == 0
		switch {
		case due:
		case c.playing:
		case st.State == draft.StateLobby:
			due = over(st.LastHuman, lim.IdleLobbyAfter) ||
				(st.Solo && st.Connected == 0 && over(st.LastHuman, SoloDraftAbandonedAfter))
		default:
			due = (st.Connected == 0 && over(st.LastHuman, lim.IdleBotsAfter)) ||
				over(st.LastHuman, lim.IdleGameAfter)
		}
		if due {
			l.closeDraft(c.id)
		}
	}
}
