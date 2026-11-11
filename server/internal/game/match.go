package game

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// match.go is the match (ADR 0222 §3): one or three games between the
// same two players at one table.
//
// ONE PATH. Every table of a format that plays matches (FormatSpec.
// Matches) has a Game.Match from Start, and best-of-one is a match of
// one game: its one game ends the match, and the table ends as it did
// before there were matches. A Commander table has no Match at all, and
// nothing in this file runs for it.
//
// BETWEEN GAMES. When a game of a best-of-three ends and the match is
// not decided, the table is StateBetweenGames: the finished game stays
// on the board for everyone to read, no game action is legal (every
// engine door asks for StateActive), and the sideboarding window
// (Match.Between) is open. Each player may move cards between their
// deck and sideboard, the loser of the game chooses who plays first,
// and both ready up or the timer ends it. Then the next game begins in
// place, the way a restarted game does (restart.go, ADR 0201 §2): a
// fresh NewGame adopted by RestoreFrom, so every game-scoped field is
// at its NewGame value.
//
// The rules, from the Comprehensive Rules:
//
//	CR 103.1: "At the start of a game, the players determine which one
//	of them will choose who takes the first turn. In the first game of
//	a match (including a single-game match), the players may use any
//	mutually agreeable method (flipping a coin, rolling dice, etc.) to
//	do so. In a match of several games, the loser of the previous game
//	chooses who takes the first turn. If the previous game was a draw,
//	the player who made the choice in that game makes the choice in
//	this game."
//
//	CR 100.4a: "In constructed play, a sideboard may contain no more
//	than fifteen cards. The four-card limit (see rule 100.2a) applies
//	to the combined deck and sideboard."
//
//	CR 104.3a: "A player can concede the game at any time." Conceding a
//	game of a match is that; conceding the match is ConcedeMatch.
//
// A DECK IS A POOL. Start keeps each seat's registered deck and
// sideboard as pristine cards (MatchDeck). Every game after the first
// is seated from that pool with new instance IDs, so nothing a game
// did to a card (a token copy, a card brought in by a wish, a card an
// opponent owns for now) leaks into the next. Sideboarding moves cards
// inside the pool and never touches the saved deck or the list the
// lobby keeps for a rematch: the registered list is what the next
// match starts from.

// StateBetweenGames is a table between two games of a match: the last
// game is over, the match is not, and the sideboarding window is open.
// CurrentState reports it as StateActive, because to everything outside
// the engine the table is still in play; SideboardingOpen tells the two
// apart.
const StateBetweenGames State = "between_games"

// SideboardingTime is how long the window between two games of a match
// stays open before the next game starts without waiting.
const SideboardingTime = 3 * time.Minute

// sideboardingRestoreGrace is the least time a table restored in the
// middle of sideboarding gives its players to come back.
const sideboardingRestoreGrace = time.Minute

// SettingBestOf is the settings key of the match length.
const SettingBestOf = "best_of"

// OutcomeCauseMatchConceded is the cause of a table's outcome when a
// player conceded the match between two games.
const OutcomeCauseMatchConceded = "match_conceded"

// The two Labels EventMatchGame carries: what the chooser chose.
const (
	MatchChosePlay = "play"
	MatchChoseDraw = "draw"
)

var (
	// ErrBestOfLocked is returned by UpdateSettings when a patch
	// changes the match length after Start. Nothing is applied.
	ErrBestOfLocked = errors.New("game: the match length cannot change once the game has started")
	// ErrBestOfFormat is returned by UpdateSettings when a patch asks
	// for a best-of-three in a format that plays no matches.
	ErrBestOfFormat = errors.New("game: that format does not play matches")
	// ErrSideboardingOpen refuses every action but sideboarding's own
	// while the table is between two games of a match.
	ErrSideboardingOpen = errors.New("game: the table is between games of the match")
	// ErrNotSideboarding is returned by sideboarding's verbs when the
	// table is not between two games of a match.
	ErrNotSideboarding = errors.New("game: the table is not between games of a match")
	// ErrNoMatch is returned by ConcedeMatch at a table with no match
	// in progress.
	ErrNoMatch = errors.New("game: there is no match in progress")
	// ErrSideboardReady is returned when a player who has readied up
	// changes their deck or their choice. They unready first.
	ErrSideboardReady = errors.New("game: you are ready; unready to change your deck")
	// ErrSideboardUnknownCard is returned when a submitted deck names a
	// card that is not in the player's deck or sideboard, or names one
	// twice.
	ErrSideboardUnknownCard = errors.New("game: that card is not in your deck or sideboard")
	// ErrSideboardIllegal wraps the reason a sideboarded deck fails the
	// format's deck check. The deck is left as it was.
	ErrSideboardIllegal = errors.New("game: that deck is not legal")
	// ErrNotMatchChooser is returned when anyone but the player CR
	// 103.1 names tries to choose who plays first.
	ErrNotMatchChooser = errors.New("game: the loser of the previous game chooses who plays first")
)

// Match is a table's match. Plain data: carried by Clone, RestoreFrom
// and the snapshot (additive), where it is embedded as it is, so the
// JSON tags are the snapshot's shape.
type Match struct {
	// BestOf is 1 or 3.
	BestOf int `json:"bestOf"`
	// Games is every finished game, in order.
	Games []MatchGame `json:"games,omitempty"`
	// Decks is each seat's pool, in seat order. Nil in a best-of-one,
	// which never seats a second game.
	Decks []MatchDeck `json:"decks,omitempty"`
	// Chooser is the player who chose who took the first turn of the
	// game in play (CR 103.1), which a drawn game hands the next choice
	// to. uuid.Nil until the first game's starting player is known.
	Chooser uuid.UUID `json:"chooser,omitempty"`
	// Between is the open sideboarding window, nil while a game is in
	// play and once the match is over.
	Between *MatchBetween `json:"between,omitempty"`
	// Over says the match is decided, drawn or conceded.
	Over bool `json:"over,omitempty"`
	// Winner is the player who won the match, uuid.Nil for a draw.
	Winner uuid.UUID `json:"winner,omitempty"`
	// ConcededBy is the player who conceded the match, if one did.
	ConcededBy uuid.UUID `json:"concededBy,omitempty"`
}

// MatchGame is one finished game of a match.
type MatchGame struct {
	// Kind is OutcomeWin or OutcomeDraw, Winner the winner of a win and
	// Cause the outcome's cause.
	Kind   string    `json:"kind"`
	Winner uuid.UUID `json:"winner,omitempty"`
	Cause  string    `json:"cause,omitempty"`
	// StartingSeat is the seat that took the first turn, and Turns the
	// turn number the game ended on.
	StartingSeat int `json:"startingSeat"`
	Turns        int `json:"turns"`
	// Chooser is the player who chose who took the first turn.
	Chooser uuid.UUID `json:"chooser,omitempty"`
}

// MatchDeck is one seat's pool: the cards it registered, split the way
// the seat has them now. The cards are pristine (never in a game) and
// their instance IDs are the pool's own, stable for the whole match;
// each game is seated from copies with new IDs.
//
// A pool card is kept as the snapshot's card record, not as a Card: it
// is plain data, so the match adds no route from Game to a closure
// (the ADR 0041 ratchet), and a card comes out of the pool the way it
// comes out of a restore point, its abilities read from the catalog.
type MatchDeck struct {
	Player uuid.UUID `json:"player"`
	// Main is the deck and Side the sideboard, as the seat has them for
	// the next game.
	Main []cardSnapshot `json:"main,omitempty"`
	Side []cardSnapshot `json:"side,omitempty"`
	// Fixed is what sideboarding never moves: a card that begins the
	// game in the command zone, in a format that has one.
	Fixed []cardSnapshot `json:"fixed,omitempty"`
	// Registered is the pool IDs of the registered deck, sorted, so the
	// screen can offer the way back and the engine can tell a deck
	// that was changed from one that was not.
	Registered []uuid.UUID `json:"registered,omitempty"`
}

// poolCards is the pool's record of cards that are in no game: nobody
// knows them but the screen that shows them to their owner.
func poolCards(cs []Card) []cardSnapshot {
	if len(cs) == 0 {
		return nil
	}
	var cen ContinuationCensus
	out := make([]cardSnapshot, len(cs))
	for i, c := range cs {
		c = cloneCard(c)
		c.KnownBy = nil
		out[i] = snapshotCard(c, &cen)
	}
	return out
}

// cardsOfPool is the pool's cards as cards, each a new value.
func cardsOfPool(cs []cardSnapshot) []Card {
	out := make([]Card, len(cs))
	for i := range cs {
		out[i] = restoreCard(&cs[i])
		out[i].stampPrinted()
	}
	return out
}

// MatchPoolLocked is a player's deck and sideboard as they stand for
// the next game, for the screen that shows them to that player. The
// cards carry their pool IDs, which SetMatchDeck takes. Nil at a table
// that keeps no pools. Caller must hold g.mu (read or write).
func (g *Game) MatchPoolLocked(player uuid.UUID) (main, side []Card, registered []uuid.UUID) {
	if g.Match == nil {
		return nil, nil, nil
	}
	d := g.Match.deck(player)
	if d == nil {
		return nil, nil, nil
	}
	return cardsOfPool(d.Main), cardsOfPool(d.Side), append([]uuid.UUID(nil), d.Registered...)
}

// MatchBetween is the open sideboarding window.
type MatchBetween struct {
	// Deadline is when the next game starts whoever is ready.
	Deadline time.Time `json:"deadline"`
	// Ready is the players who have readied up. A bot seat is always
	// ready and is not listed.
	Ready []uuid.UUID `json:"ready,omitempty"`
	// Chooser chooses who plays first (CR 103.1), and PlayFirst is the
	// choice as it stands: true, the default, is the chooser.
	Chooser   uuid.UUID `json:"chooser"`
	PlayFirst bool      `json:"playFirst"`
}

// EffectiveBestOf is the table's match length: 3 when the setting says
// so and the format plays matches, 1 otherwise.
func (s TableSettings) EffectiveBestOf() int {
	if s.BestOf == 3 && s.FormatSpec().Matches {
		return 3
	}
	return 1
}

// validateBestOf range-checks a patch's match length.
func validateBestOf(n int) error {
	if n != 1 && n != 3 {
		return fmt.Errorf("%w: %s %d (want 1 or 3)", ErrInvalidSetting, SettingBestOf, n)
	}
	return nil
}

// storedBestOf is how TableSettings keeps n: best-of-one as the zero
// value, so only a best-of-three table's snapshot has the field.
func storedBestOf(n int) int {
	if n == 3 {
		return 3
	}
	return 0
}

// Wins is how many games of the match player has won.
func (m *Match) Wins(player uuid.UUID) int {
	n := 0
	for _, g := range m.Games {
		if g.Kind == OutcomeWin && g.Winner == player {
			n++
		}
	}
	return n
}

// GameNumber is the number of the game in play, or of the last game
// played while the table is between games or the match is over.
func (m *Match) GameNumber() int {
	if m.Between != nil || m.Over {
		return len(m.Games)
	}
	return len(m.Games) + 1
}

func (m *Match) deck(player uuid.UUID) *MatchDeck {
	for i := range m.Decks {
		if m.Decks[i].Player == player {
			return &m.Decks[i]
		}
	}
	return nil
}

func cloneMatch(m *Match) *Match {
	if m == nil {
		return nil
	}
	out := *m
	out.Games = append([]MatchGame(nil), m.Games...)
	if m.Decks != nil {
		out.Decks = make([]MatchDeck, len(m.Decks))
		for i, d := range m.Decks {
			// A pool card is never written after it is recorded, only
			// moved between the two lists, so the lists are copied and
			// the records shared.
			out.Decks[i] = MatchDeck{
				Player:     d.Player,
				Main:       append([]cardSnapshot(nil), d.Main...),
				Side:       append([]cardSnapshot(nil), d.Side...),
				Fixed:      append([]cardSnapshot(nil), d.Fixed...),
				Registered: append([]uuid.UUID(nil), d.Registered...),
			}
		}
	}
	if m.Between != nil {
		b := *m.Between
		b.Ready = append([]uuid.UUID(nil), m.Between.Ready...)
		out.Between = &b
	}
	return &out
}

// openMatchLocked gives a table of a format that plays matches its
// Match, as Start runs and before anything is shuffled. A best-of-three
// keeps each seat's registered pool; a best-of-one needs none.
//
// Caller must hold g.mu (write).
func (g *Game) openMatchLocked() {
	if !g.Settings.FormatSpec().Matches {
		return
	}
	m := &Match{BestOf: g.Settings.EffectiveBestOf()}
	if m.BestOf > 1 {
		for _, p := range g.Seats {
			if p == nil {
				continue
			}
			d := MatchDeck{Player: p.ID}
			d.Fixed = poolCards(p.Command.Cards)
			d.Main = poolCards(p.Library.Cards)
			if p.Companion != nil {
				d.Side = append(d.Side, poolCards(p.Companion.Cards)...)
			}
			d.Side = append(d.Side, poolCards(p.SideboardCards())...)
			for _, c := range d.Main {
				d.Registered = append(d.Registered, c.InstanceID)
			}
			sortIDs(d.Registered)
			m.Decks = append(m.Decks, d)
		}
	}
	g.Match = m
}

func sortIDs(ids []uuid.UUID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
}

// noteMatchChooserLocked records who chose the starting player of the
// first game: the winner of the opening roll, or the starting player
// at a table that rolled none. Later games' choosers are set as their
// game begins.
//
// Caller must hold g.mu (write).
func (g *Game) noteMatchChooserLocked(chooser uuid.UUID) {
	if g.Match != nil && g.Match.Chooser == uuid.Nil {
		g.Match.Chooser = chooser
	}
}

// matchGameEndedLocked is the match's half of endGameLocked: the game
// that just ended is recorded, and the match is over, or the table goes
// between games. It never begins the next game: the caller is deep in
// whatever ended this one, and holds its players.
//
// Caller must hold g.mu (write).
func (g *Game) matchGameEndedLocked(o GameOutcome) {
	m := g.Match
	if m == nil || m.Over {
		return
	}
	m.Games = append(m.Games, MatchGame{
		Kind:         o.Kind,
		Winner:       o.Winner,
		Cause:        o.Cause,
		StartingSeat: g.StartingSeat,
		Turns:        g.Turn.Round,
		Chooser:      m.Chooser,
	})
	if m.ConcededBy != uuid.Nil {
		m.Over = true
		m.Winner = g.matchOpponentLocked(m.ConcededBy)
		return
	}
	if winner, over := g.matchDecidedLocked(); over {
		m.Over = true
		m.Winner = winner
		return
	}
	// CR 103.1: the loser of this game chooses who takes the first
	// turn of the next; after a draw, whoever chose for this one.
	chooser := m.Chooser
	if o.Kind == OutcomeWin {
		chooser = g.matchOpponentLocked(o.Winner)
	}
	if chooser == uuid.Nil {
		chooser = g.Seats[g.StartingSeat].ID
	}
	m.Between = &MatchBetween{
		Deadline:  time.Now().UTC().Add(SideboardingTime),
		Chooser:   chooser,
		PlayFirst: true,
	}
	g.State = StateBetweenGames
}

// matchDecidedLocked reports whether the match is over on its games,
// and who won it (uuid.Nil for a drawn match). A match is won by the
// first player to win most of its games, and is over once it has
// played them all.
func (g *Game) matchDecidedLocked() (uuid.UUID, bool) {
	m := g.Match
	need := m.BestOf/2 + 1
	best, bestWins, tied := uuid.Nil, 0, false
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		switch w := m.Wins(p.ID); {
		case w >= need:
			return p.ID, true
		case w > bestWins:
			best, bestWins, tied = p.ID, w, false
		case w == bestWins:
			tied = true
		}
	}
	if len(m.Games) < m.BestOf {
		return uuid.Nil, false
	}
	if tied {
		return uuid.Nil, true
	}
	return best, true
}

// matchOpponentLocked is the first seated player who is not player.
func (g *Game) matchOpponentLocked(player uuid.UUID) uuid.UUID {
	for _, p := range g.Seats {
		if p != nil && p.ID != player {
			return p.ID
		}
	}
	return uuid.Nil
}

// SideboardingOpen reports whether the table is between two games of a
// match. Takes the read lock.
func (g *Game) SideboardingOpen() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.State == StateBetweenGames
}

// MatchGameNumber is the number of the match's game in play (or last
// played), 0 at a table with no match. Takes the read lock.
func (g *Game) MatchGameNumber() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Match == nil {
		return 0
	}
	return g.Match.GameNumber()
}

// MatchMark is where the table is in its match: the number of the game
// in play (or last played) and whether the table is between games. The
// room compares it across a commit, because a commit that moves it has
// ended a game of the match or begun the next, and neither is taken
// back. (0, false) at a table with no match. Takes the read lock.
func (g *Game) MatchMark() (number int, between bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Match == nil {
		return 0, false
	}
	return g.Match.GameNumber(), g.State == StateBetweenGames
}

// MatchSnapshot is a copy of the table's match, nil when it has none.
// Takes the read lock.
func (g *Game) MatchSnapshot() *Match {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return cloneMatch(g.Match)
}

// SideboardingDeadline is when the open sideboarding window ends, and
// false when none is open. A window everyone at is ready for ends now.
// Takes the read lock.
func (g *Game) SideboardingDeadline() (time.Time, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.State != StateBetweenGames || g.Match == nil || g.Match.Between == nil {
		return time.Time{}, false
	}
	if g.allReadyLocked() {
		return time.Time{}, true
	}
	return g.Match.Between.Deadline, true
}

// ExtendSideboarding gives an open window at least the restore grace
// from now: a table restored in the middle of sideboarding waits for
// its players to come back rather than starting on a deadline that
// passed while the server was down.
func (g *Game) ExtendSideboarding(now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateBetweenGames || g.Match == nil || g.Match.Between == nil {
		return
	}
	if least := now.UTC().Add(sideboardingRestoreGrace); g.Match.Between.Deadline.Before(least) {
		g.Match.Between.Deadline = least
	}
}

// sideboardReadyLocked reports whether p is ready for the next game. A
// bot seat does not sideboard and is always ready (ADR 0222 §3), and so
// is an agent seat, whose client has no sideboarding verb.
func (g *Game) sideboardReadyLocked(p *Player) bool {
	if p.IsBot || p.Agent {
		return true
	}
	for _, id := range g.Match.Between.Ready {
		if id == p.ID {
			return true
		}
	}
	return false
}

func (g *Game) allReadyLocked() bool {
	for _, p := range g.Seats {
		if p != nil && !g.sideboardReadyLocked(p) {
			return false
		}
	}
	return true
}

// betweenGamesLocked is the open window and the caller's seat, or the
// error a sideboarding verb answers with.
func (g *Game) betweenGamesLocked(playerID uuid.UUID) (*MatchBetween, *Player, error) {
	if g.State != StateBetweenGames || g.Match == nil || g.Match.Between == nil {
		return nil, nil, ErrNotSideboarding
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, nil, ErrPlayerNotFound
	}
	return g.Match.Between, p, nil
}

// SetMatchDeck is a player's sideboarding: main names, by pool ID, the
// cards of their pool that are their deck for the next game, and every
// other card of the pool is their sideboard. The format's deck check
// runs on the result (CheckMatchDeck), and a deck that fails it is
// refused with ErrSideboardIllegal and changes nothing.
func (g *Game) SetMatchDeck(playerID uuid.UUID, main []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, p, err := g.betweenGamesLocked(playerID)
	if err != nil {
		return err
	}
	if g.sideboardReadyLocked(p) {
		return ErrSideboardReady
	}
	d := g.Match.deck(playerID)
	if d == nil {
		return ErrPlayerNotFound
	}
	want := make(map[uuid.UUID]bool, len(main))
	for _, id := range main {
		if want[id] {
			return ErrSideboardUnknownCard
		}
		want[id] = true
	}
	var newMain, newSide []cardSnapshot
	found := 0
	for _, c := range append(append([]cardSnapshot(nil), d.Main...), d.Side...) {
		if want[c.InstanceID] {
			found++
			newMain = append(newMain, c)
		} else {
			newSide = append(newSide, c)
		}
	}
	if found != len(want) {
		return ErrSideboardUnknownCard
	}
	spec := g.Settings.FormatSpec()
	// A name the registered pool holds more of than the format's limit
	// was accepted by the deck check at registration because the card's
	// own text allows it ("A deck can have any number of cards named
	// …"). The engine holds no rules text, so it reads that from the
	// pool, which no card joins or leaves between games.
	ownText := map[string]bool{}
	copies := map[string]int{}
	for _, c := range append(append([]cardSnapshot(nil), newMain...), newSide...) {
		if copies[c.Name]++; copies[c.Name] > spec.Deck.MaxCopies {
			ownText[c.Name] = true
		}
	}
	// The check reads a card's name and type line, which the pool's
	// record holds as they are.
	counted := func(cs []cardSnapshot) []Card {
		out := make([]Card, len(cs))
		for i, c := range cs {
			out[i] = Card{Name: c.Name, TypeLine: c.TypeLine}
		}
		return out
	}
	if err := CheckMatchDeck(spec, counted(newMain), counted(newSide), len(d.Fixed), ownText); err != nil {
		return err
	}
	d.Main, d.Side = newMain, newSide
	return nil
}

// CheckMatchDeck is the format's deck check on a sideboarded deck: the
// numbers of the format's DeckRules row, the same row the deck checker
// (internal/deck) reads when the deck is registered.
//
//   - The deck's size (CR 100.2a), fixed counting the cards sideboarding
//     never moves.
//   - The sideboard's size (CR 100.4a).
//   - The copies of one English name, counted across the deck and the
//     sideboard together (CR 100.4a), basic lands aside (CR 100.2a).
//
// ownText names the cards whose own text sets their copy limit, which
// the count leaves alone. The pool is the registered one, so its card
// legality is not asked again: no card joins or leaves a pool between
// games.
//
// The error wraps ErrSideboardIllegal and says what is wrong, in words
// for the player.
func CheckMatchDeck(spec FormatSpec, main, side []Card, fixed int, ownText map[string]bool) error {
	rules := spec.Deck
	var problems []string
	total := fixed + len(main)
	switch {
	case rules.ExactCards && total != rules.MinCards:
		problems = append(problems, fmt.Sprintf("the deck has %d cards; a %s deck has exactly %d", total, spec.Name, rules.MinCards))
	case total < rules.MinCards:
		problems = append(problems, fmt.Sprintf("the deck has %d cards; a %s deck needs at least %d", total, spec.Name, rules.MinCards))
	}
	if len(side) > rules.SideboardSize {
		problems = append(problems, fmt.Sprintf("the sideboard has %d cards; a %s sideboard can have at most %d", len(side), spec.Name, rules.SideboardSize))
	}
	if rules.MaxCopies > 0 {
		copies := map[string]int{}
		var names []string
		for _, c := range append(append([]Card(nil), main...), side...) {
			if isBasicLandCard(c) || ownText[c.Name] {
				continue
			}
			if copies[c.Name] == 0 {
				names = append(names, c.Name)
			}
			copies[c.Name]++
		}
		sort.Strings(names)
		for _, name := range names {
			if n := copies[name]; n > rules.MaxCopies {
				problems = append(problems, fmt.Sprintf("%q appears %d times across the deck and sideboard; a %s deck can have at most %d", name, n, spec.Name, rules.MaxCopies))
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrSideboardIllegal, strings.Join(problems, "; "))
}

// isBasicLandCard is the deck checker's test for a basic land (CR
// 100.2a), on a game card.
func isBasicLandCard(c Card) bool {
	return strings.Contains(c.TypeLine, "Basic Land")
}

// SetSideboardReady readies a player up for the next game, or takes it
// back. The next game begins when every player is ready.
func (g *Game) SetSideboardReady(playerID uuid.UUID, ready bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	b, p, err := g.betweenGamesLocked(playerID)
	if err != nil {
		return err
	}
	kept := b.Ready[:0]
	for _, id := range b.Ready {
		if id != p.ID {
			kept = append(kept, id)
		}
	}
	b.Ready = kept
	if ready {
		b.Ready = append(b.Ready, p.ID)
	}
	if g.allReadyLocked() {
		g.beginNextMatchGameLocked()
	}
	return nil
}

// ChoosePlayFirst is CR 103.1's choice for the next game: the chooser
// (the loser of the previous game) says whether they take the first
// turn. It stands as chosen when the window closes; unchosen, the
// chooser plays first.
func (g *Game) ChoosePlayFirst(playerID uuid.UUID, playFirst bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	b, p, err := g.betweenGamesLocked(playerID)
	if err != nil {
		return err
	}
	if b.Chooser != p.ID {
		return ErrNotMatchChooser
	}
	if g.sideboardReadyLocked(p) {
		return ErrSideboardReady
	}
	b.PlayFirst = playFirst
	return nil
}

// AdvanceMatch begins the next game of the match when the sideboarding
// window is done: everyone is ready, or now is at or past the deadline.
// Reports whether it did. The table's timer calls it; so does anything
// that drives a table with no timer (the arena, a test).
func (g *Game) AdvanceMatch(now time.Time) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateBetweenGames || g.Match == nil || g.Match.Between == nil {
		return false
	}
	if !g.allReadyLocked() && now.Before(g.Match.Between.Deadline) {
		return false
	}
	g.beginNextMatchGameLocked()
	return true
}

// ConcedeMatch concedes the whole match (ADR 0222 §3), which Concede,
// a concession of the game in play (CR 104.3a), does not. During a
// game it is that concession and ends the match with it; between games
// it ends the match where it stands. Either way the opponent wins the
// match whatever the score.
func (g *Game) ConcedeMatch(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	m := g.Match
	if m == nil || m.Over {
		return ErrNoMatch
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	switch g.State {
	case StateActive:
		if p.Eliminated {
			return ErrPlayerEliminated
		}
		m.ConcededBy = p.ID
		g.concedeLocked(p)
		return nil
	case StateBetweenGames:
		g.concedeMatchBetweenGamesLocked(p)
		return nil
	}
	return ErrGameNotActive
}

// concedeMatchBetweenGamesLocked ends the match for a player who
// concedes it with no game in play. The table's outcome names the
// opponent the winner, with the cause that says why.
func (g *Game) concedeMatchBetweenGamesLocked(p *Player) {
	m := g.Match
	winner := g.matchOpponentLocked(p.ID)
	m.ConcededBy = p.ID
	m.Over = true
	m.Winner = winner
	m.Between = nil
	g.State = StateEnded
	g.Outcome = &GameOutcome{Kind: OutcomeWin, Winner: winner, Cause: OutcomeCauseMatchConceded}
	g.EmitEvent(Event{Kind: EventConcede, Actor: p.ID})
	g.EmitEvent(Event{Kind: EventGameOver, Actor: winner, Label: OutcomeCauseMatchConceded})
}

// markMatchConcededLocked makes a concession somebody else makes for a
// seat (an admin removing it, the table moving on without a player who
// is away) a concession of the match: that player is not coming back
// for the next game.
//
// Caller must hold g.mu (write).
func (g *Game) markMatchConcededLocked(p *Player) {
	if g.Match != nil && !g.Match.Over && g.Match.ConcededBy == uuid.Nil {
		g.Match.ConcededBy = p.ID
	}
}

// matchResultLocked is the result of a finished best-of-three: the
// match's, not its last game's. ok is false at any other table, which
// reads its result from Game.Outcome as it always did.
func (g *Game) matchResultLocked() (kind string, winner uuid.UUID, ok bool) {
	m := g.Match
	if m == nil || m.BestOf <= 1 || !m.Over || g.State != StateEnded {
		return "", uuid.Nil, false
	}
	if m.Winner == uuid.Nil {
		return OutcomeDraw, uuid.Nil, true
	}
	return OutcomeWin, m.Winner, true
}

// beginNextMatchGameLocked closes the sideboarding window and begins
// the next game of the match in place: the same seats, each deck seated
// from its pool as the player left it, fresh life totals and libraries,
// and the starting player CR 103.1's chooser named.
//
// Caller must hold g.mu (write), at the top of a public verb: nothing
// up the stack may hold a *Player of the finished game.
func (g *Game) beginNextMatchGameLocked() {
	m, b := g.Match, g.Match.Between
	next := cloneMatch(m)
	next.Between = nil
	next.Chooser = b.Chooser

	ng := g.newGameAtThisTableLocked(matchRNGKey(g.rngKey))
	ng.Match = next
	ng.Seats = make([]*Player, len(g.Seats))
	start := 0
	for i, p := range g.Seats {
		if p == nil {
			continue
		}
		np := restartedPlayer(p)
		// The outside-the-game containers are seated again from the
		// pool: sideboarding may have moved cards across them.
		np.Companion, np.Sideboard, np.HasSideboard = nil, nil, false
		np.StickerSheetPool = nil
		for j, c := range next.deck(p.ID).gameCards() {
			c.Owner, c.Controller = np.ID, np.ID
			ng.setDeckSourceOrdinalLocked(c.InstanceID, np.Seat, j)
			seatDeckCard(np, c)
		}
		ng.Seats[i] = np
		if (p.ID == b.Chooser) == b.PlayFirst {
			start = i
		}
	}

	g.RestoreFrom(ng)
	g.resetTransientsForNewGameLocked()
	if g.Settings.Planechase {
		if err := g.setUpPlanechaseLocked(); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, Actor: b.Chooser, ErrorMsg: err.Error()})
		}
	}
	g.beginPregameLocked()
	choice := MatchChosePlay
	if !b.PlayFirst {
		choice = MatchChoseDraw
	}
	g.EmitEvent(Event{
		Kind:   EventMatchGame,
		Actor:  b.Chooser,
		Target: g.Seats[start].ID,
		Amount: next.GameNumber(),
		Label:  choice,
	})
	g.dealAndBeginLocked(start)
}

// gameCards is the deck as one game seats it: copies of the pool's
// cards with new instance IDs (the event log is kept across the games
// of a match, and a reused ID would let one game's history speak for a
// card in the next, ADR 0201 §4), the sideboard stamped as it is at a
// table's first seating.
//
// A companion is revealed only by the deck it was registered with: its
// deck-building condition was checked against that list, so a deck
// changed between games keeps the card as a plain sideboard card. That
// is weaker than the rules allow and never stronger.
func (d *MatchDeck) gameCards() []Card {
	if d == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(d.Main))
	for _, c := range d.Main {
		ids = append(ids, c.InstanceID)
	}
	sortIDs(ids)
	registered := len(ids) == len(d.Registered)
	for i := 0; registered && i < len(ids); i++ {
		registered = ids[i] == d.Registered[i]
	}
	out := make([]Card, 0, len(d.Fixed)+len(d.Main)+len(d.Side))
	add := func(c Card, side bool) {
		c.InstanceID = uuid.New()
		switch {
		case !side:
			c.IsSideboard, c.IsCompanion = false, false
		case c.IsCompanion && registered:
			c.IsSideboard = false
		default:
			c.IsSideboard, c.IsCompanion = true, false
		}
		out = append(out, c)
	}
	for _, c := range cardsOfPool(d.Fixed) {
		c.InstanceID = uuid.New()
		out = append(out, c)
	}
	for _, c := range cardsOfPool(d.Main) {
		add(c, false)
	}
	for _, c := range cardsOfPool(d.Side) {
		add(c, true)
	}
	return out
}

// matchRNGKey derives the next game's RNG key from this one's, so its
// shuffles are not the last game's streams replayed from turn 0.
func matchRNGKey(old [32]byte) [32]byte {
	h := sha256.New()
	h.Write(old[:])
	h.Write([]byte("cmdctrl/next-game-of-the-match"))
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// SetSideboardingDeadlineForTest moves the open sideboarding window's
// deadline, so a test of the table's timer does not wait out
// SideboardingTime. A no-op when no window is open.
func (g *Game) SetSideboardingDeadlineForTest(deadline time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State == StateBetweenGames && g.Match != nil && g.Match.Between != nil {
		g.Match.Between.Deadline = deadline
	}
}
