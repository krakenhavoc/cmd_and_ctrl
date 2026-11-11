package lobby

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/drafthistory"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draft_matches.go is the lobby's half of a draft's rounds (ADR 0222
// §6, #5440): the limited tables its matches are played at, and what
// is kept of a finished draft.
//
// # A match's table
//
// The draft (internal/draft) pairs the rounds and keeps the results;
// it never touches a game. Whenever a draft changes, watchDraft asks
// it which matches of the round in play have a person in them and no
// table yet, and opens one for each: a `limited` table (game/format.go)
// with the two seats already at it, holding the decks they built, each
// deck's sideboard the rest of its pool (CR 100.4b). A seat no person
// holds is a bot seat (ADR 0033) playing the deck its draft seat
// built. The table starts at once, with the opening roll open.
//
// It is made the way the practice table is (practice.go), and for the
// same reasons: it has no invite, so nobody else can join or watch it;
// it is not in GET /games; it is never written to the store, so a
// restart ends it with its draft; and it keeps no games row, so it is
// not in "My games" and writes no statistics record. What is kept of a
// match is its result, in the draft, and for a signed-in person in
// their draft history (drafthistory).
//
// A person reaches their seat with POST /draft/{id}/match, which mints
// the table seat's session from their draft seat's (drafts_http.go).
//
// # The result
//
// watchDraftMatch follows the table. When its match ends, the winner
// and the games each seat won go to the draft, which begins the next
// round when it was the last result of this one. A table nobody
// touches for draftMatchIdle is recorded as a draw and closed, so one
// person walking away does not hold a draft's round open for ever. If
// that was the draft's only match in play and nobody is on the draft's
// own page either, the draft is closed instead: everyone has walked
// away, and the next round would only seat a bot to wait for them. A
// finished table stays up for draftMatchLinger, for the two players
// to look at the board, and is then dropped.
//
// # The load cap
//
// Each open match table counts against ADR 0148's load cap as a table
// with its seats, like any other (entryCostLocked). During the rounds
// the draft itself counts nothing more: the people in it are counted
// at their tables. A round's tables are opened even when the server is
// at its cap, because the draft was admitted when it was made and a
// drafter must not be turned away from the match they drafted for;
// while they are open they count, so new tables and drafts are
// refused.

const (
	// draftMatchIdle is how long a match table may go without a commit
	// before it is recorded as a draw and closed.
	draftMatchIdle = 20 * time.Minute
	// draftMatchMaxAge is the longest a match table lives.
	draftMatchMaxAge = 3 * time.Hour
	// draftMatchLinger is how long a finished match's table stays up.
	draftMatchLinger = 10 * time.Minute
)

// DraftMatchLimits bounds a draft's match tables. A zero field means
// its default. Tests shrink them; production keeps the defaults.
type DraftMatchLimits struct {
	Idle, MaxAge, Linger time.Duration
}

func (d DraftMatchLimits) withDefaults() DraftMatchLimits {
	if d.Idle <= 0 {
		d.Idle = draftMatchIdle
	}
	if d.MaxAge <= 0 {
		d.MaxAge = draftMatchMaxAge
	}
	if d.Linger <= 0 {
		d.Linger = draftMatchLinger
	}
	return d
}

// SetDraftMatchLimits replaces the bounds of the match tables opened
// from now on.
func (l *Lobby) SetDraftMatchLimits(d DraftMatchLimits) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.draftMatchLimits = d
}

// DraftRecorder keeps a signed-in person's finished drafts
// (drafthistory.Store). Nil keeps nothing.
type DraftRecorder interface {
	Save(ctx context.Context, r drafthistory.Record) error
}

// SetDraftRecorder wires the draft history store in after construction.
func (l *Lobby) SetDraftRecorder(r DraftRecorder) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.draftRecorder = r
}

// draftMatchRef ties a limited table to the match of a draft it is.
type draftMatchRef struct {
	draft        uuid.UUID
	round, index int
	// players is the table's two game players and seats the draft
	// seats they are: [0] is the pairing's seat A, [1] its seat B.
	players [2]uuid.UUID
	seats   [2]uuid.UUID
	// numbers is the two draft seats by number.
	numbers [2]int
}

// draftMatchSeat is one seat of a match about to be opened.
type draftMatchSeat struct {
	number int
	id     uuid.UUID
	name   string
	bot    bool
	user   uuid.UUID
	main   []draft.Card
	basics map[string]int
	side   []draft.Card
}

// watchDraft opens the tables a draft's rounds need and keeps the
// draft's history, on every change to it, until the draft closes.
func (l *Lobby) watchDraft(id uuid.UUID, e *draftEntry) {
	for {
		select {
		case <-e.done:
			return
		case <-e.room.Wake():
		}
		l.syncDraft(id, e)
	}
}

// syncDraft opens every table the round in play is waiting for, and
// records the draft once it is finished.
func (l *Lobby) syncDraft(id uuid.UUID, e *draftEntry) {
	e.sync.Lock()
	defer e.sync.Unlock()

	type open struct {
		need  draft.TableNeed
		seats [2]draftMatchSeat
	}
	var opens []open
	var name string
	var bestOf int
	var phase draft.Phase
	e.room.With(func(d *draft.Draft) {
		phase = d.Phase()
		name = d.Config().Name
		bestOf = d.BestOf()
		all := d.Seats()
		for _, n := range d.TablesNeeded() {
			o := open{need: n}
			for i, num := range []int{n.A, n.B} {
				s := all[num]
				seat := draftMatchSeat{number: num, id: s.ID, name: s.Name, bot: s.Bot, user: s.UserID}
				seat.main, seat.basics, seat.side, _ = d.DeckCards(num)
				o.seats[i] = seat
			}
			opens = append(opens, o)
		}
	})
	for _, o := range opens {
		tableID, err := l.openDraftMatch(id, e, name, bestOf, o.need, o.seats)
		if err != nil {
			// A table that cannot be opened must not hold the round: the
			// match is recorded as a draw.
			slog.Default().Error("draft match table could not be opened; recording a draw",
				"draft_id", id, "round", o.need.Round, "err", err)
			_ = e.room.ReportMatch(o.need.Round, o.need.Index, draft.NoSeat, 0, 0)
			continue
		}
		if err := e.room.SetTable(o.need.Round, o.need.Index, tableID); err != nil {
			l.dropDraftMatch(tableID, "draft moved on")
		}
	}
	if phase == draft.PhaseFinished {
		l.recordDraft(e, uuid.Nil)
	}
}

// draftBotTier is the tier a draft's bot seats play at: the heuristic
// bot, which plays whatever deck it holds (ADR 0222 §7), or the first
// tier this server offers when it has no heuristic.
func draftBotTier(host BotHost) string {
	if host == nil {
		return ""
	}
	tiers := host.Tiers()
	for _, t := range tiers {
		if t == StandInTier {
			return t
		}
	}
	if len(tiers) > 0 {
		return tiers[0]
	}
	return ""
}

// basicLandCard is the printing of a basic land a limited deck adds:
// the drafted set's own when it has one, else any, else a plain card
// with the land's type line, which is all a basic land is.
func basicLandCard(idx *cards.Index, name, set string) cards.Card {
	if idx != nil {
		if c, ok := idx.FindByName(name); ok {
			if own, ok := idx.FindPrintingOf(c, uuid.Nil, set, ""); ok {
				return own
			}
			return c
		}
	}
	return cards.Card{
		ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte("c2 basic land "+name)),
		Name:     name,
		TypeLine: "Basic Land — " + name,
		Layout:   "normal",
	}
}

// draftGameCards turns a seat's deck into the cards it brings to the
// table: the deck, then its sideboard, which is the rest of its pool
// (CR 100.4b).
func draftGameCards(idx *cards.Index, set string, s draftMatchSeat) []game.Card {
	out := make([]game.Card, 0, len(s.main)+len(s.side)+17)
	lookup := func(c draft.Card) cards.Card {
		if idx != nil {
			if full, ok := idx.Get(c.ScryfallID); ok {
				return full
			}
		}
		return cards.Card{ID: c.ScryfallID, OracleID: c.OracleID, Name: c.Name, TypeLine: c.TypeLine, ManaCost: c.ManaCost, Colors: c.Colors, Layout: "normal"}
	}
	for _, c := range s.main {
		out = append(out, deck.ToGameCard(lookup(c), false))
	}
	for _, name := range draft.BasicLandNames {
		land := basicLandCard(idx, name, set)
		for range s.basics[name] {
			out = append(out, deck.ToGameCard(land, false))
		}
	}
	for _, c := range s.side {
		card := deck.ToGameCard(lookup(c), false)
		card.IsSideboard = true
		out = append(out, card)
	}
	return out
}

// openDraftMatch opens the limited table of one match and starts it.
func (l *Lobby) openDraftMatch(draftID uuid.UUID, e *draftEntry, draftName string, bestOf int, need draft.TableNeed, seats [2]draftMatchSeat) (uuid.UUID, error) {
	idx := l.cardIndex.Load()
	set := e.set
	var decks [2][]game.Card
	for i := range seats {
		decks[i] = draftGameCards(idx, set, seats[i])
		if len(decks[i]) == 0 {
			return uuid.Nil, fmt.Errorf("lobby: draft seat %d has no deck", seats[i].number)
		}
	}
	names := [2]string{trimToLimit(seats[0].name, 40), trimToLimit(seats[1].name, 40)}
	for i := range names {
		if names[i] == "" {
			names[i] = "Player"
		}
	}
	if strings.EqualFold(names[0], names[1]) {
		// Two guests who both kept the default name.
		names[1] = trimToLimit(names[1], 36) + " (2)"
	}

	var after []func()
	defer func() {
		for _, f := range after {
			f()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.drafts[draftID] != e {
		return uuid.Nil, ErrDraftNotFound
	}
	tier := draftBotTier(l.bots)
	if tier == "" && (seats[0].bot || seats[1].bot) {
		return uuid.Nil, fmt.Errorf("lobby: bot seats are not enabled on this server")
	}

	g := game.NewGame()
	room := l.mgr.CreateEphemeral(g)
	var players [2]*game.Player
	_, _, err := room.ApplyExternal(func() error {
		if err := g.OpenLimited(bestOf); err != nil {
			return err
		}
		host := uuid.Nil
		for i, s := range seats {
			p, err := g.AddPlayer(names[i], decks[i])
			if err != nil {
				return err
			}
			// AddPlayer installs the deck; ReplaceDeck marks it as a
			// real one (DeckImported), as every other seated deck is.
			if err := g.ReplaceDeck(p.ID, decks[i]); err != nil {
				return err
			}
			if s.bot {
				if err := g.SetBot(p.ID, tier, ""); err != nil {
					return err
				}
			} else {
				if host == uuid.Nil {
					host = p.ID
				}
				l.bindPlaymat(room, p.ID, s.user)
			}
			players[i] = p
		}
		if host != uuid.Nil {
			room.SetHost(host)
		}
		return g.StartWithOpeningRoll(nil)
	})
	if err != nil {
		l.mgr.Delete(g.ID)
		return uuid.Nil, err
	}

	ref := &draftMatchRef{draft: draftID, round: need.Round, index: need.Index}
	meta := GameMeta{
		ID:        g.ID,
		Name:      trimToLimit(fmt.Sprintf("%s · Round %d", draftName, need.Round), maxTableNameLen),
		CreatedAt: g.CreatedAt,
		State:     string(game.StateLobby),
		DraftID:   draftID.String(),
	}
	for i, s := range seats {
		p := players[i]
		ref.players[i], ref.seats[i], ref.numbers[i] = p.ID, s.id, s.number
		info := SeatInfo{
			PlayerID:     p.ID,
			Name:         p.Name,
			Seat:         p.Seat,
			DeckName:     "Drafted deck",
			DeckUploaded: true,
		}
		if s.bot {
			info.IsBot, info.BotTier = true, tier
		} else if s.user != uuid.Nil {
			info.UserID = s.user.String()
		}
		meta.Players = append(meta.Players, info)
	}
	entry := &gameEntry{
		meta:         meta,
		room:         room,
		stop:         make(chan struct{}),
		startedKnown: true,
		// Unlisted and never stored, like the practice table; it has no
		// practice owner or human, so the practice routes refuse it.
		practice:   true,
		draftMatch: ref,
	}
	l.games[g.ID] = entry
	l.syncStateLocked(entry)
	l.syncHostLocked(entry)
	l.startWatchLocked(entry)
	go l.watchDraftMatch(entry, l.draftMatchLimits.withDefaults())
	if start := l.botStartLocked(entry); start != nil {
		after = append(after, start)
	}
	return g.ID, nil
}

// watchDraftMatch follows a match's table: it reports the result to
// the draft when the match ends, records a draw for a table nobody
// touches, and drops the table once it is no longer needed.
func (l *Lobby) watchDraftMatch(entry *gameEntry, limits DraftMatchLimits) {
	wake, unsubscribe := entry.room.Subscribe()
	defer unsubscribe()
	idle := time.NewTimer(limits.Idle)
	defer idle.Stop()
	age := time.NewTimer(limits.MaxAge)
	defer age.Stop()
	reported := false
	for {
		if !reported && entry.room.Game.CurrentState() == game.StateEnded {
			reported = true
			l.reportDraftMatch(entry, false)
			// The board stays up for a while, then the table goes.
			idle.Reset(limits.Linger)
		}
		var why string
		select {
		case <-entry.stop:
			return
		case <-wake:
			if !reported {
				idle.Reset(limits.Idle)
			}
			continue
		case <-idle.C:
			why = "idle"
		case <-age.C:
			why = "max_age"
		}
		if !reported {
			if l.draftAbandoned(entry) {
				// Nobody is at this table, at another, or on the draft's
				// page: closing the draft saves what its people drafted
				// and takes this table with it.
				l.closeDraft(entry.draftMatch.draft)
				return
			}
			l.reportDraftMatch(entry, true)
		} else {
			why = "finished"
		}
		l.dropDraftMatch(entry.meta.ID, why)
		return
	}
}

// draftAbandoned reports whether a stalled match's draft has nobody
// left at it: no socket on the draft's own page, and no other match of
// it still in play.
func (l *Lobby) draftAbandoned(entry *gameEntry) bool {
	ref := entry.draftMatch
	if ref == nil {
		return false
	}
	l.mu.Lock()
	e, ok := l.drafts[ref.draft]
	others := false
	for id, g := range l.games {
		if id != entry.meta.ID && g.draftMatch != nil && g.draftMatch.draft == ref.draft && g.meta.State != string(game.StateEnded) {
			others = true
		}
	}
	l.mu.Unlock()
	if !ok || others {
		return false
	}
	return e.room.Status().Connected == 0
}

// reportDraftMatch tells the draft how a match's table ended. abandoned
// records a draw for a table that never finished.
func (l *Lobby) reportDraftMatch(entry *gameEntry, abandoned bool) {
	ref := entry.draftMatch
	if ref == nil {
		return
	}
	e, ok := l.draftEntryOf(ref.draft)
	if !ok {
		return
	}
	winner, winsA, winsB := draft.NoSeat, 0, 0
	if m := entry.room.Game.MatchSnapshot(); m != nil {
		winsA, winsB = m.Wins(ref.players[0]), m.Wins(ref.players[1])
		if !abandoned {
			switch m.Winner {
			case ref.players[0]:
				winner = ref.numbers[0]
			case ref.players[1]:
				winner = ref.numbers[1]
			}
		}
	}
	if err := e.room.ReportMatch(ref.round, ref.index, winner, winsA, winsB); err != nil {
		slog.Default().Warn("draft match result was refused", "draft_id", ref.draft, "round", ref.round, "err", err)
	}
}

// dropDraftMatch closes a match's table: out of the registry and the
// room manager, its bot runner stopped and its sockets closed.
func (l *Lobby) dropDraftMatch(id uuid.UUID, why string) {
	l.mu.Lock()
	entry, ok := l.games[id]
	if !ok || entry.draftMatch == nil {
		l.mu.Unlock()
		return
	}
	after := l.dropPracticeLocked(id, "draft match: "+why)
	l.mu.Unlock()
	after()
}

// dropDraftMatches closes every table of a draft, when the draft
// closes.
func (l *Lobby) dropDraftMatches(draftID uuid.UUID) {
	l.mu.Lock()
	var afters []func()
	for id, entry := range l.games {
		if entry.draftMatch != nil && entry.draftMatch.draft == draftID {
			afters = append(afters, l.dropPracticeLocked(id, "draft match: draft closed"))
		}
	}
	l.mu.Unlock()
	for _, f := range afters {
		f()
	}
}

// draftMatchOf is the live, unfinished table a draft seat is playing
// at, with the game player it is there.
func (l *Lobby) draftMatchOf(draftID, seat uuid.UUID) (entry *gameEntry, player uuid.UUID, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.games {
		ref := e.draftMatch
		if ref == nil || ref.draft != draftID || e.meta.State == string(game.StateEnded) {
			continue
		}
		for i, s := range ref.seats {
			if s == seat {
				return e, ref.players[i], true
			}
		}
	}
	return nil, uuid.Nil, false
}

// DraftMatchSeat is the table a draft seat's match of the round in play
// is played at, and the game player the seat is there. ErrDraftNotFound
// for an unknown draft; ErrDraftNoMatch when the seat has no table to
// sit at now.
func (l *Lobby) DraftMatchSeat(draftID, seat uuid.UUID) (GameMeta, uuid.UUID, error) {
	e, ok := l.draftEntryOf(draftID)
	if !ok {
		return GameMeta{}, uuid.Nil, ErrDraftNotFound
	}
	var table uuid.UUID
	e.room.With(func(d *draft.Draft) {
		num := d.SeatNumber(seat)
		for _, p := range d.Pairings(d.MatchRound()) {
			if !p.Done && !p.Bye() && (p.A == num || p.B == num) {
				table = p.Table
			}
		}
	})
	if table == uuid.Nil {
		return GameMeta{}, uuid.Nil, ErrDraftNoMatch
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[table]
	if !ok || entry.draftMatch == nil {
		return GameMeta{}, uuid.Nil, ErrDraftNoMatch
	}
	for i, s := range entry.draftMatch.seats {
		if s == seat {
			e.room.Touch()
			return copyMeta(entry.meta), entry.draftMatch.players[i], nil
		}
	}
	return GameMeta{}, uuid.Nil, ErrDraftNoMatch
}

// concedeDraftMatch concedes the match a seat is playing, when it
// leaves its draft in the middle of one: nobody is coming back to that
// seat, and its opponent must not be left waiting on it.
func (l *Lobby) concedeDraftMatch(draftID, seat uuid.UUID) {
	entry, player, ok := l.draftMatchOf(draftID, seat)
	if !ok {
		return
	}
	var broadcast func()
	l.mu.Lock()
	if l.games[entry.meta.ID] == entry {
		var err error
		broadcast, err = l.applyLocked(entry.meta.ID, entry, func() error {
			return entry.room.Game.ConcedeMatch(player)
		})
		if err != nil {
			slog.Default().Warn("conceding a draft match for a seat that left failed", "game_id", entry.meta.ID, "err", err)
		}
	}
	l.mu.Unlock()
	if broadcast != nil {
		broadcast()
	}
}

// recordDraft writes what is kept of a draft for its signed-in people
// (ADR 0222 §9): every one of them when only is uuid.Nil, else the one
// seat. A guest's seat and a bot's write nothing, and nothing is
// written before the picks are done: only a finished draft's cards are
// a card pool.
func (l *Lobby) recordDraft(e *draftEntry, only uuid.UUID) {
	l.mu.Lock()
	rec := l.draftRecorder
	l.mu.Unlock()
	if rec == nil {
		return
	}
	var records []drafthistory.Record
	e.room.With(func(d *draft.Draft) {
		if d.Phase() == "" {
			return
		}
		cfg := d.Config()
		for num, s := range d.Seats() {
			if s.Bot || s.UserID == uuid.Nil || (only != uuid.Nil && s.ID != only) {
				continue
			}
			r := drafthistory.Record{
				DraftID: d.ID, UserID: s.UserID, Name: cfg.Name, Set: cfg.Set, SetName: cfg.SetName,
				Mode: string(cfg.Mode), Seats: cfg.Seats, BestOf: d.BestOf(), Seat: num,
				DraftedAt: time.Now().UTC(), Basics: map[string]int{}, Picks: []drafthistory.Pick{},
			}
			inDeck := map[string]bool{}
			if built, ok := d.Deck(num); ok {
				for _, id := range built.Main {
					inDeck[id] = true
				}
				for name, n := range built.Basics {
					if n > 0 {
						r.Basics[name] = n
					}
				}
			}
			for _, c := range d.Pool(s.ID) {
				colors := c.Colors
				if colors == nil {
					colors = []string{}
				}
				r.Picks = append(r.Picks, drafthistory.Pick{
					Name: c.Name, ScryfallID: c.ScryfallID.String(), Set: c.Set, CollectorNo: c.CollectorNo,
					Rarity: c.Rarity, Colors: colors, ManaCost: c.ManaCost, ManaValue: c.ManaValue,
					TypeLine: c.TypeLine, Basic: c.Basic, InDeck: inDeck[c.ID],
				})
			}
			for _, res := range d.Results(num) {
				m := drafthistory.Match{Round: res.Round, Result: res.Result, GamesWon: res.GamesWon, GamesLost: res.GamesLost}
				switch {
				case res.Bye:
					m.OpponentKind = drafthistory.OpponentBye
				case res.OpponentBot:
					m.OpponentKind = drafthistory.OpponentBot
				case res.OpponentUser != uuid.Nil:
					m.OpponentKind, m.OpponentUser = drafthistory.OpponentPerson, res.OpponentUser
				default:
					m.OpponentKind = drafthistory.OpponentGuest
				}
				r.Matches = append(r.Matches, m)
			}
			records = append(records, r)
		}
	})
	for _, r := range records {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := rec.Save(ctx, r); err != nil {
			slog.Default().Error("saving a draft's history failed", "draft_id", r.DraftID, "err", err)
		}
		cancel()
	}
}
