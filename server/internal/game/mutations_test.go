package game

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// newActiveGame returns a started 2-player game seeded from a fixed
// RNG. Shared helper for mutations_test.
func newActiveGame(t *testing.T) *Game {
	t.Helper()
	return newActiveGameWithSeats(t, 2)
}

// newFourPlayerActiveGame returns a started 4-player game seeded from
// a fixed RNG. Used by priority-rotation tests where wrap behaviour is
// only meaningful with more than two seats.
func newFourPlayerActiveGame(t *testing.T) *Game {
	t.Helper()
	return newActiveGameWithSeats(t, 4)
}

func newActiveGameWithSeats(t *testing.T, seats int) *Game {
	t.Helper()
	g := NewGame()
	for i := 0; i < seats; i++ {
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

func TestSetBattlefieldPosition(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	if err := g.SetBattlefieldPosition(card.InstanceID, 0.25, 0.75); err != nil {
		t.Fatalf("SetBattlefieldPosition: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID {
			if c.BattleX != 0.25 || c.BattleY != 0.75 {
				t.Errorf("position: got (%v, %v), want (0.25, 0.75)", c.BattleX, c.BattleY)
			}
		}
	}
}

func TestSetBattlefieldPositionClamps(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	// Out-of-range inputs should clamp to [0, 1] rather than error;
	// the wire contract is that clients send fractions and the server
	// defends the invariant without a user-facing failure.
	if err := g.SetBattlefieldPosition(card.InstanceID, -0.5, 2.0); err != nil {
		t.Fatalf("SetBattlefieldPosition: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID {
			if c.BattleX != 0 || c.BattleY != 1 {
				t.Errorf("clamp: got (%v, %v), want (0, 1)", c.BattleX, c.BattleY)
			}
		}
	}
}

func TestSetBattlefieldPositionRejectsNaNAndInf(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	// NaN and infinities must not leak into stored state: Go's
	// encoding/json errors on NaN marshalling, which would poison
	// every subsequent snapshot broadcast. `clampUnit` treats any
	// non-finite input as out-of-range and pulls it to the [0, 1] edge.
	cases := []struct {
		name  string
		x, y  float64
		wantX float64
		wantY float64
	}{
		{"nan_x", math.NaN(), 0.5, 0, 0.5},
		{"nan_y", 0.5, math.NaN(), 0.5, 0},
		{"neg_inf_x", math.Inf(-1), 0.5, 0, 0.5},
		{"pos_inf_y", 0.5, math.Inf(1), 0.5, 1},
		{"upper_clamp_x", 1.5, 0.5, 1, 0.5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := g.SetBattlefieldPosition(card.InstanceID, tc.x, tc.y); err != nil {
				t.Fatalf("SetBattlefieldPosition: %v", err)
			}
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID != card.InstanceID {
					continue
				}
				if math.IsNaN(c.BattleX) || math.IsNaN(c.BattleY) {
					t.Fatalf("stored NaN: (%v, %v)", c.BattleX, c.BattleY)
				}
				if math.IsInf(c.BattleX, 0) || math.IsInf(c.BattleY, 0) {
					t.Fatalf("stored Inf: (%v, %v)", c.BattleX, c.BattleY)
				}
				if c.BattleX != tc.wantX || c.BattleY != tc.wantY {
					t.Errorf("got (%v, %v), want (%v, %v)", c.BattleX, c.BattleY, tc.wantX, tc.wantY)
				}
			}
		})
	}
}

func TestSetBattlefieldPositionNotOnBattlefield(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Card in the library — position on a non-battlefield card is a
	// no-op at best; we return ErrCardNotFound so clients don't silently
	// stamp coordinates on a card that'll be ignored.
	lib, _ := p.Library.Top()
	if err := g.SetBattlefieldPosition(lib.InstanceID, 0.5, 0.5); err != ErrCardNotFound {
		t.Errorf("library card: got %v, want ErrCardNotFound", err)
	}
}

func TestMoveCardClearsBattlefieldPosition(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)
	_ = g.SetBattlefieldPosition(card.InstanceID, 0.33, 0.66)

	// Move to graveyard — BattleX/Y should be cleared alongside Tapped
	// and Counters.
	_ = g.TapCard(card.InstanceID, true)
	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: p.ID},
		card.InstanceID,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == card.InstanceID {
			if c.BattleX != 0 || c.BattleY != 0 {
				t.Errorf("position not cleared: got (%v, %v)", c.BattleX, c.BattleY)
			}
			if c.Tapped {
				t.Error("tapped not cleared on zone exit")
			}
		}
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

func TestPassPriorityRotatesWithoutAdvancingStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	startStep := g.Turn.Step
	for i := 1; i <= 3; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
		if g.Turn.Step != startStep {
			t.Errorf("after %d PassPriority calls: step=%q, want unchanged %q", i, g.Turn.Step, startStep)
		}
		if g.Turn.PriorityHolder != i {
			t.Errorf("after %d PassPriority calls: PriorityHolder=%d, want %d", i, g.Turn.PriorityHolder, i)
		}
	}
}

func TestPassPriorityWrapsAdvancesStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	startStep := g.Turn.Step
	// 4 passes in a 4-player game wrap priority back to the active
	// seat, which auto-advances the step.
	for i := 0; i < 4; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if g.Turn.Step == startStep {
		t.Errorf("step did not advance after wraparound: still %q", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("after wraparound: PriorityHolder=%d, want ActiveSeat=%d",
			g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
}

func TestPriorityHolderResetsOnAdvanceStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Rotate priority halfway, then advance the step directly.
	_ = g.PassPriority()
	_ = g.PassPriority()
	if g.Turn.PriorityHolder == g.Turn.ActiveSeat {
		t.Fatalf("setup: PriorityHolder unexpectedly still equals ActiveSeat=%d",
			g.Turn.ActiveSeat)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("after AdvanceStep: PriorityHolder=%d, want ActiveSeat=%d",
			g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
}

func TestPassTurnResetsPriorityHolder(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	_ = g.PassPriority()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("after PassTurn: PriorityHolder=%d, want ActiveSeat=%d",
			g.Turn.PriorityHolder, g.Turn.ActiveSeat)
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

func TestChangePlayerLifeRecordsHistory(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]

	// Sequence of changes — running total tracks Delta.
	for _, delta := range []int{-3, +1, -2, +5} {
		if _, err := g.ChangePlayerLife(p.ID, delta); err != nil {
			t.Fatalf("ChangePlayerLife(%d): %v", delta, err)
		}
	}
	if len(p.LifeHistory) != 4 {
		t.Fatalf("history length: got %d, want 4", len(p.LifeHistory))
	}
	wantTotals := []int{37, 38, 36, 41}
	for i, want := range wantTotals {
		if p.LifeHistory[i].NewTotal != want {
			t.Errorf("entry %d NewTotal: got %d, want %d",
				i, p.LifeHistory[i].NewTotal, want)
		}
	}
	if p.LifeHistory[3].At.IsZero() {
		t.Error("entry timestamp should be stamped")
	}
}

func TestChangePlayerLifeZeroDeltaIsNoop(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	if _, err := g.ChangePlayerLife(p.ID, 0); err != nil {
		t.Fatalf("ChangePlayerLife(0): %v", err)
	}
	if len(p.LifeHistory) != 0 {
		t.Errorf("zero-delta should not record history; got %d entries", len(p.LifeHistory))
	}
}

func TestChangePlayerLifeHistoryRolloverCap(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Drive the log past MaxLifeHistoryEntries — last N must survive,
	// the oldest entries fall off the front.
	for i := 0; i < MaxLifeHistoryEntries+10; i++ {
		// Alternate +1 / -1 so totals stay near the starting value.
		delta := 1
		if i%2 == 1 {
			delta = -1
		}
		if _, err := g.ChangePlayerLife(p.ID, delta); err != nil {
			t.Fatalf("ChangePlayerLife: %v", err)
		}
	}
	if len(p.LifeHistory) != MaxLifeHistoryEntries {
		t.Errorf("history length: got %d, want %d (cap)",
			len(p.LifeHistory), MaxLifeHistoryEntries)
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
