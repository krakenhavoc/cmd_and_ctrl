package game

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// newActiveGame returns a started 2-player game seeded from a fixed
// RNG. Shared helper for mutations_test.
func newActiveGame(t *testing.T) *Game {
	t.Helper()
	g := NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Commander %d", i+1))); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func TestDrawCard(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	libBefore := p.Library.Size()
	handBefore := p.Hand.Size()

	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}

	if p.Library.Size() != libBefore-1 {
		t.Errorf("library size: got %d, want %d", p.Library.Size(), libBefore-1)
	}
	if p.Hand.Size() != handBefore+1 {
		t.Errorf("hand size: got %d, want %d", p.Hand.Size(), handBefore+1)
	}
}

func TestDrawCardUnknownPlayer(t *testing.T) {
	g := newActiveGame(t)
	if err := g.DrawCard(uuid.New()); err != ErrPlayerNotFound {
		t.Errorf("unknown player: got %v, want ErrPlayerNotFound", err)
	}
}

func TestDrawCardEmptyLibrary(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	p.Library.Cards = nil
	if err := g.DrawCard(p.ID); err != ErrZoneEmpty {
		t.Errorf("empty library: got %v, want ErrZoneEmpty", err)
	}
}

func TestDrawCardRequiresActive(t *testing.T) {
	g := NewGame()
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	_, _ = g.AddPlayer("B", buildTestDeck("C"))
	if err := g.DrawCard(g.Seats[0].ID); err != ErrGameNotActive {
		t.Errorf("lobby: got %v, want ErrGameNotActive", err)
	}
}

func TestPlayCardNotInHand(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Bogus card ID that isn't in the player's hand.
	if err := g.PlayCard(p.ID, uuid.New()); err != ErrCardNotFound {
		t.Errorf("bogus card: got %v, want ErrCardNotFound", err)
	}
	// Card actually sitting on the top of the LIBRARY (not hand) —
	// should also fail because PlayCard moves from hand only.
	libCard, _ := p.Library.Top()
	if err := g.PlayCard(p.ID, libCard.InstanceID); err != ErrCardNotFound {
		t.Errorf("card in library (not hand): got %v, want ErrCardNotFound", err)
	}
}

func TestMoveCardByIDSrcEqualsDstPreservesState(t *testing.T) {
	// Regression for R5: move_card with src == dst must NOT clear
	// counters or tapped state (the underlying MoveCard path clears
	// battlefield state on leave, which would wipe a permanent's
	// counters if a client issued a redundant no-op move).
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)
	_ = g.TapCard(card.InstanceID, true)
	_ = g.AddCounter(card.InstanceID, "+1/+1", 3)

	// No-op move on battlefield.
	err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneBattlefield},
		card.InstanceID,
	)
	if err != nil {
		t.Fatalf("no-op move: %v", err)
	}
	// Find the card on the battlefield and verify state is preserved.
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != card.InstanceID {
			continue
		}
		if !c.Tapped {
			t.Error("no-op move cleared Tapped state")
		}
		if c.Counters["+1/+1"] != 3 {
			t.Errorf("no-op move clobbered counters: %v", c.Counters)
		}
	}
}

func TestMoveCardByIDSrcEqualsDstBogusCard(t *testing.T) {
	// A no-op move of a card that isn't actually in the source zone
	// should still return ErrCardNotFound — the optimization must
	// not hide the missing-card error.
	g := newActiveGame(t)
	err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneBattlefield},
		uuid.New(),
	)
	if err != ErrCardNotFound {
		t.Errorf("bogus same-zone move: got %v, want ErrCardNotFound", err)
	}
}

func TestPlayCardMovesToBattlefield(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Draw a card so we have something in hand.
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	card, _ := p.Hand.Top()

	if err := g.PlayCard(p.ID, card.InstanceID); err != nil {
		t.Fatalf("PlayCard: %v", err)
	}
	if p.Hand.Contains(card.InstanceID) {
		t.Error("card still in hand after PlayCard")
	}
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Error("card not on battlefield after PlayCard")
	}
	// Controller should be stamped to the player who played it.
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Controller != p.ID {
			t.Errorf("controller: got %v, want %v", c.Controller, p.ID)
		}
	}
}

func TestMoveCardByIDSharedToShared(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	// Now move from battlefield to exile.
	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneExile},
		card.InstanceID,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if g.Battlefield.Contains(card.InstanceID) {
		t.Error("card still on battlefield")
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("card not in exile")
	}
}

func TestMoveCardByIDUnknownZone(t *testing.T) {
	g := newActiveGame(t)
	err := g.MoveCardByID(
		ZoneRef{Kind: ZoneKind("bogus")},
		ZoneRef{Kind: ZoneExile},
		uuid.New(),
	)
	if err != ErrZoneNotFound {
		t.Errorf("bogus zone: got %v, want ErrZoneNotFound", err)
	}
}

func TestTapCard(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	if err := g.TapCard(card.InstanceID, true); err != nil {
		t.Fatalf("TapCard true: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && !c.Tapped {
			t.Error("card should be tapped")
		}
	}
	if err := g.TapCard(card.InstanceID, false); err != nil {
		t.Fatalf("TapCard false: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Tapped {
			t.Error("card should be untapped")
		}
	}
}

func TestTapCardNotOnBattlefield(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// A card in the library cannot be tapped.
	lib, _ := p.Library.Top()
	if err := g.TapCard(lib.InstanceID, true); err != ErrCardNotFound {
		t.Errorf("tap card in library: got %v, want ErrCardNotFound", err)
	}
}

func TestUntapAllOnlyControllersCards(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	// Play one card per player, both tapped.
	_ = g.DrawCard(p0.ID)
	c0, _ := p0.Hand.Top()
	_ = g.PlayCard(p0.ID, c0.InstanceID)
	_ = g.TapCard(c0.InstanceID, true)

	_ = g.DrawCard(p1.ID)
	c1, _ := p1.Hand.Top()
	_ = g.PlayCard(p1.ID, c1.InstanceID)
	_ = g.TapCard(c1.InstanceID, true)

	if err := g.UntapAll(p0.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}

	for _, c := range g.Battlefield.Cards {
		switch c.InstanceID {
		case c0.InstanceID:
			if c.Tapped {
				t.Error("p0's card should be untapped")
			}
		case c1.InstanceID:
			if !c.Tapped {
				t.Error("p1's card should still be tapped")
			}
		}
	}
}

func TestPassPriorityIsAdvanceStep(t *testing.T) {
	g := newActiveGame(t)
	startStep := g.Turn.Step
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if g.Turn.Step == startStep {
		t.Errorf("step did not advance: still %q", g.Turn.Step)
	}
}

func TestPassTurnSkipsToNextSeat(t *testing.T) {
	g := newActiveGame(t)
	// Put the cursor mid-turn.
	_, _ = g.AdvanceStep()
	_, _ = g.AdvanceStep()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("seat: got %d, want 1", g.Turn.ActiveSeat)
	}
	if g.Turn.Step != StepUntap {
		t.Errorf("step after PassTurn: got %q, want %q", g.Turn.Step, StepUntap)
	}
}

func TestPassTurnWrapsToNextRound(t *testing.T) {
	g := newActiveGame(t)
	// Only 2 players; two PassTurns should wrap to seat 0 turn 2.
	_ = g.PassTurn()
	_ = g.PassTurn()
	if g.Turn.Number != 2 {
		t.Errorf("turn number: got %d, want 2", g.Turn.Number)
	}
	if g.Turn.ActiveSeat != 0 {
		t.Errorf("seat: got %d, want 0", g.Turn.ActiveSeat)
	}
}

func TestMulliganShufflesHandBack(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Draw an initial hand.
	for range 7 {
		_ = g.DrawCard(p.ID)
	}
	if p.Hand.Size() != 7 {
		t.Fatalf("opening hand: got %d, want 7", p.Hand.Size())
	}

	if err := g.Mulligan(p.ID, 6); err != nil {
		t.Fatalf("Mulligan: %v", err)
	}
	if p.Hand.Size() != 6 {
		t.Errorf("after mulligan: hand size %d, want 6", p.Hand.Size())
	}
	// 99 library cards - 6 drawn = 93 remaining.
	// (The 100-card deck has 1 commander routed to the command zone.)
	if p.Library.Size() != 93 {
		t.Errorf("library size after mulligan: got %d, want 93", p.Library.Size())
	}
}

func TestMulliganRejectsNegative(t *testing.T) {
	g := newActiveGame(t)
	if err := g.Mulligan(g.Seats[0].ID, -1); err != ErrInvalidParam {
		t.Errorf("negative mulligan: got %v, want ErrInvalidParam", err)
	}
}

func TestShuffleLibrary(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	before := append([]Card(nil), p.Library.Cards...)
	if err := g.ShuffleLibrary(p.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	// After a shuffle, at least one position should differ on a
	// 99-card deck (probability of no-op is astronomical).
	diff := false
	for i := range before {
		if before[i].InstanceID != p.Library.Cards[i].InstanceID {
			diff = true
			break
		}
	}
	if !diff {
		t.Error("library unchanged after shuffle")
	}
}

func TestChangePlayerLife(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	got, err := g.ChangePlayerLife(p.ID, -7)
	if err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if got != 33 {
		t.Errorf("life: got %d, want 33", got)
	}
}

func TestAddCounter(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	if err := g.AddCounter(card.InstanceID, "+1/+1", 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID {
			if c.Counters["+1/+1"] != 3 {
				t.Errorf("counter: got %d, want 3", c.Counters["+1/+1"])
			}
		}
	}

	// Negative delta should decrement.
	_ = g.AddCounter(card.InstanceID, "+1/+1", -2)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Counters["+1/+1"] != 1 {
			t.Errorf("counter after -2: got %d, want 1", c.Counters["+1/+1"])
		}
	}

	// Decrementing to zero should delete the counter entry.
	_ = g.AddCounter(card.InstanceID, "+1/+1", -5)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Counters != nil {
			t.Errorf("counters should be nil after decrement to <=0, got %v", c.Counters)
		}
	}
}

func TestAddCounterEmptyNameRejected(t *testing.T) {
	g := newActiveGame(t)
	if err := g.AddCounter(uuid.New(), "", 1); err != ErrInvalidParam {
		t.Errorf("empty name: got %v, want ErrInvalidParam", err)
	}
}

func TestSetCommanderDamage(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	if err := g.SetCommanderDamage(p0.ID, p1.ID, 7); err != nil {
		t.Fatalf("SetCommanderDamage: %v", err)
	}
	if p1.CommanderDamage[p0.ID] != 7 {
		t.Errorf("damage: got %d, want 7", p1.CommanderDamage[p0.ID])
	}
	// Set semantics (not add): resetting to 5 should overwrite.
	_ = g.SetCommanderDamage(p0.ID, p1.ID, 5)
	if p1.CommanderDamage[p0.ID] != 5 {
		t.Errorf("damage after set: got %d, want 5", p1.CommanderDamage[p0.ID])
	}
}

func TestSetCommanderDamageClampsNegative(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	_ = g.SetCommanderDamage(p0.ID, p1.ID, -3)
	if p1.CommanderDamage[p0.ID] != 0 {
		t.Errorf("negative clamped: got %d, want 0", p1.CommanderDamage[p0.ID])
	}
}

func TestSetCommanderDamageValidatesFrom(t *testing.T) {
	g := newActiveGame(t)
	p1 := g.Seats[1]
	ghost := uuid.New()
	if err := g.SetCommanderDamage(ghost, p1.ID, 5); err != ErrPlayerNotFound {
		t.Errorf("unseated from: got %v, want ErrPlayerNotFound", err)
	}
	// Map must not have been touched.
	if _, ok := p1.CommanderDamage[ghost]; ok {
		t.Error("target's CommanderDamage map was polluted by unseated from")
	}
}

func TestSetCommanderDamageValidatesTo(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]
	ghost := uuid.New()
	if err := g.SetCommanderDamage(p0.ID, ghost, 5); err != ErrPlayerNotFound {
		t.Errorf("unseated to: got %v, want ErrPlayerNotFound", err)
	}
}

func TestReadSnapshotConsistency(t *testing.T) {
	g := newActiveGame(t)
	var life int
	var turn Turn
	g.ReadSnapshot(func() {
		life = g.Seats[0].Life
		turn = g.Turn
	})
	if life != StartingLife {
		t.Errorf("life: got %d, want %d", life, StartingLife)
	}
	if turn.Number != 1 || turn.Step != StepUntap {
		t.Errorf("turn: %+v", turn)
	}
}
