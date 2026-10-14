package lobby

// away.go: moving the table on without a player who is gone (ADR 0177).
//
// The hub knows which seats have a connection (ws/presence.go) and puts
// that on every frame. This file decides the rest:
//
//   - WHEN a seat is movable (§2): it has been offline for the table's
//     away time (the away_after setting, §7), the table is waiting on
//     it, and nobody has chosen to keep waiting through this absence.
//     "Waiting" is the notifier's own answer (notifyTables.Waiting), so
//     a DM and an offer never disagree about whose move it is. A hidden
//     tab never makes a seat movable. Practice tables never offer. A
//     ticker (RunAway) looks at every live table and tells the room
//     which seats are movable; the hub stamps the mark on their
//     presence, and the clients of the seats that decide offer the
//     choice.
//   - WHO decides (§3): the host while their own seat is present;
//     otherwise any present, seated, non-bot player still in the game;
//     the server admin always. Re-checked when the answer arrives.
//   - WHAT happens (§4–§6): a bot (the heuristic tier) plays the seat,
//     the seat is conceded for them, or the table keeps waiting for
//     another away time. One answer wins: the first one accepted makes
//     the seat no longer movable, and every later one is refused. A
//     player whose seat a bot is playing takes it back with a button.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// StandInTier is the bot tier that plays an away player's seat (ADR
// 0177 §4). The model tiers cost money and random would throw the game.
const StandInTier = "heuristic"

// AwayTickEvery is how often RunAway looks at the tables. The offer is
// only as punctual as this, and always late rather than early.
const AwayTickEvery = 5 * time.Second

// The answers to an away offer, POST /games/{id}/seats/{player}/away.
const (
	AwayActionStandIn = "stand_in"
	AwayActionConcede = "concede"
	AwayActionWait    = "wait"
)

var (
	// ErrSeatNotMovable refuses an answer for a seat the table may not
	// move on without (any more): it came back, somebody answered
	// first, or it was never away long enough.
	ErrSeatNotMovable = errors.New("lobby: that seat is not away, or someone has already answered")
	// ErrNotAwayDecider refuses an answer from someone who may not give
	// it (ADR 0177 §3).
	ErrNotAwayDecider = errors.New("lobby: only the host, or any player at the table while the host is away, may decide")
	// ErrNoStandIn refuses a take-back from a seat no bot is playing.
	ErrNoStandIn = errors.New("lobby: no bot is playing your seat")
	// ErrStandInUnavailable refuses a hand-over on a server that cannot
	// run the stand-in bot.
	ErrStandInUnavailable = errors.New("lobby: this server cannot hand a seat to a bot")
)

// AwayPresence is the hub's half: *ws.Hub satisfies it.
type AwayPresence interface {
	// SeatPresence is "" for a present seat, protocol.PresenceOffline
	// or protocol.PresenceHidden otherwise, and since when.
	SeatPresence(game, seat uuid.UUID) (state string, since time.Time)
	// RefreshPresence sends every viewer of the game a fresh frame.
	RefreshPresence(game uuid.UUID)
}

// StandInHost is the optional half of BotHost that runs a bot on one
// person's seat without touching the table's own bots. Satisfied by
// *aiseat.Manager. A host without it cannot hand a seat over.
type StandInHost interface {
	StartStandIn(room *ws.Room, seat aiseat.SeatSpec)
	StopStandIn(gameID, seat uuid.UUID)
}

type awayKey struct{ game, seat uuid.UUID }

// awaySnooze is one "keep waiting": through the absence that began at
// absence, until until.
type awaySnooze struct{ absence, until time.Time }

// awayState is the lobby's own away bookkeeping. Guarded by its own
// mutex, never l.mu, and nothing is called while holding it.
type awayState struct {
	mu      sync.Mutex
	snoozed map[awayKey]awaySnooze
	// answerMu makes one answer at a time: the check that the seat is
	// still movable and what the answer does happen together, so the
	// first answer is the only one. Taken before l.mu, never after.
	answerMu sync.Mutex
	// seen is each table's presence as the last tick saw it, so a tick
	// sends a fresh frame only to a table where someone came or went.
	seen map[uuid.UUID]string
	// now is the clock; nil is time.Now. Tests set it.
	now func() time.Time
}

// SetAwayPresence wires the hub in at boot. Nil (tests, a server with
// no hub) offers nothing.
func (l *Lobby) SetAwayPresence(p AwayPresence) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.awayPresence = p
}

// SetAwayClock replaces the away clock. Tests only.
func (l *Lobby) SetAwayClock(now func() time.Time) {
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	l.away.now = now
}

func (l *Lobby) awayNow() time.Time {
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	if l.away.now != nil {
		return l.away.now()
	}
	return time.Now()
}

func (l *Lobby) awayHub() AwayPresence {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.awayPresence
}

// snoozedFor reports whether the table chose to keep waiting through
// the seat's absence that began at since.
func (l *Lobby) snoozedFor(k awayKey, since, now time.Time) bool {
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	s, ok := l.away.snoozed[k]
	return ok && s.absence.Equal(since) && now.Before(s.until)
}

func (l *Lobby) snooze(k awayKey, since, until time.Time) {
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	if l.away.snoozed == nil {
		l.away.snoozed = make(map[awayKey]awaySnooze)
	}
	l.away.snoozed[k] = awaySnooze{absence: since, until: until}
}

func (l *Lobby) pruneSnoozes(now time.Time) {
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	for k, s := range l.away.snoozed {
		if !now.Before(s.until) {
			delete(l.away.snoozed, k)
		}
	}
}

// awayTable is one table as the away clock needs it, copied under l.mu.
type awayTable struct {
	id   uuid.UUID
	room *ws.Room
	// seats are the seats that may be moved on: people and agents, not
	// bots and not seats a bot already plays.
	seats []uuid.UUID
	// watched are the seats whose presence the table is shown: every
	// seat but the bots.
	watched []uuid.UUID
	// after is the away time; 0 is a table that never offers.
	after time.Duration
}

// awayAfterOf is entry's away time, 0 when it never offers: a practice
// table, a table not in play, or an away_after of 0. Caller holds l.mu.
func awayAfterOf(entry *gameEntry) time.Duration {
	if entry.practice || entry.room == nil || entry.meta.Archived() ||
		entry.room.Game.CurrentState() != game.StateActive {
		return 0
	}
	return time.Duration(entry.room.Game.TableSettingsSnapshot().EffectiveAwayAfter()) * time.Minute
}

// awayTableLocked copies entry for the away clock. Caller holds l.mu.
func awayTableLocked(entry *gameEntry) awayTable {
	t := awayTable{id: entry.meta.ID, room: entry.room, after: awayAfterOf(entry)}
	for _, s := range entry.meta.Players {
		if s.IsBot {
			continue
		}
		t.watched = append(t.watched, s.PlayerID)
		if s.StandIn == "" {
			t.seats = append(t.seats, s.PlayerID)
		}
	}
	return t
}

// awayOffline is how long seat has been offline, false when it is not
// offline (present, or only hidden).
func awayOffline(p AwayPresence, gameID, seat uuid.UUID, now time.Time) (time.Duration, time.Time, bool) {
	state, since := p.SeatPresence(gameID, seat)
	if state != protocol.PresenceOffline {
		return 0, time.Time{}, false
	}
	return now.Sub(since), since, true
}

// movable reports whether the table may move on without seat now, and
// how long the seat has been offline (ADR 0177 §2).
func (l *Lobby) movable(p AwayPresence, t awayTable, seat uuid.UUID, now time.Time) (bool, time.Duration) {
	if p == nil || t.after <= 0 {
		return false, 0
	}
	gone, since, offline := awayOffline(p, t.id, seat, now)
	if !offline || gone < t.after {
		return false, 0
	}
	if l.snoozedFor(awayKey{t.id, seat}, since, now) {
		return false, 0
	}
	if waiting, _ := (notifyTables{l}).Waiting(t.id, seat); !waiting {
		return false, 0
	}
	return true, gone
}

// RunAway ticks the away clock every AwayTickEvery until ctx ends.
func (l *Lobby) RunAway(ctx context.Context) {
	t := time.NewTicker(AwayTickEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.TickAway()
		}
	}
}

// TickAway looks at every live table once: which seats may the table
// move on without? A table whose answer changed gets a fresh frame, so
// the offer appears, and goes, without waiting for a move. Exported
// for tests, which drive it with a fake clock.
func (l *Lobby) TickAway() {
	p := l.awayHub()
	if p == nil {
		return
	}
	now := l.awayNow()
	l.pruneSnoozes(now)
	l.mu.Lock()
	tables := make([]awayTable, 0, len(l.games))
	for _, e := range l.games {
		if e.room == nil {
			continue
		}
		tables = append(tables, awayTableLocked(e))
	}
	l.mu.Unlock()
	live := make(map[uuid.UUID]bool, len(tables))
	for _, t := range tables {
		live[t.id] = true
		l.refreshMovable(p, t, now)
	}
	l.away.mu.Lock()
	for id := range l.away.seen {
		if !live[id] {
			delete(l.away.seen, id)
		}
	}
	l.away.mu.Unlock()
}

// presenceChanged reports whether t's seats' presence differs from what
// the last tick saw, and remembers it.
func (l *Lobby) presenceChanged(p AwayPresence, t awayTable) bool {
	var b strings.Builder
	for _, seat := range t.watched {
		state, since := p.SeatPresence(t.id, seat)
		b.WriteString(seat.String())
		b.WriteByte('=')
		b.WriteString(state)
		if state != "" {
			b.WriteByte('@')
			b.WriteString(strconv.FormatInt(since.Unix(), 10))
		}
		b.WriteByte(';')
	}
	sig := b.String()
	l.away.mu.Lock()
	defer l.away.mu.Unlock()
	if l.away.seen == nil {
		l.away.seen = make(map[uuid.UUID]string)
	}
	old, ok := l.away.seen[t.id]
	l.away.seen[t.id] = sig
	return ok && old != sig
}

// refreshMovable recomputes t's movable seats and, when they or anyone's
// presence changed, sends the table a fresh frame (ADR 0177 §1–§2).
func (l *Lobby) refreshMovable(p AwayPresence, t awayTable, now time.Time) {
	set := make(map[uuid.UUID]bool)
	for _, seat := range t.seats {
		if ok, _ := l.movable(p, t, seat, now); ok {
			set[seat] = true
		}
	}
	changed := t.room.SetMovable(set)
	if l.presenceChanged(p, t) || changed {
		p.RefreshPresence(t.id)
	}
}

// awayDecider checks that caller may answer for the table entry
// describes about seat (ADR 0177 §3). It reads the hub, so the caller
// must not hold l.mu.
func (l *Lobby) awayDecider(p AwayPresence, caller auth.Principal, admin bool, meta GameMeta, room *ws.Room, seat uuid.UUID) error {
	if admin {
		return nil
	}
	if caller.Role != auth.RolePlayer || caller.GameID != meta.ID || caller.PlayerID == uuid.Nil || caller.PlayerID == seat {
		return ErrNotAwayDecider
	}
	mine, ok := findSeat(meta.Players, caller.PlayerID)
	if !ok || mine.IsBot || room.Game.SeatRemoved(caller.PlayerID) || seatEliminated(room, caller.PlayerID) {
		return ErrNotAwayDecider
	}
	present := func(id uuid.UUID) bool {
		state, _ := p.SeatPresence(meta.ID, id)
		return state != protocol.PresenceOffline
	}
	if !present(caller.PlayerID) {
		return ErrNotAwayDecider
	}
	host := room.HostPlayerID()
	if host != uuid.Nil && host != seat && host != caller.PlayerID && present(host) {
		return ErrNotAwayDecider
	}
	return nil
}

// seatEliminated reports whether seat has left the game.
func seatEliminated(room *ws.Room, seat uuid.UUID) bool {
	out := false
	room.Game.ReadSnapshot(func() {
		for _, pl := range room.Game.Seats {
			if pl != nil && pl.ID == seat {
				out = pl.Eliminated
			}
		}
	})
	return out
}

// AnswerAway carries out one answer to the away offer for seat at table
// id (ADR 0177 §2): action is AwayActionStandIn, AwayActionConcede or
// AwayActionWait. It returns the table as it now stands.
func (l *Lobby) AnswerAway(id uuid.UUID, caller auth.Principal, admin bool, seat uuid.UUID, action string) (GameMeta, error) {
	switch action {
	case AwayActionStandIn, AwayActionConcede, AwayActionWait:
	default:
		return GameMeta{}, httpError(http.StatusBadRequest, `action must be "stand_in", "concede" or "wait"`)
	}
	p := l.awayHub()
	if p == nil {
		return GameMeta{}, ErrSeatNotMovable
	}
	l.away.answerMu.Lock()
	defer l.away.answerMu.Unlock()
	l.mu.Lock()
	entry, ok := l.games[id]
	if !ok {
		l.mu.Unlock()
		return GameMeta{}, ErrGameNotFound
	}
	if _, ok := findSeat(entry.meta.Players, seat); !ok {
		l.mu.Unlock()
		return GameMeta{}, ErrPlayerNotInGame
	}
	meta, room := copyMeta(entry.meta), entry.room
	t := awayTableLocked(entry)
	standIns, _ := l.bots.(StandInHost)
	l.mu.Unlock()

	if err := l.awayDecider(p, caller, admin, meta, room, seat); err != nil {
		return GameMeta{}, err
	}
	if action == AwayActionStandIn && (standIns == nil || !l.standInTierAvailable()) {
		return GameMeta{}, ErrStandInUnavailable
	}
	now := l.awayNow()
	movable := false
	var gone time.Duration
	for _, s := range t.seats {
		if s == seat {
			movable, gone = l.movable(p, t, seat, now)
		}
	}
	if !movable {
		return GameMeta{}, ErrSeatNotMovable
	}
	minutes := int(gone / time.Minute)
	actor := caller.PlayerID
	if admin && (caller.GameID != id || caller.Role != auth.RolePlayer) {
		actor = uuid.Nil
	}

	var (
		out   GameMeta
		err   error
		after func()
	)
	switch action {
	case AwayActionWait:
		_, since, _ := awayOffline(p, id, seat, now)
		l.snooze(awayKey{id, seat}, since, now.Add(t.after))
		out, err = l.Get(id)
	case AwayActionStandIn:
		out, after, err = l.handOver(id, seat, actor, minutes, now, standIns)
	case AwayActionConcede:
		out, err = l.concedeAway(id, seat, actor, minutes)
	}
	if err != nil {
		return GameMeta{}, err
	}
	if after != nil {
		after()
	}
	// The offer is gone for every decider at once: the seat is no
	// longer movable, so the recount changes and the table is sent a
	// frame.
	l.mu.Lock()
	if e, ok := l.games[id]; ok {
		t = awayTableLocked(e)
	}
	l.mu.Unlock()
	l.refreshMovable(p, t, now)
	slog.Default().Info("away seat answered", "game_id", id, "seat", seat, "action", action, "offline_minutes", minutes)
	return out, nil
}

// standInTierAvailable reports whether the bot host offers StandInTier.
func (l *Lobby) standInTierAvailable() bool {
	l.mu.Lock()
	bots := l.bots
	l.mu.Unlock()
	if bots == nil {
		return false
	}
	for _, t := range bots.Tiers() {
		if t == StandInTier {
			return true
		}
	}
	return false
}

// handOver records that a bot plays seat from now (ADR 0177 §4) and
// returns the thunk that starts it, to run once l.mu is released.
func (l *Lobby) handOver(id, seat, actor uuid.UUID, minutes int, now time.Time, host StandInHost) (GameMeta, func(), error) {
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, nil, ErrGameNotFound
	}
	idx := seatIndexOf(entry.meta.Players, seat)
	if idx < 0 {
		return GameMeta{}, nil, ErrPlayerNotInGame
	}
	if entry.meta.Players[idx].IsBot || entry.meta.Players[idx].StandIn != "" {
		return GameMeta{}, nil, ErrSeatNotMovable
	}
	// Marked on the room first, so the capture ApplyExternal takes
	// carries it, and the seat's own actions are refused from here on.
	at := now.UTC()
	entry.room.SetStandIn(seat, StandInTier, at)
	view, seq, err := entry.room.ApplyExternal(func() error {
		return entry.room.Game.NoteAwaySeat(actor, seat, game.AwayStandIn, minutes)
	})
	if err != nil {
		entry.room.ClearStandIn(seat)
		if errors.Is(err, game.ErrGameNotActive) {
			return GameMeta{}, nil, ErrSeatNotMovable
		}
		return GameMeta{}, nil, err
	}
	entry.meta.Players[idx].StandIn = StandInTier
	entry.meta.Players[idx].StandInSince = &at
	l.persistSeatsLocked(entry)
	if l.broadcast != nil {
		b := l.broadcast
		broadcast = func() { b.BroadcastState(id, seq, view) }
	}
	room := entry.room
	spec := aiseat.SeatSpec{PlayerID: seat, Tier: StandInTier}
	return copyMeta(entry.meta), func() { host.StartStandIn(room, spec) }, nil
}

// concedeAway concedes seat for its player (ADR 0177 §5).
func (l *Lobby) concedeAway(id, seat, actor uuid.UUID, minutes int) (GameMeta, error) {
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if s, ok := findSeat(entry.meta.Players, seat); !ok || s.IsBot || s.StandIn != "" {
		return GameMeta{}, ErrSeatNotMovable
	}
	view, seq, err := entry.room.ApplyExternalClearingUndo(func() error {
		return entry.room.Game.ConcedeForAway(actor, seat, minutes)
	})
	if err != nil {
		if errors.Is(err, game.ErrGameNotActive) || errors.Is(err, game.ErrPlayerEliminated) {
			return GameMeta{}, ErrSeatNotMovable
		}
		return GameMeta{}, err
	}
	if l.broadcast != nil {
		b := l.broadcast
		broadcast = func() { b.BroadcastState(id, seq, view) }
	}
	// The concession can end the game.
	l.syncStateLocked(entry)
	l.syncHostLocked(entry)
	return copyMeta(entry.meta), nil
}

// TakeBack gives seat back to its player from the bot standing in (ADR
// 0177 §6). The runner is stopped first, at its next decision boundary,
// and only then may the player act.
func (l *Lobby) TakeBack(id, seat uuid.UUID) (GameMeta, error) {
	l.mu.Lock()
	entry, ok := l.games[id]
	if !ok {
		l.mu.Unlock()
		return GameMeta{}, ErrGameNotFound
	}
	s, ok := findSeat(entry.meta.Players, seat)
	if !ok {
		l.mu.Unlock()
		return GameMeta{}, ErrPlayerNotInGame
	}
	if s.StandIn == "" {
		l.mu.Unlock()
		return GameMeta{}, ErrNoStandIn
	}
	host, _ := l.bots.(StandInHost)
	l.mu.Unlock()

	if host != nil {
		host.StopStandIn(id, seat)
	}

	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok = l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	idx := seatIndexOf(entry.meta.Players, seat)
	if idx < 0 {
		return GameMeta{}, ErrPlayerNotInGame
	}
	if entry.meta.Players[idx].StandIn == "" {
		// A second press raced this one, and won.
		return copyMeta(entry.meta), nil
	}
	entry.meta.Players[idx].StandIn = ""
	entry.meta.Players[idx].StandInSince = nil
	entry.room.ClearStandIn(seat)
	l.persistSeatsLocked(entry)
	view, seq, err := entry.room.ApplyExternal(func() error {
		return entry.room.Game.NoteAwaySeat(seat, seat, game.AwayTakeBack, 0)
	})
	switch {
	case err == nil:
	case errors.Is(err, game.ErrGameNotActive):
		// The game ended while the bot played: the seat is theirs
		// again, and there is no game left to log it in.
		view, seq, err = entry.room.Snapshot()
		if err != nil {
			return copyMeta(entry.meta), nil
		}
	default:
		return GameMeta{}, err
	}
	if l.broadcast != nil {
		b := l.broadcast
		broadcast = func() { b.BroadcastState(id, seq, view) }
	}
	return copyMeta(entry.meta), nil
}

// standInStartLocked returns a thunk that starts the runner of every
// seat of entry a bot stands in for, or nil when there are none. For a
// restore and an unarchive, beside botStartLocked. Callers hold l.mu
// and run the thunk after releasing it.
func (l *Lobby) standInStartLocked(entry *gameEntry) func() {
	host, ok := l.bots.(StandInHost)
	if !ok || entry == nil || entry.room == nil {
		return nil
	}
	var specs []aiseat.SeatSpec
	for _, s := range entry.meta.Players {
		if s.StandIn != "" && !s.IsBot {
			specs = append(specs, aiseat.SeatSpec{PlayerID: s.PlayerID, Tier: s.StandIn})
		}
	}
	if len(specs) == 0 {
		return nil
	}
	room := entry.room
	return func() {
		for _, spec := range specs {
			host.StartStandIn(room, spec)
		}
	}
}

// markStandInsLocked hands a restored entry's stand-ins to its room, so
// the mark is on its views and the seat's own actions are refused while
// the bot plays it. Caller holds l.mu.
func markStandInsLocked(entry *gameEntry) {
	for _, s := range entry.meta.Players {
		if s.StandIn == "" {
			continue
		}
		since := time.Now().UTC()
		if s.StandInSince != nil {
			since = *s.StandInSince
		}
		entry.room.SetStandIn(s.PlayerID, s.StandIn, since)
	}
}

func seatIndexOf(seats []SeatInfo, playerID uuid.UUID) int {
	for i, s := range seats {
		if s.PlayerID == playerID {
			return i
		}
	}
	return -1
}

// awaySeatRequest is the body of POST /games/{id}/seats/{player}/away.
type awaySeatRequest struct {
	Action string `json:"action"`
}

// answerAwaySeat handles POST /games/{id}/seats/{player}/away: one
// answer to the offer the table gets when a seat it is waiting on has
// been offline for its away time (ADR 0177 §2).
//
//	200  done; the table as it now stands
//	400  the action is not one of stand_in, concede, wait
//	401  anonymous
//	403  not the host, nor a present player while the host is away, nor the
//	     admin; or player is not a seat here
//	404  no such table
//	409  the seat is not movable: it came back, someone answered first, or
//	     it has not been away long enough
//	503  stand_in on a server that cannot run the bot
func answerAwaySeat(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	seat, err := uuid.Parse(r.PathValue("player"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid player id")
	}
	var req awaySeatRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	admin := c.isAdmin(p)
	meta, err := c.Lobby.AnswerAway(id, p, admin, seat, req.Action)
	if err != nil {
		return awayHTTPError(err)
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, admin, id, meta))
}

// takeBackSeat handles POST /games/{id}/seats/me/take-back: the
// caller's own seat, which a bot has been playing while they were away,
// is theirs to play again (ADR 0177 §6).
//
//	200  done; the table as it now stands
//	401  anonymous
//	403  not a seat at this table
//	409  no bot is playing the caller's seat
func takeBackSeat(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RolePlayer || p.GameID != id || p.PlayerID == uuid.Nil {
		return httpError(http.StatusForbidden, "only a seat at this table can take its seat back")
	}
	meta, err := c.Lobby.TakeBack(id, p.PlayerID)
	if err != nil {
		return awayHTTPError(err)
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, c.isAdmin(p), id, meta))
}

// awayHTTPError gives the away errors their status codes.
func awayHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrSeatNotMovable), errors.Is(err, ErrNoStandIn):
		return httpError(http.StatusConflict, strings.TrimPrefix(err.Error(), "lobby: "))
	case errors.Is(err, ErrNotAwayDecider):
		return httpError(http.StatusForbidden, strings.TrimPrefix(err.Error(), "lobby: "))
	case errors.Is(err, ErrStandInUnavailable):
		return httpError(http.StatusServiceUnavailable, strings.TrimPrefix(err.Error(), "lobby: "))
	}
	return err
}
