package lobby

import (
	"crypto/subtle"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
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
}

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
	Source  draft.Source
	Mode    draft.Mode
	Seats   int
	// PickTimeout is the pick timer; zero is the default.
	PickTimeout time.Duration

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
		Name:        name,
		Set:         spec.Set,
		SetName:     spec.SetName,
		Booster:     spec.Booster,
		Mode:        spec.Mode,
		Seats:       spec.Seats,
		Seed:        seed,
		PickTimeout: spec.PickTimeout,
		Solo:        spec.Solo,
	}, time.Now())
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	entry := &draftEntry{room: draft.NewRoom(d, spec.Source)}
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

	var replaced []*draft.Room
	l.mu.Lock()
	// The creator's earlier unfinished draft goes first, so its cost
	// does not count against the new one.
	if entry.owner != "" {
		for id, e := range l.drafts {
			if e.owner == entry.owner && e.room.Status().State != draft.StateDone {
				replaced = append(replaced, e.room)
				delete(l.drafts, id)
			}
		}
	}
	if !l.limits.Fits(l.loadLocked(), l.limits.TableCost(1, 0)) {
		l.mu.Unlock()
		for _, r := range replaced {
			r.Close()
		}
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
	for _, r := range replaced {
		r.Close()
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
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
	humans, err := e.room.Leave(seat)
	if err != nil {
		return err
	}
	if humans == 0 {
		l.closeDraft(id)
	}
	return nil
}

// closeDraft ends a draft and forgets it.
func (l *Lobby) closeDraft(id uuid.UUID) {
	l.mu.Lock()
	e, ok := l.drafts[id]
	delete(l.drafts, id)
	l.mu.Unlock()
	if ok {
		e.room.Close()
	}
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
// seat's weight stands for. A finished draft counts nothing, like an
// ended table.
func (l *Lobby) draftCostLocked(e *draftEntry) int {
	st := e.room.Status()
	if st.Closed || st.State == draft.StateDone {
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
	}
	cands := make([]cand, 0, len(l.drafts))
	for id, e := range l.drafts {
		cands = append(cands, cand{id, e.room})
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
