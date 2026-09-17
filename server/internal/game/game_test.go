package game

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// buildTestDeck returns a synthetic 99-card deck plus a commander,
// owned by a placeholder UUID (the real owner is stamped by AddPlayer).
// Card identities and names are placeholders — S02 does not care
// about real card data, which arrives with the Scryfall pipeline in
// S04.
func buildTestDeck(commanderName string) []Card {
	owner := uuid.Nil
	deck := make([]Card, 0, 100)
	deck = append(deck, NewCommander(commanderName, owner))
	for i := 0; i < 99; i++ {
		deck = append(deck, NewCard(fmt.Sprintf("Filler Card %d", i+1), owner))
	}
	return deck
}

func TestNewGameStartsInLobby(t *testing.T) {
	g := NewGame()
	if g.State != StateLobby {
		t.Errorf("state: got %q, want %q", g.State, StateLobby)
	}
	if g.ID == uuid.Nil {
		t.Error("game ID is nil")
	}
	if g.Battlefield == nil || g.Stack == nil || g.Exile == nil {
		t.Error("shared zones should be pre-constructed")
	}
	if g.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestAddPlayerRoutesCommanderToCommandZone(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("Alice", buildTestDeck("Atraxa, Praetors' Voice"))
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if p.Command.Size() != 1 {
		t.Errorf("command zone: got %d cards, want 1", p.Command.Size())
	}
	if p.Library.Size() != 99 {
		t.Errorf("library: got %d cards, want 99", p.Library.Size())
	}
	cmdr, _ := p.Command.Top()
	if !cmdr.IsCommander {
		t.Error("card in command zone should be flagged as commander")
	}
	if cmdr.Owner != p.ID {
		t.Error("commander owner should be stamped to player ID by AddPlayer")
	}
}

func TestAddPlayerRejectsEmptyName(t *testing.T) {
	g := NewGame()
	if _, err := g.AddPlayer("", buildTestDeck("X")); err != ErrEmptyName {
		t.Errorf("empty name: got %v, want ErrEmptyName", err)
	}
}

func TestAddPlayerRejectsEmptyDeck(t *testing.T) {
	g := NewGame()
	if _, err := g.AddPlayer("A", nil); err != ErrEmptyDeck {
		t.Errorf("empty deck: got %v, want ErrEmptyDeck", err)
	}
}

func TestAddPlayerFullGame(t *testing.T) {
	g := NewGame()
	for i := 0; i < MaxPlayers; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C")); err != nil {
			t.Fatalf("AddPlayer #%d: %v", i+1, err)
		}
	}
	if _, err := g.AddPlayer("Overflow", buildTestDeck("C")); err != ErrGameFull {
		t.Errorf("5th player: got %v, want ErrGameFull", err)
	}
}

func TestStartRequiresMinPlayers(t *testing.T) {
	g := NewGame()
	if err := g.Start(nil); err != ErrNotEnoughPlayers {
		t.Errorf("start with 0 players: got %v, want ErrNotEnoughPlayers", err)
	}
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	if err := g.Start(nil); err != ErrNotEnoughPlayers {
		t.Errorf("start with 1 player: got %v, want ErrNotEnoughPlayers", err)
	}
}

func TestStartTransitionsAndShuffles(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}

	// Capture pre-shuffle order to verify Start actually shuffles.
	before := make([]uuid.UUID, len(g.Seats[0].Library.Cards))
	for i, c := range g.Seats[0].Library.Cards {
		before[i] = c.InstanceID
	}

	// Use a deterministic RNG so this test is reproducible.
	r := rand.New(rand.NewPCG(1, 2))
	if err := g.Start(r); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.State != StateActive {
		t.Errorf("state: got %q, want %q", g.State, StateActive)
	}
	if g.Turn.Number != 1 || g.Turn.ActiveSeat != 0 || g.Turn.Step != StepUntap {
		t.Errorf("starting turn: got %+v", g.Turn)
	}

	// Verify the library was actually shuffled — at least one index
	// should differ after shuffling a 99-card list (probability of a
	// no-op shuffle is astronomically small).
	differs := false
	for i, c := range g.Seats[0].Library.Cards {
		if c.InstanceID != before[i] {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("library order is unchanged after Start; shuffle did not run")
	}
}

func TestDoubleStartRejected(t *testing.T) {
	g := NewGame()
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	_, _ = g.AddPlayer("B", buildTestDeck("C"))
	_ = g.Start(nil)
	if err := g.Start(nil); err != ErrGameAlreadyStarted {
		t.Errorf("double start: got %v, want ErrGameAlreadyStarted", err)
	}
}

func TestAdvanceStepRequiresActiveGame(t *testing.T) {
	g := NewGame()
	if _, err := g.AdvanceStep(); err != ErrGameNotActive {
		t.Errorf("advance in lobby: got %v, want ErrGameNotActive", err)
	}
}

// TestFullFourPlayerTurnCycle is the S02 exit criteria test: it
// stands up a 4-player Commander game, walks through a complete turn
// cycle, and asserts the expected state at each step. After S13,
// Untap and Cleanup do not grant priority and auto-advance from
// their step entry hooks, so AdvanceStep callers see the cursor walk
// the externally-visible sequence Upkeep → Draw → ... → End →
// (next seat's) Upkeep — a single AdvanceStep from End walks the
// cursor all the way through Cleanup, the next seat's Untap, and
// lands on Upkeep.
func TestFullFourPlayerTurnCycle(t *testing.T) {
	g := NewGame()
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	for i, name := range names {
		p, err := g.AddPlayer(name, buildTestDeck(fmt.Sprintf("Commander %d", i+1)))
		if err != nil {
			t.Fatalf("AddPlayer %s: %v", name, err)
		}
		if p.Seat != i {
			t.Errorf("seat for %s: got %d, want %d", name, p.Seat, i)
		}
		if p.Life != StartingLife {
			t.Errorf("starting life for %s: got %d, want %d", name, p.Life, StartingLife)
		}
		if p.Command.Size() != 1 {
			t.Errorf("%s command zone: got %d, want 1", name, p.Command.Size())
		}
		if p.Library.Size() != 99 {
			t.Errorf("%s library: got %d, want 99", name, p.Library.Size())
		}
	}

	if len(g.Seats) != 4 {
		t.Fatalf("seats: got %d, want 4", len(g.Seats))
	}

	if err := g.Start(rand.New(rand.NewPCG(7, 13))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.State != StateActive {
		t.Fatalf("state after start: %q", g.State)
	}

	// Close the mulligan window so the first auto-untap fires and
	// the cursor lands on Upkeep with priority on seat 0.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand %s: %v", p.Name, err)
		}
	}

	// S13.4: this test walks ~4 turns and the auto-draw step would
	// push every hand past the default 7-card cap, freezing the
	// cursor at cleanup waiting for an interactive discard. The
	// turn-cycle test isn't about discard — uncapping every seat
	// keeps the focus on the priority + step sequence.
	for _, p := range g.Seats {
		p.MaxHandSize = NoMaxHandSize
	}

	active := g.ActivePlayer()
	if active == nil || active.Name != "Alice" {
		t.Fatalf("active player: got %v, want Alice", active)
	}

	// External step sequence as observed by AdvanceStep callers
	// post-S13. Untap and Cleanup are auto-advanced through their
	// entry hooks and are never visible at the AdvanceStep boundary.
	visible := []Step{
		StepUpkeep,
		StepDraw,
		StepPrecombatMain,
		StepBeginCombat,
		StepDeclareAttackers,
		StepDeclareBlockers,
		StepCombatDamage,
		StepEndCombat,
		StepPostcombatMain,
		StepEnd,
	}

	if g.Turn.Step != visible[0] {
		t.Errorf("turn starts at step %q after mulligan close, want %q", g.Turn.Step, visible[0])
	}
	if g.Turn.Phase != PhaseOf(visible[0]) {
		t.Errorf("turn starts at phase %q, want %q", g.Turn.Phase, PhaseOf(visible[0]))
	}
	if g.Turn.PriorityHolder != 0 {
		t.Errorf("PriorityHolder after mulligan close: got %d, want 0", g.Turn.PriorityHolder)
	}

	// Walk the remaining visible steps of Alice's turn.
	for i := 1; i < len(visible); i++ {
		turn, err := g.AdvanceStep()
		if err != nil {
			t.Fatalf("AdvanceStep %d: %v", i, err)
		}
		if turn.Step != visible[i] {
			t.Errorf("step %d: got %q, want %q", i, turn.Step, visible[i])
		}
		if turn.Phase != PhaseOf(visible[i]) {
			t.Errorf("step %d phase: got %q, want %q", i, turn.Phase, PhaseOf(visible[i]))
		}
		if turn.ActiveSeat != 0 {
			t.Errorf("seat drifted off 0 mid-turn at step %d: seat=%d", i, turn.ActiveSeat)
		}
		if turn.Number != 1 {
			t.Errorf("turn number drifted off 1 mid-turn at step %d: number=%d", i, turn.Number)
		}
	}

	// One more AdvanceStep from End walks End → (auto) Cleanup →
	// (wrap, auto) Untap → Upkeep on seat 1.
	turn, err := g.AdvanceStep()
	if err != nil {
		t.Fatalf("wrap AdvanceStep: %v", err)
	}
	if turn.ActiveSeat != 1 {
		t.Errorf("seat after wrap: got %d, want 1", turn.ActiveSeat)
	}
	if turn.Number != 1 {
		t.Errorf("turn number after 1→2 seat wrap: got %d, want 1", turn.Number)
	}
	if turn.Step != StepUpkeep {
		t.Errorf("step after wrap: got %q, want %q (S13 auto-advance through Untap)",
			turn.Step, StepUpkeep)
	}

	active = g.ActivePlayer()
	if active == nil || active.Name != "Bob" {
		t.Fatalf("active player after wrap: got %v, want Bob", active)
	}

	// Burn through seats 1, 2, 3 to verify the full 4-player round
	// cycles back to seat 0 with turn number 2. Each seat starts on
	// Upkeep (post-wrap), needs len(visible)-1 AdvanceSteps to reach
	// End, and one more to wrap to the next seat's Upkeep.
	for seatsToFinish := 3; seatsToFinish > 0; seatsToFinish-- {
		for i := 1; i < len(visible); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep during seat burn: %v", err)
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep wrap: %v", err)
		}
	}

	if g.Turn.Number != 2 {
		t.Errorf("turn number after full round: got %d, want 2", g.Turn.Number)
	}
	if g.Turn.ActiveSeat != 0 {
		t.Errorf("seat after full round: got %d, want 0", g.Turn.ActiveSeat)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step after full round: got %q, want %q", g.Turn.Step, StepUpkeep)
	}
}

func TestPlayerByID(t *testing.T) {
	g := NewGame()
	p, _ := g.AddPlayer("A", buildTestDeck("C"))
	if got := g.PlayerByID(p.ID); got != p {
		t.Errorf("PlayerByID: got %v, want %v", got, p)
	}
	if got := g.PlayerByID(uuid.New()); got != nil {
		t.Errorf("PlayerByID(unknown): got %v, want nil", got)
	}
}

// TestS13AutoUntapOnMulliganClose verifies that closing the mulligan
// window fires the first step-entry hook: the active seat's tapped
// battlefield cards untap and the cursor advances to Upkeep.
func TestS13AutoUntapOnMulliganClose(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Park a tapped card on seat 0's battlefield so we can observe
	// the auto-untap firing on mulligan-close.
	c := NewCard("Tapped Filler", g.Seats[0].ID)
	c.Controller = g.Seats[0].ID
	c.Tapped = true
	g.Battlefield.PushTop(c)

	if g.Turn.Step != StepUntap || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("pre-keep: step=%q priority=%d, want Untap/NoPriority", g.Turn.Step, g.Turn.PriorityHolder)
	}

	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand %s: %v", p.Name, err)
		}
	}

	if g.Turn.Step != StepUpkeep {
		t.Errorf("step after last KeepHand: got %q, want Upkeep (auto-untap + advance)", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 0 {
		t.Errorf("PriorityHolder after auto-untap advance: got %d, want 0", g.Turn.PriorityHolder)
	}
	if g.Battlefield.Cards[0].Tapped {
		t.Error("parked card should have been auto-untapped on StepUntap entry")
	}
}

// TestS13AutoDrawOnStepEntry verifies the auto-draw turn-based action
// fires on StepDraw and respects the turn-1 skip-draw rule (CR 103.8a).
func TestS13AutoDrawOnStepEntry(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		_ = g.KeepHand(p.ID)
	}

	// Cursor at Upkeep (seat 0, turn 1). Advancing fires StepDraw
	// entry hook — which should NOT draw because seat 0 is the
	// starting player on turn 1 (CR 103.8a).
	handBefore := g.Seats[0].Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into Draw: %v", err)
	}
	if g.Turn.Step != StepDraw {
		t.Fatalf("expected cursor at StepDraw, got %q", g.Turn.Step)
	}
	if g.Seats[0].Hand.Size() != handBefore {
		t.Errorf("turn-1 starting seat drew: hand %d → %d (CR 103.8a violation)",
			handBefore, g.Seats[0].Hand.Size())
	}

	// Walk Bob's turn: AdvanceStep until we're on seat 1 at Draw.
	// Seat 1 should draw normally (not the starting seat).
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// Cursor is now on seat 1 at Upkeep (post-wrap auto-advance).
	seat1HandBefore := g.Seats[1].Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into seat 1 Draw: %v", err)
	}
	if g.Turn.Step != StepDraw || g.Turn.ActiveSeat != 1 {
		t.Fatalf("expected seat 1 StepDraw, got seat=%d step=%q", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Seats[1].Hand.Size() != seat1HandBefore+1 {
		t.Errorf("seat 1 auto-draw: hand %d → %d, want +1",
			seat1HandBefore, g.Seats[1].Hand.Size())
	}

	// Seat 0's second turn (turn 2) — should draw normally because
	// the skip is turn-1-only.
	for !(g.Turn.ActiveSeat == 0 && g.Turn.Number == 2) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	seat0Turn2HandBefore := g.Seats[0].Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into seat 0 turn 2 Draw: %v", err)
	}
	if g.Turn.Step != StepDraw || g.Turn.Number != 2 {
		t.Fatalf("expected seat 0 turn-2 Draw, got turn=%d step=%q", g.Turn.Number, g.Turn.Step)
	}
	if g.Seats[0].Hand.Size() != seat0Turn2HandBefore+1 {
		t.Errorf("seat 0 turn-2 auto-draw: hand %d → %d, want +1 (skip is turn-1-only)",
			seat0Turn2HandBefore, g.Seats[0].Hand.Size())
	}
}

// TestTurn1DrawSkipIsTwoPlayerOnly is the #692 guard. CR 103.8a skips
// the starting player's first draw step in a two-player game and
// CR 103.8b in Two-Headed Giant; CR 103.8c says that in every other
// multiplayer game no player skips it. The engine used to skip at any
// table size, which left the starting seat of a Commander pod one card
// behind for the rest of the game.
func TestTurn1DrawSkipIsTwoPlayerOnly(t *testing.T) {
	for _, tc := range []struct {
		name     string
		seats    int
		wantDraw bool
	}{
		{"two seats skip (CR 103.8a)", 2, false},
		{"three seats draw (CR 103.8c)", 3, true},
		{"four seats draw (CR 103.8c)", 4, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGameWithSeats(t, tc.seats)
			if g.Turn.Number != 1 || g.Turn.ActiveSeat != g.StartingSeat {
				t.Fatalf("harness did not park on the starting seat's first turn: turn=%d seat=%d starting=%d",
					g.Turn.Number, g.Turn.ActiveSeat, g.StartingSeat)
			}
			before := g.Seats[g.StartingSeat].Hand.Size()
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep into Draw: %v", err)
			}
			if g.Turn.Step != StepDraw {
				t.Fatalf("expected cursor at StepDraw, got %q", g.Turn.Step)
			}
			got := g.Seats[g.StartingSeat].Hand.Size()
			want := before
			if tc.wantDraw {
				want = before + 1
			}
			if got != want {
				t.Errorf("%d-seat game: starting seat hand %d -> %d, want %d",
					tc.seats, before, got, want)
			}
		})
	}
}

// TestTurn1DrawSkipIgnoresEliminations pins the "player count at the
// start of the game" reading of CR 103.8. A three-player game whose
// third seat is already out is still a three-player game for 103.8c —
// it does not collapse into a 103.8a two-player game — so the starting
// seat still draws on turn 1.
func TestTurn1DrawSkipIgnoresEliminations(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	g.Seats[2].Eliminated = true

	before := g.Seats[0].Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into Draw: %v", err)
	}
	if g.Turn.Step != StepDraw {
		t.Fatalf("expected cursor at StepDraw, got %q", g.Turn.Step)
	}
	if got := g.Seats[0].Hand.Size(); got != before+1 {
		t.Errorf("starting seat hand %d -> %d, want +1: an elimination must not turn a "+
			"three-player game into a CR 103.8a two-player one", before, got)
	}
}

// TestS13StartingSeatRecorded verifies Start() captures the starting
// seat (0 in the standard case) and exposes it via Snapshot / Clone
// for the protocol view to surface.
func TestS13StartingSeatRecorded(t *testing.T) {
	g := NewGame()
	for i := 0; i < 4; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.StartingSeat != 0 {
		t.Errorf("StartingSeat after Start: got %d, want 0", g.StartingSeat)
	}
	// Clone must preserve StartingSeat (replay/undo round-trip).
	clone := g.Clone()
	if clone.StartingSeat != 0 {
		t.Errorf("cloned StartingSeat: got %d, want 0", clone.StartingSeat)
	}
}

func TestEndTransitionsToEnded(t *testing.T) {
	g := NewGame()
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	_, _ = g.AddPlayer("B", buildTestDeck("C"))
	_ = g.Start(nil)
	g.End()
	if g.State != StateEnded {
		t.Errorf("state after End: got %q, want %q", g.State, StateEnded)
	}
}

// TestTurnAdvanceClearsPerTurnCaches is the S25 (#77) regression for
// a bypassed hook. `onTurnBeganLocked` clears the "once per turn"
// bookkeeping — loyalty activations (CR 606.3), the spell tally, the
// land-drop count — but it was only reached from AdvanceStep, which
// compares the step BEFORE the advance to the step after. In normal
// play the cursor never rests on Cleanup: End → Cleanup → the next
// seat's Untap all happens inside one runStepEntryHooksLocked
// recursion, and that recursion advanced the cursor directly. So
// AdvanceStep saw End → Cleanup, same seat, no new turn, and the
// caches survived the turn boundary forever.
//
// Routing every advance through `advanceCursorLocked` is the fix. S31's
// bot fuzzer (#287) found the same bug independently, hours apart, and
// its version of that seam also skips eliminated seats per CR 800.4a;
// this test guards that implementation. It walks a real turn boundary
// and checks all three caches.
func TestTurnAdvanceClearsPerTurnCaches(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C")); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}

	me := g.Seats[g.Turn.ActiveSeat]
	walker := uuid.New()
	g.WithWriteLock(func() {
		g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{walker: true}
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{me.ID: {Total: 2, Noncreature: 1}}
		g.LandsPlayedThisTurn = map[uuid.UUID]int{me.ID: 1}
	})

	start := g.Turn.ActiveSeat
	for i := 0; i < 30 && g.Turn.ActiveSeat == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == start {
		t.Fatalf("cursor never left seat %d", start)
	}

	if g.LoyaltyActivatedThisTurn[walker] {
		t.Error("LoyaltyActivatedThisTurn survived the turn boundary — a spent loyalty ability never refreshes")
	}
	if tally := g.CastTallyFor(me.ID); tally.Total != 0 {
		t.Errorf("SpellsCastThisTurn survived the turn boundary: %+v", tally)
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("LandsPlayedThisTurn survived the turn boundary: %d", n)
	}
}
