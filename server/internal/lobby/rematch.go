package lobby

// rematch.go is the rematch (#3285, ADR 0164): the host of a finished
// table offers the same game again, the other humans accept or leave,
// and the host starts it.
//
// # The shape
//
// A rematch is an ordinary lobby table, created when the host offers it:
// the same table settings, the host's seat with the host's deck, and a
// bot seat for every bot of the old table with the same tier, deck and
// name. Nothing about it is special once it exists, so the host starts it,
// adds a bot to a free seat, or leaves a seat empty through the routes and
// the lobby screen every table has. It starts with the opening roll, so the
// turn order is random, as at any new game.
//
// The old table keeps its state, its replay and its place in History. The
// new table is a new game with a new id.
//
// # Who is asked, and for how long
//
// Every human seat of the old table other than the host is asked, on the
// results screen, through GameView.Rematch (protocol.RematchView). An
// accept seats that person at the new table with their own deck and hands
// them a session for the new seat; leaving is a no. The offer is open
// until every human who has not answered has left the results screen
// (the owner's decision 3 on #3285): the watcher polls the hub for a
// socket on each such seat, and an absence of rematchGrace closes it, so a
// reload does not. A host with no other human to ask closes it at once.
// Closing keeps the new table and whoever it already holds; it only stops
// anyone else being carried over.
//
// # Private
//
// The new table has no invite and no listing: its two invites are revoked
// as it is made, and only the people the offer carries over ever hold a
// seat there (decision 4). The host may fill a free seat with a bot or
// leave it empty. Nothing reopens it to the lobby.
//
// # Decks
//
// A started game no longer holds its decks, so Start keeps a deep copy of
// every seat's deck on the entry (game.Game.LobbyDecks). It is dropped
// when the offer closes, and rematchKeep after the game ended whether or
// not anyone offered. A table the server restored from disk has none, and
// a rematch of it is refused.
//
// # Agents
//
// An MCP seat is not carried over: its session lives in the client's
// process, which has no way to learn of a new table. It is left out of the
// rematch and does not hold the offer open.

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

var (
	// ErrRematchNotEnded refuses an offer on a table that is not over.
	ErrRematchNotEnded = errors.New("lobby: a rematch can be offered once the game is over")
	// ErrRematchNotHost refuses an offer from anyone but the host seat.
	ErrRematchNotHost = errors.New("lobby: only the table's host can offer a rematch")
	// ErrRematchUnavailable refuses an offer the server cannot honour: a
	// practice table, an archived one, or one whose decks are no longer
	// kept (it was restored after a restart, or ended long ago).
	ErrRematchUnavailable = errors.New("lobby: this table can no longer be rematched")
	// ErrRematchOffered refuses a second offer on the same table.
	ErrRematchOffered = errors.New("lobby: a rematch was already offered at this table")
	// ErrNoRematch answers an accept or a leave when there is no offer.
	ErrNoRematch = errors.New("lobby: nobody has offered a rematch at this table")
	// ErrRematchClosed refuses an accept after the offer closed: every
	// other human left, or this person already answered.
	ErrRematchClosed = errors.New("lobby: the rematch offer is closed")
	// ErrRematchNotAsked refuses an answer from a seat the offer does not
	// ask: a bot, an agent, or the host, who offered.
	ErrRematchNotAsked = errors.New("lobby: this seat is not asked to join the rematch")
	// ErrRematchAccepted refuses leaving an offer already accepted: the
	// seat is at the new table.
	ErrRematchAccepted = errors.New("lobby: you already joined the rematch")
)

const (
	// rematchKeep is how long after a game ends its decks are kept for a
	// rematch no one has offered yet.
	rematchKeep = 4 * time.Hour
	// defaultRematchPoll and defaultRematchGrace are how often the
	// watcher asks the hub who is still on the results screen, and how
	// long a seat may be gone before it counts as having left.
	defaultRematchPoll  = time.Second
	defaultRematchGrace = 15 * time.Second
	// rematchSuffix is added to the old table's name.
	rematchSuffix = " (rematch)"
)

// SeatPresence says which seats of a game have a live socket. Satisfied
// by *ws.Hub; narrow so lobby tests can script who is watching.
type SeatPresence interface {
	// ConnectedSeats is the set of player ids that hold a socket on the
	// game, however many tabs each has open.
	ConnectedSeats(gameID uuid.UUID) map[uuid.UUID]bool
}

// SetSeatPresence wires the hub in after construction, like
// SetStateBroadcaster. Without it a rematch offer never closes by itself.
func (l *Lobby) SetSeatPresence(p SeatPresence) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.presence = p
}

// setRematchTiming shrinks the watcher's clocks for tests.
func (l *Lobby) setRematchTiming(poll, grace time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rematchPoll, l.rematchGrace = poll, grace
}

// rematchOffer is the host's offer on one ended table. Guarded by l.mu.
type rematchOffer struct {
	newID    uuid.UUID
	host     uuid.UUID
	hostName string
	open     bool
	// accepted maps an old seat to its seat at the new table; a value of
	// uuid.Nil is an accept still being seated.
	accepted map[uuid.UUID]uuid.UUID
	declined map[uuid.UUID]bool
	// absentSince is when the watcher first found each unanswered seat
	// without a socket.
	absentSince map[uuid.UUID]time.Time
}

// RematchResult is what a successful offer or accept returns: the new
// table and the caller's seat at it.
type RematchResult struct {
	Game     GameMeta
	PlayerID uuid.UUID
}

// rematchSeatSpec is one human seat to put at the new table.
type rematchSeatSpec struct {
	name      string
	identity  DiscordIdentity
	userID    uuid.UUID
	deckName  string
	deckID    string
	deckPower float64
	cards     []game.Card
	host      bool
}

// humanAsked reports whether the offer asks seat s: a human other than
// the host. A bot is carried over by the lobby and an agent is left out.
func humanAsked(s SeatInfo, host uuid.UUID) bool {
	return !s.IsBot && !s.IsAgent && s.PlayerID != host
}

// rematchHostLocked is who may offer the rematch: the table's host. A
// host who lost passes hosting on, but when every human has lost (the
// bots won) the engine has nobody left to host, and a person against bots
// would never get a rematch. Then the first human seat is the host. The
// client mirrors this rule (lib/rematch.ts). Callers hold l.mu.
func rematchHostLocked(entry *gameEntry) uuid.UUID {
	if h := entry.room.HostPlayerID(); h != uuid.Nil {
		return h
	}
	seats := append([]SeatInfo(nil), entry.meta.Players...)
	sortSeats(seats)
	for _, s := range seats {
		if !s.IsBot && !s.IsAgent {
			return s.PlayerID
		}
	}
	return uuid.Nil
}

// rematchTableName is the new table's name: the old one, once marked.
func rematchTableName(old string) string {
	old = strings.TrimSuffix(strings.TrimSpace(old), rematchSuffix)
	return trimToLimit(old, maxTableNameLen-len(rematchSuffix)) + rematchSuffix
}

// keepRematchDecksLocked stores the decks a table is starting with, and
// arms their release. Callers hold l.mu, before the game leaves the lobby.
func (l *Lobby) keepRematchDecksLocked(entry *gameEntry) {
	if entry.practice {
		return
	}
	entry.rematchDecks = entry.room.Game.LobbyDecks()
}

// armRematchReleaseLocked drops the kept decks rematchKeep after the game
// ended, if no offer has taken them over by then. Callers hold l.mu.
func (l *Lobby) armRematchReleaseLocked(entry *gameEntry) {
	if entry.rematchDecks == nil || entry.rematchTimer != nil {
		return
	}
	entry.rematchTimer = time.AfterFunc(rematchKeep, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if entry.rematch == nil {
			entry.rematchDecks = nil
		}
	})
}

// rematchViewLocked is the offer as the room stamps it, nil for none.
func rematchViewLocked(entry *gameEntry) *protocol.RematchView {
	o := entry.rematch
	if o == nil {
		return nil
	}
	v := &protocol.RematchView{
		GameID:   o.newID.String(),
		Host:     o.host.String(),
		HostName: o.hostName,
		Open:     o.open,
	}
	for id, n := range o.accepted {
		if n != uuid.Nil {
			v.Accepted = append(v.Accepted, id.String())
		}
	}
	for id := range o.declined {
		v.Left = append(v.Left, id.String())
	}
	sort.Strings(v.Accepted)
	sort.Strings(v.Left)
	return v
}

// pushRematchLocked restamps the old table's room with the offer and
// returns the thunk that broadcasts it, to be run after l.mu is released.
func (l *Lobby) pushRematchLocked(entry *gameEntry) (func(), error) {
	return l.applyLocked(entry.meta.ID, entry, func() error {
		entry.room.SetRematch(rematchViewLocked(entry))
		return nil
	})
}

// undecidedLocked lists the seats the offer is still waiting on.
func (o *rematchOffer) undecidedLocked(entry *gameEntry) []uuid.UUID {
	var out []uuid.UUID
	for _, s := range entry.meta.Players {
		if !humanAsked(s, o.host) {
			continue
		}
		if _, ok := o.accepted[s.PlayerID]; ok || o.declined[s.PlayerID] {
			continue
		}
		out = append(out, s.PlayerID)
	}
	return out
}

// closeRematchLocked ends the offer: nobody else is carried over, and the
// kept decks go. Callers hold l.mu.
func (l *Lobby) closeRematchLocked(entry *gameEntry) {
	o := entry.rematch
	if o == nil || !o.open {
		return
	}
	o.open = false
	entry.rematchDecks = nil
	if entry.rematchTimer != nil {
		entry.rematchTimer.Stop()
	}
}

// OfferRematch opens a rematch of the ended table oldID on behalf of
// hostPlayer, which must be its host. It makes the new table, seats the
// host and the old bots at it, and asks the other humans. The result is
// the new table and the host's seat there.
func (l *Lobby) OfferRematch(oldID, hostPlayer uuid.UUID) (RematchResult, error) {
	l.mu.Lock()
	entry, ok := l.games[oldID]
	if !ok {
		l.mu.Unlock()
		return RematchResult{}, ErrGameNotFound
	}
	switch {
	case entry.practice || entry.meta.Archived():
		l.mu.Unlock()
		return RematchResult{}, ErrRematchUnavailable
	case entry.room.Game.CurrentState() != game.StateEnded:
		l.mu.Unlock()
		return RematchResult{}, ErrRematchNotEnded
	case hostPlayer == uuid.Nil || rematchHostLocked(entry) != hostPlayer:
		l.mu.Unlock()
		return RematchResult{}, ErrRematchNotHost
	case entry.rematch != nil:
		l.mu.Unlock()
		return RematchResult{}, ErrRematchOffered
	case entry.rematchDecks == nil:
		l.mu.Unlock()
		return RematchResult{}, ErrRematchUnavailable
	}
	seats := append([]SeatInfo(nil), entry.meta.Players...)
	sortSeats(seats)
	decks := entry.rematchDecks
	settings := entry.room.Game.TableSettingsSnapshot()
	name := rematchTableName(entry.meta.Name)
	var hostSeat *SeatInfo
	for i := range seats {
		if seats[i].PlayerID == hostPlayer {
			hostSeat = &seats[i]
		}
	}
	if hostSeat == nil || len(decks[hostPlayer]) == 0 {
		l.mu.Unlock()
		return RematchResult{}, ErrRematchUnavailable
	}
	// Claimed before l.mu is released, so a second offer is refused while
	// the table is being made.
	offer := &rematchOffer{
		host:        hostPlayer,
		hostName:    hostSeat.Name,
		accepted:    map[uuid.UUID]uuid.UUID{},
		declined:    map[uuid.UUID]bool{},
		absentSince: map[uuid.UUID]time.Time{},
	}
	entry.rematch = offer
	l.mu.Unlock()

	res, err := l.makeRematchTable(name, seats, decks, settings, hostSeat)
	if err != nil {
		l.mu.Lock()
		if entry.rematch == offer {
			entry.rematch = nil
		}
		l.mu.Unlock()
		return RematchResult{}, err
	}

	var broadcast func()
	l.mu.Lock()
	offer.newID = res.Game.ID
	offer.open = len(offer.undecidedLocked(entry)) > 0
	if !offer.open {
		l.closeRematchLocked(entry)
	}
	var perr error
	broadcast, perr = l.pushRematchLocked(entry)
	if offer.open && l.presence != nil {
		go l.watchRematch(entry, offer)
	}
	l.mu.Unlock()
	if perr != nil {
		// The table exists and the host sits at it; only the prompt to
		// the others failed to go out. Not worth undoing.
		slog.Default().Warn("rematch: pushing the offer failed", "game_id", oldID, "err", perr)
	} else {
		broadcast()
	}
	return res, nil
}

// makeRematchTable builds the new table: created for the host's user,
// private, with the old settings, the host and the old bots seated.
func (l *Lobby) makeRematchTable(name string, seats []SeatInfo, decks map[uuid.UUID][]game.Card, settings game.TableSettings, hostSeat *SeatInfo) (RematchResult, error) {
	meta, err := l.CreateWith(name, parseUUIDOrNil(hostSeat.UserID), "")
	if err != nil {
		return RematchResult{}, err
	}
	newID := meta.ID
	fail := func(err error) (RematchResult, error) {
		_ = l.Delete(newID)
		return RematchResult{}, err
	}
	if err := l.makePrivate(newID, meta); err != nil {
		return fail(err)
	}
	// Only a table that was set up differently from a fresh one needs the
	// patch, so a default table gets no "settings changed" log line.
	fresh, err := l.tableSettings(newID)
	if err != nil {
		return fail(err)
	}
	if settings != fresh {
		if _, err := l.UpdateSettings(newID, uuid.Nil, completeSettingsPatch(settings)); err != nil {
			return fail(err)
		}
	}
	hostNew, err := l.seatRematch(newID, seatSpecOf(*hostSeat, decks[hostSeat.PlayerID], true))
	if err != nil {
		return fail(err)
	}
	for _, s := range seats {
		if !s.IsBot {
			continue
		}
		cards := decks[s.PlayerID]
		if len(cards) == 0 {
			continue
		}
		// A bot that cannot be seated (its tier gone, the load cap) is
		// skipped: the host can add another from the new table.
		if s.BotKeyOwner != uuid.Nil {
			// ADR 0182: a bot on someone's own model key comes along
			// only when that someone is the one asking for the
			// rematch. Nobody else may put a seat on their key, and
			// the seat without it would be a model tier this server
			// may not be able to play.
			if s.BotKeyOwner != parseUUIDOrNil(hostSeat.UserID) {
				continue
			}
			if _, _, err := l.AddBotOnKey(newID, s.Name, s.BotTier, s.BotDeck, s.DeckName, cards, s.BotKeyOwner, s.BotKeyModel); err != nil {
				slog.Default().Warn("rematch: seating a bot failed", "game_id", newID, "bot", s.Name, "err", err)
			}
			continue
		}
		if _, _, err := l.AddBot(newID, s.Name, s.BotTier, s.BotDeck, s.DeckName, cards); err != nil {
			slog.Default().Warn("rematch: seating a bot failed", "game_id", newID, "bot", s.Name, "err", err)
		}
	}
	out, err := l.Get(newID)
	if err != nil {
		return fail(err)
	}
	return RematchResult{Game: out, PlayerID: hostNew}, nil
}

// tableSettings reads a table's settings.
func (l *Lobby) tableSettings(id uuid.UUID) (game.TableSettings, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[id]
	if !ok {
		return game.TableSettings{}, ErrGameNotFound
	}
	return e.room.Game.TableSettingsSnapshot(), nil
}

// makePrivate revokes both invites of a table just made and forgets their
// plaintext, so no link to it exists.
func (l *Lobby) makePrivate(id uuid.UUID, meta GameMeta) error {
	now := time.Now().UTC()
	for _, tok := range []string{meta.InviteToken, meta.SpectatorInvite} {
		h, ok := hashInvite(tok)
		if !ok {
			continue
		}
		ctx, cancel := storeCtx()
		err := l.store.RevokeInvite(ctx, h, now)
		cancel()
		if err != nil {
			return fmt.Errorf("revoking a rematch invite: %w", err)
		}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.games[id]; ok {
		e.meta.InviteToken = ""
		e.meta.SpectatorInvite = ""
	}
	return nil
}

// seatSpecOf is the seat s at the old table as a spec for the new one.
func seatSpecOf(s SeatInfo, cards []game.Card, host bool) rematchSeatSpec {
	spec := rematchSeatSpec{
		name:      s.Name,
		userID:    parseUUIDOrNil(s.UserID),
		deckName:  s.DeckName,
		deckID:    s.DeckID,
		deckPower: s.DeckPower,
		cards:     cards,
		host:      host,
	}
	if s.DiscordID != "" {
		spec.identity = DiscordIdentity{ID: s.DiscordID, GlobalName: s.DisplayName, AvatarHash: s.DiscordAvatarHash}
	}
	return spec
}

// seatRematch seats one human at the unstarted new table with their deck.
// It is join without the invite: the table has none, and the caller has
// already been checked against the old table.
func (l *Lobby) seatRematch(id uuid.UUID, spec rematchSeatSpec) (uuid.UUID, error) {
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
		return uuid.Nil, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
		return uuid.Nil, ErrGameStarted
	}
	if len(entry.meta.Players) >= game.MaxPlayers {
		return uuid.Nil, ErrGameFull
	}
	if spec.userID != uuid.Nil {
		if _, taken := seatOfUser(entry.meta.Players, spec.userID); taken {
			return uuid.Nil, ErrAlreadySeated
		}
	}
	var p *game.Player
	var err error
	broadcast, err = l.applyLocked(id, entry, func() error {
		added, addErr := entry.room.Game.AddPlayer(spec.name, spec.cards)
		if addErr != nil {
			return addErr
		}
		p = added
		// AddPlayer installs the deck; ReplaceDeck marks it a real one.
		if err := entry.room.Game.ReplaceDeck(p.ID, spec.cards); err != nil {
			return err
		}
		if spec.host {
			entry.room.SetHost(p.ID)
		}
		l.bindPlaymat(entry.room, p.ID, spec.userID)
		if spec.identity.Populated() {
			_ = entry.room.Game.SetDiscordIdentity(p.ID, spec.identity.ID, spec.identity.AvatarHash, spec.identity.DisplayName())
		}
		return nil
	})
	if err != nil {
		switch err {
		case game.ErrGameFull:
			return uuid.Nil, ErrGameFull
		case game.ErrGameNotInLobby:
			return uuid.Nil, ErrGameStarted
		}
		return uuid.Nil, err
	}
	seat := SeatInfo{
		PlayerID:     p.ID,
		Name:         p.Name,
		Seat:         p.Seat,
		DeckName:     spec.deckName,
		DeckUploaded: true,
		DeckID:       spec.deckID,
		DeckPower:    spec.deckPower,
	}
	if spec.identity.Populated() {
		seat.DiscordID = spec.identity.ID
		seat.DiscordAvatarHash = spec.identity.AvatarHash
		seat.DisplayName = spec.identity.DisplayName()
	}
	if spec.userID != uuid.Nil {
		seat.UserID = spec.userID.String()
	}
	entry.meta.Players = append(entry.meta.Players, seat)
	l.persistSeatsLocked(entry)
	l.syncHostLocked(entry)
	return p.ID, nil
}

// AcceptRematch seats the old seat playerID at the new table with its
// deck. The caller has been checked to hold that seat.
func (l *Lobby) AcceptRematch(oldID, playerID uuid.UUID) (RematchResult, error) {
	l.mu.Lock()
	entry, ok := l.games[oldID]
	if !ok {
		l.mu.Unlock()
		return RematchResult{}, ErrGameNotFound
	}
	o := entry.rematch
	if o == nil || o.newID == uuid.Nil {
		l.mu.Unlock()
		return RematchResult{}, ErrNoRematch
	}
	var seat *SeatInfo
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == playerID {
			seat = &entry.meta.Players[i]
		}
	}
	switch {
	case seat == nil:
		l.mu.Unlock()
		return RematchResult{}, ErrPlayerNotInGame
	case !humanAsked(*seat, o.host):
		l.mu.Unlock()
		return RematchResult{}, ErrRematchNotAsked
	case !o.open:
		l.mu.Unlock()
		return RematchResult{}, ErrRematchClosed
	}
	if _, done := o.accepted[playerID]; done {
		l.mu.Unlock()
		return RematchResult{}, ErrRematchClosed
	}
	cards := entry.rematchDecks[playerID]
	if len(cards) == 0 {
		l.mu.Unlock()
		return RematchResult{}, ErrRematchUnavailable
	}
	spec := seatSpecOf(*seat, cards, false)
	newID := o.newID
	o.accepted[playerID] = uuid.Nil // seating; see rematchOffer.accepted
	l.mu.Unlock()

	newPlayer, err := l.seatRematch(newID, spec)

	var broadcast func()
	l.mu.Lock()
	if entry.rematch != o {
		l.mu.Unlock()
		return RematchResult{}, ErrNoRematch
	}
	if err != nil {
		delete(o.accepted, playerID)
		l.mu.Unlock()
		return RematchResult{}, err
	}
	o.accepted[playerID] = newPlayer
	delete(o.declined, playerID)
	if len(o.undecidedLocked(entry)) == 0 {
		l.closeRematchLocked(entry)
	}
	var perr error
	broadcast, perr = l.pushRematchLocked(entry)
	l.mu.Unlock()
	if perr == nil {
		broadcast()
	}
	out, gerr := l.Get(newID)
	if gerr != nil {
		return RematchResult{}, gerr
	}
	return RematchResult{Game: out, PlayerID: newPlayer}, nil
}

// DeclineRematch records that playerID will not join. It may change its
// mind with AcceptRematch while the offer is open.
func (l *Lobby) DeclineRematch(oldID, playerID uuid.UUID) error {
	var broadcast func()
	l.mu.Lock()
	entry, ok := l.games[oldID]
	if !ok {
		l.mu.Unlock()
		return ErrGameNotFound
	}
	o := entry.rematch
	if o == nil {
		l.mu.Unlock()
		return ErrNoRematch
	}
	var seat *SeatInfo
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == playerID {
			seat = &entry.meta.Players[i]
		}
	}
	switch {
	case seat == nil:
		l.mu.Unlock()
		return ErrPlayerNotInGame
	case !humanAsked(*seat, o.host):
		l.mu.Unlock()
		return ErrRematchNotAsked
	}
	if _, joined := o.accepted[playerID]; joined {
		l.mu.Unlock()
		return ErrRematchAccepted
	}
	if o.declined[playerID] {
		l.mu.Unlock()
		return nil
	}
	o.declined[playerID] = true
	if o.open && len(o.undecidedLocked(entry)) == 0 {
		l.closeRematchLocked(entry)
	}
	broadcast, err := l.pushRematchLocked(entry)
	l.mu.Unlock()
	if err == nil {
		broadcast()
	}
	return err
}

// watchRematch closes the offer once every human it still waits on has
// been off the results screen for rematchGrace. It polls the hub, which
// is the only thing that knows who has a socket, and exits when the offer
// closes or the old table is dropped.
func (l *Lobby) watchRematch(entry *gameEntry, offer *rematchOffer) {
	l.mu.Lock()
	poll, grace, presence := l.rematchPoll, l.rematchGrace, l.presence
	l.mu.Unlock()
	if poll <= 0 {
		poll = defaultRematchPoll
	}
	if grace <= 0 {
		grace = defaultRematchGrace
	}
	if presence == nil {
		return
	}
	tick := time.NewTicker(poll)
	defer tick.Stop()
	for {
		select {
		case <-entry.stop:
			return
		case <-tick.C:
		}
		// The hub is asked outside l.mu: it takes its own lock, and the
		// lobby never holds its own across a call into the hub.
		present := presence.ConnectedSeats(entry.meta.ID)

		var broadcast func()
		l.mu.Lock()
		if entry.rematch != offer || !offer.open {
			l.mu.Unlock()
			return
		}
		now := time.Now()
		waiting := false
		for _, id := range offer.undecidedLocked(entry) {
			if present[id] {
				delete(offer.absentSince, id)
				waiting = true
				continue
			}
			since, seen := offer.absentSince[id]
			if !seen {
				offer.absentSince[id] = now
				waiting = true
				continue
			}
			if now.Sub(since) < grace {
				waiting = true
			}
		}
		if waiting {
			l.mu.Unlock()
			continue
		}
		l.closeRematchLocked(entry)
		b, err := l.pushRematchLocked(entry)
		if err == nil {
			broadcast = b
		}
		l.mu.Unlock()
		if broadcast != nil {
			broadcast()
		}
		return
	}
}
