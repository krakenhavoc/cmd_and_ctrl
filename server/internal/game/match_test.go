package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/google/uuid"
)

// match_test.go: matches, best-of-three with sideboarding (ADR 0222 §3).

// buildMatchDeck is a constructed deck with a sideboard: sixty cards
// and fifteen more stamped for the sideboard, named by owner so a test
// can find a card it moved.
func buildMatchDeck(owner string) []Card {
	deck := make([]Card, 0, 75)
	for i := 0; i < 60; i++ {
		deck = append(deck, NewCard(fmt.Sprintf("%s Main %d", owner, i+1), uuid.Nil))
	}
	for i := 0; i < 15; i++ {
		c := NewCard(fmt.Sprintf("%s Side %d", owner, i+1), uuid.Nil)
		c.IsSideboard = true
		deck = append(deck, c)
	}
	return deck
}

// matchTable is a started Standard table of the given length, Alice in
// seat 0 (the starting player of game 1) and Bob in seat 1.
func matchTable(t *testing.T, bestOf int) (g *Game, alice, bob *Player) {
	t.Helper()
	g = NewGame()
	setFormat(t, g, FormatStandard)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &bestOf}); err != nil {
		t.Fatalf("UpdateSettings(best_of %d): %v", bestOf, err)
	}
	for _, name := range []string{"Alice", "Bob"} {
		if _, err := g.AddPlayer(name, buildMatchDeck(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 9))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, g.Seats[0], g.Seats[1]
}

// lose makes the player lose the game in play by conceding it.
func lose(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	if err := g.Concede(id); err != nil {
		t.Fatalf("Concede: %v", err)
	}
}

// readyUp readies both players, which begins the next game.
func readyUp(t *testing.T, g *Game, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if err := g.SetSideboardReady(id, true); err != nil {
			t.Fatalf("SetSideboardReady: %v", err)
		}
	}
}

func score(g *Game, a, b uuid.UUID) string {
	return fmt.Sprintf("%d-%d", g.Match.Wins(a), g.Match.Wins(b))
}

// assertFreshGame checks the table is in a new game of the match: both
// seats in it, at the format's starting life, sixty cards between hand
// and library, the mulligan window open and nothing left of the last
// game on the board.
func assertFreshGame(t *testing.T, g *Game, number int) {
	t.Helper()
	if g.State != StateActive || !g.MulligansOpen {
		t.Fatalf("state %q, mulligans open %v; want an active game in its mulligan window", g.State, g.MulligansOpen)
	}
	if got := g.Match.GameNumber(); got != number {
		t.Errorf("game number = %d, want %d", got, number)
	}
	if g.Outcome != nil {
		t.Errorf("the new game carries an outcome: %+v", g.Outcome)
	}
	if g.Match.Between != nil {
		t.Error("the sideboarding window is still open")
	}
	for _, p := range g.Seats {
		if p.Eliminated {
			t.Errorf("%s is still out of the game", p.Name)
		}
		if p.Life != 20 {
			t.Errorf("%s is at %d life, want 20", p.Name, p.Life)
		}
		if p.Hand.Size() != OpeningHandSize || p.Hand.Size()+p.Library.Size() != 60 {
			t.Errorf("%s has %d in hand and %d in library, want 7 and 53", p.Name, p.Hand.Size(), p.Library.Size())
		}
		if n := len(p.SideboardCards()); n != 15 {
			t.Errorf("%s has %d sideboard cards, want 15", p.Name, n)
		}
		if p.Graveyard.Size() != 0 {
			t.Errorf("%s has a graveyard from the last game", p.Name)
		}
	}
	if g.Battlefield.Size() != 0 || g.Exile.Size() != 0 || g.Stack.Size() != 0 {
		t.Error("the board is not empty")
	}
}

// A best-of-three that goes 2-0: two games, one finished table.
func TestBestOfThreeTwoNil(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	if g.Match == nil || g.Match.BestOf != 3 || len(g.Match.Decks) != 2 {
		t.Fatalf("match = %+v, want a best-of-three with two pools", g.Match)
	}
	firstGameIDs := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		for _, z := range []*Zone{p.Hand, p.Library, p.Sideboard} {
			for _, c := range z.Cards {
				firstGameIDs[c.InstanceID] = true
			}
		}
	}

	lose(t, g, bob.ID)
	if g.State != StateBetweenGames {
		t.Fatalf("after game 1 the state is %q, want between games", g.State)
	}
	if got := g.CurrentState(); got != StateActive {
		t.Errorf("CurrentState = %q between games; the table is in play to the lobby", got)
	}
	if _, ok := g.WinnerSeat(); ok {
		t.Error("a table between games names a winner")
	}
	if got := g.OutcomeKind(); got != "" {
		t.Errorf("a table between games has outcome %q", got)
	}
	if got := score(g, alice.ID, bob.ID); got != "1-0" {
		t.Errorf("score = %s, want 1-0", got)
	}
	// No game action is legal on the finished board.
	if err := g.Concede(alice.ID); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("Concede between games: %v, want ErrGameNotActive", err)
	}
	if err := g.KeepHand(alice.ID); err == nil {
		t.Error("KeepHand between games was accepted")
	}

	readyUp(t, g, alice.ID)
	if g.State != StateBetweenGames {
		t.Fatal("the next game began with one player ready")
	}
	readyUp(t, g, bob.ID)
	assertFreshGame(t, g, 2)
	alice, bob = g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		for _, z := range []*Zone{p.Hand, p.Library, p.Sideboard} {
			for _, c := range z.Cards {
				if firstGameIDs[c.InstanceID] {
					t.Fatalf("%s's %s is the same object it was in game 1", p.Name, c.Name)
				}
				if c.Owner != p.ID {
					t.Fatalf("%s holds a card owned by %s", p.Name, c.Owner)
				}
			}
		}
	}

	lose(t, g, bob.ID)
	if g.State != StateEnded {
		t.Fatalf("after 2-0 the state is %q, want ended", g.State)
	}
	if !g.Match.Over || g.Match.Winner != alice.ID || len(g.Match.Games) != 2 {
		t.Errorf("match = over %v, winner %v, %d games; want Alice in two", g.Match.Over, g.Match.Winner, len(g.Match.Games))
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != alice.Seat {
		t.Errorf("WinnerSeat = %d, %v; want Alice's", seat, ok)
	}
	if got := g.OutcomeKind(); got != OutcomeWin {
		t.Errorf("OutcomeKind = %q, want win", got)
	}
	if err := g.SetSideboardReady(alice.ID, true); !errors.Is(err, ErrNotSideboarding) {
		t.Errorf("readying up after the match: %v, want ErrNotSideboarding", err)
	}
}

// A best-of-three that goes 2-1.
func TestBestOfThreeTwoOne(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	lose(t, g, bob.ID)
	readyUp(t, g, alice.ID, bob.ID)
	lose(t, g, alice.ID)
	if g.State != StateBetweenGames {
		t.Fatalf("at 1-1 the state is %q, want between games", g.State)
	}
	if got := score(g, alice.ID, bob.ID); got != "1-1" {
		t.Errorf("score = %s, want 1-1", got)
	}
	readyUp(t, g, alice.ID, bob.ID)
	assertFreshGame(t, g, 3)
	lose(t, g, alice.ID)
	if g.State != StateEnded || !g.Match.Over || g.Match.Winner != bob.ID {
		t.Fatalf("state %q, over %v, winner %v; want Bob's match", g.State, g.Match.Over, g.Match.Winner)
	}
	if got := score(g, alice.ID, bob.ID); got != "1-2" {
		t.Errorf("score = %s, want 1-2", got)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != bob.Seat {
		t.Errorf("WinnerSeat = %d, %v; want Bob's", seat, ok)
	}
}

// The match's winner need not have won its last game: one win and two
// draws is a match won, and the table's result is the match's.
func TestMatchResultIsNotTheLastGames(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	lose(t, g, bob.ID)
	for i := 0; i < 2; i++ {
		readyUp(t, g, alice.ID, bob.ID)
		g.mu.Lock()
		g.endGameLocked(GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseAllLost})
		g.mu.Unlock()
	}
	if g.State != StateEnded || !g.Match.Over || g.Match.Winner != alice.ID {
		t.Fatalf("state %q, over %v, winner %v; want Alice's match on one win", g.State, g.Match.Over, g.Match.Winner)
	}
	if g.Outcome == nil || g.Outcome.Kind != OutcomeDraw {
		t.Fatalf("the last game's outcome is %+v, want its draw", g.Outcome)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != alice.Seat {
		t.Errorf("WinnerSeat = %d, %v; want Alice's", seat, ok)
	}
	if got := g.OutcomeKind(); got != OutcomeWin {
		t.Errorf("OutcomeKind = %q, want win", got)
	}

	// Three games with a win each and a draw is a drawn match.
	g, alice, bob = matchTable(t, 3)
	lose(t, g, bob.ID)
	readyUp(t, g, alice.ID, bob.ID)
	lose(t, g, alice.ID)
	readyUp(t, g, alice.ID, bob.ID)
	g.mu.Lock()
	g.endGameLocked(GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseAllLost})
	g.mu.Unlock()
	if !g.Match.Over || g.Match.Winner != uuid.Nil {
		t.Fatalf("over %v, winner %v; want a drawn match", g.Match.Over, g.Match.Winner)
	}
	if _, ok := g.WinnerSeat(); ok {
		t.Error("a drawn match names a winner")
	}
	if got := g.OutcomeKind(); got != OutcomeDraw {
		t.Errorf("OutcomeKind = %q, want draw", got)
	}
}

// CR 103.1: "In a match of several games, the loser of the previous
// game chooses who takes the first turn. If the previous game was a
// draw, the player who made the choice in that game makes the choice
// in this game."
func TestLoserChoosesWhoPlaysFirst(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	if g.Match.Chooser != alice.ID {
		t.Fatalf("game 1's chooser = %v, want its starting player", g.Match.Chooser)
	}
	lose(t, g, bob.ID)
	b := g.Match.Between
	if b.Chooser != bob.ID || !b.PlayFirst {
		t.Fatalf("between = %+v, want Bob choosing, playing first unless he says otherwise", b)
	}
	if err := g.ChoosePlayFirst(alice.ID, true); !errors.Is(err, ErrNotMatchChooser) {
		t.Errorf("the winner choosing: %v, want ErrNotMatchChooser", err)
	}
	if err := g.ChoosePlayFirst(bob.ID, false); err != nil {
		t.Fatalf("ChoosePlayFirst: %v", err)
	}
	events := len(g.Events)
	readyUp(t, g, alice.ID, bob.ID)
	if g.StartingSeat != alice.Seat || g.Turn.ActiveSeat != alice.Seat {
		t.Errorf("starting seat %d, active seat %d; Bob chose to draw, so Alice plays first", g.StartingSeat, g.Turn.ActiveSeat)
	}
	var began *Event
	for i := events; i < len(g.Events); i++ {
		if g.Events[i].Kind == EventMatchGame {
			began = &g.Events[i]
		}
	}
	if began == nil || began.Actor != bob.ID || began.Target != alice.ID || began.Amount != 2 || began.Label != MatchChoseDraw {
		t.Errorf("match_game event = %+v, want Bob choosing to draw in game 2", began)
	}

	// Game 2, unchosen: the loser plays first.
	lose(t, g, alice.ID)
	if g.Match.Between.Chooser != alice.ID {
		t.Fatalf("after Alice lost, the chooser is %v", g.Match.Between.Chooser)
	}
	readyUp(t, g, alice.ID, bob.ID)
	if g.StartingSeat != alice.Seat {
		t.Errorf("starting seat %d; the loser who chose nothing plays first", g.StartingSeat)
	}

	// After a draw, whoever chose for the drawn game chooses again.
	g, alice, bob = matchTable(t, 3)
	g.mu.Lock()
	g.endGameLocked(GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseAllLost})
	g.mu.Unlock()
	if g.State != StateBetweenGames || g.Match.Between.Chooser != alice.ID {
		t.Fatalf("after a drawn game 1 the chooser is %v, want game 1's (Alice)", g.Match.Between.Chooser)
	}
	_ = bob
}

func poolIDs(cs []Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cs))
	for i, c := range cs {
		ids[i] = c.InstanceID
	}
	return ids
}

// Sideboarding: a legal swap is the deck the next game is played with,
// and a swap that makes an illegal deck is refused and changes nothing.
func TestSideboardSwap(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	registered := append([]uuid.UUID(nil), g.Match.deck(bob.ID).Registered...)
	lose(t, g, bob.ID)
	d := g.Match.deck(bob.ID)
	if len(d.Main) != 60 || len(d.Side) != 15 {
		t.Fatalf("Bob's pool is %d and %d, want 60 and 15", len(d.Main), len(d.Side))
	}

	// One card out with nothing in: fifty-nine cards (CR 100.2a).
	short := poolIDs(d.Main[1:])
	err := g.SetMatchDeck(bob.ID, short)
	if !errors.Is(err, ErrSideboardIllegal) {
		t.Fatalf("a 59-card deck: %v, want ErrSideboardIllegal", err)
	}
	if d = g.Match.deck(bob.ID); len(d.Main) != 60 || len(d.Side) != 15 {
		t.Fatalf("a refused deck changed the pool to %d and %d", len(d.Main), len(d.Side))
	}
	// A card that is not Bob's, and one named twice.
	foreign := append(poolIDs(d.Main), g.Match.deck(alice.ID).Side[0].InstanceID)
	if err := g.SetMatchDeck(bob.ID, foreign); !errors.Is(err, ErrSideboardUnknownCard) {
		t.Errorf("a deck with the opponent's card: %v, want ErrSideboardUnknownCard", err)
	}
	twice := append(poolIDs(d.Main), d.Main[0].InstanceID)
	if err := g.SetMatchDeck(bob.ID, twice); !errors.Is(err, ErrSideboardUnknownCard) {
		t.Errorf("a deck naming a card twice: %v, want ErrSideboardUnknownCard", err)
	}

	// One out, one in.
	out, in := d.Main[0], d.Side[0]
	swapped := append(poolIDs(d.Main[1:]), in.InstanceID)
	if err := g.SetMatchDeck(bob.ID, swapped); err != nil {
		t.Fatalf("a one-for-one swap: %v", err)
	}
	// Ready is final until it is taken back.
	readyUp(t, g, bob.ID)
	if err := g.SetMatchDeck(bob.ID, poolIDs(d.Main)); !errors.Is(err, ErrSideboardReady) {
		t.Errorf("changing the deck while ready: %v, want ErrSideboardReady", err)
	}
	readyUp(t, g, alice.ID)
	assertFreshGame(t, g, 2)

	bob = g.Seats[1]
	names := func(zs ...*Zone) map[string]bool {
		m := map[string]bool{}
		for _, z := range zs {
			for _, c := range z.Cards {
				m[c.Name] = true
			}
		}
		return m
	}
	deck, side := names(bob.Hand, bob.Library), names(bob.Sideboard)
	if !deck[in.Name] || deck[out.Name] {
		t.Errorf("game 2's deck: has %q = %v, has %q = %v; want the card brought in and not the one taken out", in.Name, deck[in.Name], out.Name, deck[out.Name])
	}
	if !side[out.Name] || side[in.Name] {
		t.Errorf("game 2's sideboard: has %q = %v, has %q = %v", out.Name, side[out.Name], in.Name, side[in.Name])
	}
	for _, c := range bob.Sideboard.Cards {
		if !c.IsSideboard {
			t.Fatalf("%q is in the sideboard without its stamp", c.Name)
		}
	}
	for _, c := range append(append([]Card(nil), bob.Hand.Cards...), bob.Library.Cards...) {
		if c.IsSideboard {
			t.Fatalf("%q is in the deck with the sideboard's stamp", c.Name)
		}
	}

	// The deck stays as sideboarded for game 3, and the registered list
	// is still on record for when the match ends.
	lose(t, g, g.Seats[0].ID)
	readyUp(t, g, g.Seats[0].ID, g.Seats[1].ID)
	bob = g.Seats[1]
	if deck := names(bob.Hand, bob.Library); !deck[in.Name] || deck[out.Name] {
		t.Error("game 3 went back to the registered list; a deck returns to it only when the match ends")
	}
	got := g.Match.deck(bob.ID).Registered
	if len(got) != len(registered) {
		t.Fatalf("registered list has %d cards, want %d", len(got), len(registered))
	}
	for i := range got {
		if got[i] != registered[i] {
			t.Fatal("sideboarding changed the registered list")
		}
	}
}

// The deck check a sideboarded deck goes through, on its own.
func TestCheckMatchDeck(t *testing.T) {
	spec := FormatStandard.Spec()
	cards := func(name string, n int, typeLine string) []Card {
		out := make([]Card, n)
		for i := range out {
			out[i] = NewCard(name, uuid.Nil)
			out[i].TypeLine = typeLine
		}
		return out
	}
	filler := func(n int) []Card {
		var out []Card
		for i := 0; i < n; i++ {
			out = append(out, NewCard(fmt.Sprintf("Filler %d", i), uuid.Nil))
		}
		return out
	}
	if err := CheckMatchDeck(spec, filler(60), filler(15), 0, nil); err != nil {
		t.Errorf("sixty and fifteen: %v", err)
	}
	if err := CheckMatchDeck(spec, filler(59), filler(15), 0, nil); !errors.Is(err, ErrSideboardIllegal) {
		t.Errorf("fifty-nine cards: %v", err)
	}
	if err := CheckMatchDeck(spec, filler(60), filler(16), 0, nil); !errors.Is(err, ErrSideboardIllegal) {
		t.Errorf("a sixteen-card sideboard: %v", err)
	}
	// CR 100.4a: the four-card limit counts the deck and the sideboard
	// together. Three in the deck and two in the sideboard is five.
	main := append(filler(57), cards("Lightning Strike", 3, "Instant")...)
	side := append(filler(13), cards("Lightning Strike", 2, "Instant")...)
	err := CheckMatchDeck(spec, main, side, 0, nil)
	if !errors.Is(err, ErrSideboardIllegal) {
		t.Errorf("five copies across deck and sideboard: %v", err)
	}
	// Basic lands are not counted (CR 100.2a), nor a card whose own
	// text sets its limit.
	main = append(filler(40), cards("Mountain", 20, "Basic Land — Mountain")...)
	if err := CheckMatchDeck(spec, main, filler(15), 0, nil); err != nil {
		t.Errorf("twenty Mountains: %v", err)
	}
	main = append(filler(50), cards("Hare Apparent", 10, "Creature — Rabbit Noble")...)
	if err := CheckMatchDeck(spec, main, filler(15), 0, map[string]bool{"Hare Apparent": true}); err != nil {
		t.Errorf("ten of a card whose text allows any number: %v", err)
	}
}

// The timer ends sideboarding: at the deadline the next game begins
// with nobody ready, each deck as its player left it.
func TestTimerEndsSideboarding(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	before := time.Now().UTC()
	lose(t, g, bob.ID)
	deadline, open := g.SideboardingDeadline()
	if !open {
		t.Fatal("no sideboarding window after game 1")
	}
	if d := deadline.Sub(before); d < SideboardingTime-time.Second || d > SideboardingTime+time.Minute {
		t.Errorf("the window is %v long, want %v", d, SideboardingTime)
	}
	d := g.Match.deck(alice.ID)
	in := d.Side[3]
	if err := g.SetMatchDeck(alice.ID, append(poolIDs(d.Main[1:]), in.InstanceID)); err != nil {
		t.Fatal(err)
	}
	if g.AdvanceMatch(deadline.Add(-time.Second)) {
		t.Fatal("the next game began before the deadline with nobody ready")
	}
	if g.State != StateBetweenGames {
		t.Fatal("the window closed early")
	}
	if !g.AdvanceMatch(deadline) {
		t.Fatal("the deadline did not begin the next game")
	}
	assertFreshGame(t, g, 2)
	found := false
	for _, z := range []*Zone{g.Seats[0].Hand, g.Seats[0].Library} {
		for _, c := range z.Cards {
			found = found || c.Name == in.Name
		}
	}
	if !found {
		t.Error("the swap made before the timer ran out is not in game 2's deck")
	}
	if g.AdvanceMatch(deadline.Add(time.Hour)) {
		t.Error("AdvanceMatch began a game with one in play")
	}
	if _, open := g.SideboardingDeadline(); open {
		t.Error("a window is reported open during a game")
	}
}

func matchRoundTrip(t *testing.T, g *Game) *Game {
	t.Helper()
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	back, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	return back
}

// A restore in the middle of a match keeps the score, the pools and
// who chose, and the match goes on to its end.
func TestSnapshotRestoreMidMatch(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	lose(t, g, bob.ID)
	d := g.Match.deck(bob.ID)
	in := d.Side[0]
	if err := g.SetMatchDeck(bob.ID, append(poolIDs(d.Main[1:]), in.InstanceID)); err != nil {
		t.Fatal(err)
	}
	readyUp(t, g, alice.ID, bob.ID)

	back := matchRoundTrip(t, g)
	if back.State != StateActive || back.Match == nil {
		t.Fatalf("restored state %q, match %v", back.State, back.Match)
	}
	m := back.Match
	if m.BestOf != 3 || len(m.Games) != 1 || m.Games[0].Winner != alice.ID || m.GameNumber() != 2 || m.Chooser != bob.ID {
		t.Errorf("restored match = %+v", m)
	}
	if got := score(back, alice.ID, bob.ID); got != "1-0" {
		t.Errorf("restored score = %s, want 1-0", got)
	}
	rd := m.deck(bob.ID)
	if rd == nil || len(rd.Main) != 60 || len(rd.Side) != 15 || len(rd.Registered) != 60 {
		t.Fatalf("restored pool = %+v", rd)
	}
	inMain := false
	for _, c := range rd.Main {
		inMain = inMain || c.InstanceID == in.InstanceID
	}
	if !inMain {
		t.Error("the restored pool lost the sideboard swap")
	}

	// The match plays on from the restore.
	lose(t, back, alice.ID)
	if back.State != StateBetweenGames {
		t.Fatalf("after game 2 on the restored table the state is %q", back.State)
	}
	readyUp(t, back, alice.ID, bob.ID)
	assertFreshGame(t, back, 3)
	lose(t, back, alice.ID)
	if back.State != StateEnded || back.Match.Winner != bob.ID {
		t.Errorf("state %q, winner %v; want Bob's match", back.State, back.Match.Winner)
	}
}

// A restore during sideboarding restores into the window: the score,
// what each player has done so far, who is ready and who chooses.
func TestSnapshotRestoreDuringSideboarding(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	lose(t, g, bob.ID)
	d := g.Match.deck(alice.ID)
	in := d.Side[5]
	if err := g.SetMatchDeck(alice.ID, append(poolIDs(d.Main[1:]), in.InstanceID)); err != nil {
		t.Fatal(err)
	}
	if err := g.ChoosePlayFirst(bob.ID, false); err != nil {
		t.Fatal(err)
	}
	readyUp(t, g, alice.ID)
	deadline := g.Match.Between.Deadline

	back := matchRoundTrip(t, g)
	if back.State != StateBetweenGames || !back.SideboardingOpen() {
		t.Fatalf("restored state %q, want between games", back.State)
	}
	b := back.Match.Between
	if b == nil || b.Chooser != bob.ID || b.PlayFirst || len(b.Ready) != 1 || b.Ready[0] != alice.ID || !b.Deadline.Equal(deadline) {
		t.Fatalf("restored window = %+v", b)
	}
	if back.Outcome == nil || back.Outcome.Winner != alice.ID {
		t.Errorf("the finished game's outcome was lost: %+v", back.Outcome)
	}
	// A table restored after its deadline waits for its players.
	later := deadline.Add(time.Hour)
	back.ExtendSideboarding(later)
	if got := back.Match.Between.Deadline; got.Before(later.Add(sideboardingRestoreGrace)) {
		t.Errorf("deadline after a late restore = %v, want at least %v from now", got, sideboardingRestoreGrace)
	}
	if back.AdvanceMatch(later) {
		t.Fatal("a restored table began the next game the moment it came back")
	}
	// Bob may still sideboard, and readying up begins game 2.
	bd := back.Match.deck(bob.ID)
	if err := back.SetMatchDeck(bob.ID, append(poolIDs(bd.Main[1:]), bd.Side[0].InstanceID)); err != nil {
		t.Fatalf("sideboarding after the restore: %v", err)
	}
	readyUp(t, back, bob.ID)
	assertFreshGame(t, back, 2)
	if back.StartingSeat != alice.Seat {
		t.Errorf("starting seat %d; Bob chose to draw before the restore", back.StartingSeat)
	}
	found := false
	for _, z := range []*Zone{back.Seats[0].Hand, back.Seats[0].Library} {
		for _, c := range z.Cards {
			found = found || c.Name == in.Name
		}
	}
	if !found {
		t.Error("Alice's swap from before the restore is not in game 2's deck")
	}
}

// A Commander table has no match anything: no Match, nothing in its
// snapshot or its settings, no sideboarding step when a game ends, and
// every match verb refused.
func TestCommanderTableHasNoMatch(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C")); err != nil {
			t.Fatal(err)
		}
	}
	three := 3
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &three}); !errors.Is(err, ErrBestOfFormat) {
		t.Fatalf("best-of-three at a Commander table: %v, want ErrBestOfFormat", err)
	}
	one := 1
	events := len(g.Events)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &one}); err != nil {
		t.Fatalf("naming best-of-one at a Commander table: %v", err)
	}
	if len(g.Events) != events || g.Settings != DefaultTableSettings() {
		t.Errorf("naming best-of-one changed a Commander table: %+v", g.Settings)
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	if g.Match != nil {
		t.Fatalf("a Commander table has a match: %+v", g.Match)
	}
	if g.MatchGameNumber() != 0 || g.SideboardingOpen() || g.MatchSnapshot() != nil {
		t.Error("a Commander table reports a match")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"match"`, `"bestOf"`, `"between_games"`} {
		if bytes.Contains(raw, []byte(key)) {
			t.Errorf("a Commander table's snapshot contains %s", key)
		}
	}
	a, b := g.Seats[0], g.Seats[1]
	if err := g.ConcedeMatch(a.ID); !errors.Is(err, ErrNoMatch) {
		t.Errorf("ConcedeMatch at a Commander table: %v, want ErrNoMatch", err)
	}
	for _, err := range []error{
		g.SetSideboardReady(a.ID, true),
		g.SetMatchDeck(a.ID, nil),
		g.ChoosePlayFirst(a.ID, true),
	} {
		if !errors.Is(err, ErrNotSideboarding) {
			t.Errorf("a sideboarding verb at a Commander table: %v, want ErrNotSideboarding", err)
		}
	}
	if g.AdvanceMatch(time.Now().Add(24 * time.Hour)) {
		t.Error("AdvanceMatch did something at a Commander table")
	}
	lose(t, g, b.ID)
	if g.State != StateEnded || g.Match != nil {
		t.Fatalf("a Commander game ended in state %q with match %v", g.State, g.Match)
	}
	if g.Outcome == nil || g.Outcome.Winner != a.ID || g.Outcome.Cause != OutcomeCauseLastStanding {
		t.Errorf("outcome = %+v", g.Outcome)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != a.Seat {
		t.Errorf("WinnerSeat = %d, %v", seat, ok)
	}
	// And a snapshot from before matches restores with none.
	if back := matchRoundTrip(t, g); back.Match != nil || back.State != StateEnded {
		t.Errorf("restored Commander table: match %v, state %q", back.Match, back.State)
	}
}

// A best-of-one Standard table is a match of one game, and it ends the
// way it did before there were matches: the game's end is the table's,
// with the game's own outcome, and no sideboarding step.
func TestBestOfOneStandardEndsWithItsGame(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	raw, _ := json.Marshal(g.Settings)
	if bytes.Contains(raw, []byte("bestOf")) {
		t.Errorf("a best-of-one table's settings serialise a match length: %s", raw)
	}
	if g.Settings.EffectiveBestOf() != 1 {
		t.Fatalf("a new Standard table is best-of-%d", g.Settings.EffectiveBestOf())
	}
	for _, name := range []string{"Alice", "Bob"} {
		if _, err := g.AddPlayer(name, buildMatchDeck(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	if g.Match == nil || g.Match.BestOf != 1 || g.Match.Decks != nil {
		t.Fatalf("match = %+v, want a best-of-one that keeps no pools", g.Match)
	}
	alice, bob := g.Seats[0], g.Seats[1]
	lose(t, g, bob.ID)
	if g.State != StateEnded {
		t.Fatalf("state %q, want ended", g.State)
	}
	if g.Outcome == nil || g.Outcome.Kind != OutcomeWin || g.Outcome.Winner != alice.ID || g.Outcome.Cause != OutcomeCauseLastStanding {
		t.Errorf("outcome = %+v, want Alice the last one standing", g.Outcome)
	}
	if !g.Match.Over || g.Match.Winner != alice.ID || g.Match.Between != nil {
		t.Errorf("match = %+v", g.Match)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != alice.Seat {
		t.Errorf("WinnerSeat = %d, %v", seat, ok)
	}
	if got := g.OutcomeKind(); got != OutcomeWin {
		t.Errorf("OutcomeKind = %q", got)
	}
	if g.SideboardingOpen() {
		t.Error("a best-of-one opened a sideboarding window")
	}
}

// Conceding a game concedes that game; conceding the match is its own
// verb, and ends the match whatever the score.
func TestConcedeTheGameOrTheMatch(t *testing.T) {
	// During a game, at 1-0 up: the conceder still loses the match.
	g, alice, bob := matchTable(t, 3)
	lose(t, g, bob.ID)
	readyUp(t, g, alice.ID, bob.ID)
	if err := g.ConcedeMatch(alice.ID); err != nil {
		t.Fatalf("ConcedeMatch: %v", err)
	}
	if g.State != StateEnded || !g.Match.Over || g.Match.Winner != bob.ID || g.Match.ConcededBy != alice.ID {
		t.Fatalf("state %q, match %+v; want Bob's match by concession", g.State, g.Match)
	}
	if len(g.Match.Games) != 2 || g.Match.Games[1].Winner != bob.ID {
		t.Errorf("games = %+v; the game in play is lost with the match", g.Match.Games)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != bob.Seat {
		t.Errorf("WinnerSeat = %d, %v; want Bob's", seat, ok)
	}

	// Between games.
	g, alice, bob = matchTable(t, 3)
	lose(t, g, bob.ID)
	if err := g.ConcedeMatch(bob.ID); err != nil {
		t.Fatalf("ConcedeMatch between games: %v", err)
	}
	if g.State != StateEnded || g.Match.Winner != alice.ID || g.Match.Between != nil {
		t.Fatalf("state %q, match %+v", g.State, g.Match)
	}
	if g.Outcome == nil || g.Outcome.Cause != OutcomeCauseMatchConceded || g.Outcome.Winner != alice.ID {
		t.Errorf("outcome = %+v", g.Outcome)
	}
	if len(g.Match.Games) != 1 {
		t.Errorf("%d games on record, want the one that was played", len(g.Match.Games))
	}
	if err := g.ConcedeMatch(alice.ID); !errors.Is(err, ErrNoMatch) {
		t.Errorf("conceding a finished match: %v, want ErrNoMatch", err)
	}

	// A seat an admin removes mid-game is not coming back.
	g, alice, bob = matchTable(t, 3)
	if err := g.RemoveSeatFromGame(bob.ID); err != nil {
		t.Fatal(err)
	}
	if g.State != StateEnded || g.Match.Winner != alice.ID {
		t.Errorf("after a seat was removed: state %q, match %+v", g.State, g.Match)
	}
}

// A bot seat plays every game of the match and does not sideboard: it
// is always ready, so the next game waits only on the person.
func TestBotSeatIsAlwaysReady(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	bob.IsBot = true
	lose(t, g, bob.ID)
	if deadline, open := g.SideboardingDeadline(); !open || deadline.IsZero() {
		t.Fatal("the window should be open and waiting on Alice")
	}
	d := g.Match.deck(bob.ID)
	if err := g.SetMatchDeck(bob.ID, poolIDs(d.Main)); !errors.Is(err, ErrSideboardReady) {
		t.Errorf("a bot sideboarding: %v, want ErrSideboardReady", err)
	}
	readyUp(t, g, alice.ID)
	assertFreshGame(t, g, 2)
	if !g.Seats[1].IsBot {
		t.Fatal("the bot seat is not a bot in game 2")
	}
	// The bot lost, so it chose, and a bot plays first.
	if g.StartingSeat != 1 {
		t.Errorf("starting seat %d, want the bot's", g.StartingSeat)
	}

	// Two bots: nobody to wait for, so the window is over at once and
	// whatever drives the table begins the next game.
	g, alice, bob = matchTable(t, 3)
	alice.IsBot, bob.IsBot = true, true
	lose(t, g, bob.ID)
	if deadline, open := g.SideboardingDeadline(); !open || !deadline.IsZero() {
		t.Fatalf("deadline %v, open %v; a window everyone is ready for ends now", deadline, open)
	}
	if !g.AdvanceMatch(time.Now()) {
		t.Fatal("AdvanceMatch did not begin the next game for two ready bots")
	}
	assertFreshGame(t, g, 2)
}

// Undoing the action that ended a game takes the score back with it.
func TestUndoTakesTheScoreBack(t *testing.T) {
	g, alice, bob := matchTable(t, 3)
	before := g.Clone()
	lose(t, g, bob.ID)
	if g.State != StateBetweenGames || len(g.Match.Games) != 1 {
		t.Fatal("game 1 did not end")
	}
	g.mu.Lock()
	g.RestoreFrom(before)
	g.mu.Unlock()
	if g.State != StateActive || len(g.Match.Games) != 0 || g.Match.Between != nil {
		t.Errorf("after the undo: state %q, match %+v", g.State, g.Match)
	}
	_ = alice
}

// The match length is a lobby setting, and only for a format that
// plays matches.
func TestBestOfSetting(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	three, two, one := 3, 2, 1
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &two}); !errors.Is(err, ErrInvalidSetting) {
		t.Errorf("best-of-two: %v, want ErrInvalidSetting", err)
	}
	events := len(g.Events)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &three}); err != nil {
		t.Fatal(err)
	}
	if g.Settings.EffectiveBestOf() != 3 {
		t.Fatal("the table is not best-of-three")
	}
	last := g.Events[len(g.Events)-1]
	if len(g.Events) != events+1 || last.Kind != EventSettingsChanged || last.Label != SettingBestOf || last.SettingOld != "1" || last.SettingNew != "3" {
		t.Errorf("events after the change: %+v", g.Events[events:])
	}
	// Back to Commander: no match length left behind.
	setFormat(t, g, FormatCommander)
	if g.Settings.BestOf != 0 || g.Settings.EffectiveBestOf() != 1 {
		t.Errorf("a Commander table kept best-of-%d", g.Settings.BestOf)
	}
	// A format and its match length in one patch (a saved setup).
	std := FormatStandard
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &std, BestOf: &three}); err != nil {
		t.Fatalf("format and best-of in one patch: %v", err)
	}
	if g.Settings.EffectiveBestOf() != 3 {
		t.Error("the patch's match length was not applied")
	}
	for _, name := range []string{"Alice", "Bob"} {
		_, _ = g.AddPlayer(name, buildMatchDeck(name))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &one}); !errors.Is(err, ErrBestOfLocked) {
		t.Errorf("changing the match length after Start: %v, want ErrBestOfLocked", err)
	}
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{BestOf: &three}); err != nil {
		t.Errorf("naming the length it already has: %v", err)
	}
}

// A companion is revealed only by the deck it was registered with.
func TestCompanionOnlyWithTheRegisteredDeck(t *testing.T) {
	d := MatchDeck{}
	for i := 0; i < 3; i++ {
		c := NewCard(fmt.Sprintf("Main %d", i), uuid.Nil)
		d.Main = append(d.Main, c)
		d.Registered = append(d.Registered, c.InstanceID)
	}
	sortIDs(d.Registered)
	comp := NewCard("Companion", uuid.Nil)
	comp.IsCompanion = true
	side := NewCard("Side", uuid.Nil)
	side.IsSideboard = true
	d.Side = []Card{comp, side}

	find := func(cs []Card, name string) Card {
		for _, c := range cs {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("no %q among the seated cards", name)
		return Card{}
	}
	if c := find(d.gameCards(), "Companion"); !c.IsCompanion || c.IsSideboard {
		t.Errorf("with the registered deck the companion is seated as %+v", c)
	}
	// One for one.
	d.Main[0], d.Side[1] = d.Side[1], d.Main[0]
	seated := d.gameCards()
	if c := find(seated, "Companion"); c.IsCompanion || !c.IsSideboard {
		t.Error("a changed deck still reveals its companion")
	}
	if c := find(seated, "Side"); c.IsSideboard {
		t.Error("a card brought in is still stamped for the sideboard")
	}
	if c := find(seated, "Main 0"); !c.IsSideboard {
		t.Error("a card taken out is not stamped for the sideboard")
	}
}
