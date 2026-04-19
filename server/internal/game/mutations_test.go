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

// TestPassPriorityRunsStepEntryHooks regression-tests the bug where
// stepping the cursor forward via PassPriority's wrap branch did not
// fire the auto-resolve / auto-clear hooks that AdvanceStep runs.
// Symptom: a player driving combat exclusively with pass_priority (or
// pass_until_end_of_turn, which loops it) saw combat declarations
// persist past end_combat — arrows would stay drawn for the rest of
// the turn.
//
// The fix extracted runStepEntryHooksLocked from AdvanceStep and
// taught PassPriority to call it on the wrap branch; this test
// covers both hooks (resolve into combat_damage, clear into
// end_combat) using priority wraps only — never AdvanceStep.
func TestPassPriorityRunsStepEntryHooks(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	defender := g.Seats[1]
	startLife := defender.Life
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	// Wrap priority twice in a 2-player game to advance past
	// declare_blockers and into combat_damage. Each wrap = 2 passes.
	wrapStep := func() {
		_ = g.PassPriority()
		_ = g.PassPriority()
	}
	for g.Turn.Step != StepCombatDamage {
		before := g.Turn.Step
		wrapStep()
		if g.Turn.Step == before {
			t.Fatalf("priority-wrap loop didn't progress past %q", before)
		}
	}
	// Auto-resolve should have fired on entering combat_damage.
	if defender.Life != startLife-2 {
		t.Errorf("defender life after pass-priority into combat_damage: got %d, want %d (resolve hook didn't fire)",
			defender.Life, startLife-2)
	}

	// One more wrap moves into end_combat; the clear hook should
	// nil AttackingTarget on every battlefield card.
	for g.Turn.Step != StepEndCombat {
		before := g.Turn.Step
		wrapStep()
		if g.Turn.Step == before {
			t.Fatalf("priority-wrap loop didn't progress past %q", before)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			t.Errorf("AttackingTarget not cleared after pass-priority into end_combat: %v",
				c.AttackingTarget)
		}
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

// playToBattlefield helper: draws a card to the player's hand and
// plays it onto the battlefield, returning the card's instance ID.
// Drawn cards are non-creature placeholders by default — combat
// tests that need a creature should use pushCreatureToBattlefield.
func playToBattlefield(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	c, err := p.Hand.Top()
	if err != nil {
		t.Fatalf("Hand.Top: %v", err)
	}
	if err := g.PlayCard(p.ID, c.InstanceID); err != nil {
		t.Fatalf("PlayCard: %v", err)
	}
	return c.InstanceID
}

// pushCreatureToBattlefield builds a 2/2 creature controlled by p and
// drops it directly onto the battlefield, returning the instance ID.
// Bypasses the draw/play loop so combat tests don't need to wrangle
// a deck full of creatures. The TypeLine satisfies IsCreature; Power
// 2 makes ResolveCombatDamage assertions easy.
func pushCreatureToBattlefield(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Test Creature", p.ID)
	c.TypeLine = "Creature — Test"
	c.Power = 2
	c.Toughness = 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// advanceTo walks the turn cursor forward until reaching the named
// step. Used by combat tests to position the game in the
// declare_attackers / declare_blockers / combat_damage windows
// without each test re-implementing the loop.
func advanceTo(t *testing.T, g *Game, step Step) {
	t.Helper()
	for g.Turn.Step != step {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn == before {
			t.Fatalf("advanceTo loop didn't progress past %v", before)
		}
	}
}

func TestDeclareAttackerSetsTarget(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	advanceTo(t, g, StepDeclareAttackers)
	defender := g.Seats[1].ID

	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == attacker && c.AttackingTarget != defender {
			t.Errorf("AttackingTarget: got %v, want %v", c.AttackingTarget, defender)
		}
	}
}

func TestDeclareAttackerRejectsOutsideStep(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	// Default step is Untap — attackers not legal here.
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != ErrWrongStep {
		t.Errorf("declare in untap: got %v, want ErrWrongStep", err)
	}
}

func TestDeclareAttackerRejectsNonCreature(t *testing.T) {
	g := newActiveGame(t)
	// Non-creature: NewCard with no TypeLine.
	land := NewCard("Some Land", g.Seats[0].ID)
	land.TypeLine = "Land"
	g.Battlefield.PushTop(land)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(land.InstanceID, g.Seats[1].ID); err != ErrNotACreature {
		t.Errorf("non-creature attacker: got %v, want ErrNotACreature", err)
	}
}

func TestDeclareAttackerRejectsUnknownTarget(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, uuid.New()); err != ErrPlayerNotFound {
		t.Errorf("unknown target: got %v, want ErrPlayerNotFound", err)
	}
}

func TestDeclareAttackerRejectsUnknownCard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(uuid.New(), g.Seats[1].ID); err != ErrCardNotFound {
		t.Errorf("unknown attacker: got %v, want ErrCardNotFound", err)
	}
}

func TestDeclareAttackerOverwritesPreviousTarget(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("first DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("second DeclareAttacker: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == attacker && c.AttackingTarget != g.Seats[2].ID {
			t.Errorf("AttackingTarget: got %v, want seat 2 %v",
				c.AttackingTarget, g.Seats[2].ID)
		}
	}
}

func TestDeclareBlockerSetsTarget(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	blocker := pushCreatureToBattlefield(t, g, g.Seats[1])
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == blocker && c.BlockingTarget != attacker {
			t.Errorf("BlockingTarget: got %v, want %v", c.BlockingTarget, attacker)
		}
	}
}

func TestDeclareBlockerRejectsOutsideStep(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	blocker := pushCreatureToBattlefield(t, g, g.Seats[1])
	advanceTo(t, g, StepDeclareAttackers)
	_ = g.DeclareAttacker(attacker, g.Seats[1].ID)
	// Still in declare_attackers — blocking not legal here.
	if err := g.DeclareBlocker(blocker, attacker); err != ErrWrongStep {
		t.Errorf("block in declare_attackers: got %v, want ErrWrongStep", err)
	}
}

func TestDeclareBlockerRejectsMissingAttacker(t *testing.T) {
	g := newActiveGame(t)
	blocker := pushCreatureToBattlefield(t, g, g.Seats[1])
	advanceTo(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, uuid.New()); err != ErrCardNotFound {
		t.Errorf("missing attacker: got %v, want ErrCardNotFound", err)
	}
}

func TestClearCombatResetsAllAttackersAndBlockers(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	blocker := pushCreatureToBattlefield(t, g, g.Seats[1])
	advanceTo(t, g, StepDeclareAttackers)
	_ = g.DeclareAttacker(attacker, g.Seats[1].ID)
	advanceTo(t, g, StepDeclareBlockers)
	_ = g.DeclareBlocker(blocker, attacker)

	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			t.Errorf("card %v still attacking %v after ClearCombat", c.Name, c.AttackingTarget)
		}
		if c.BlockingTarget != uuid.Nil {
			t.Errorf("card %v still blocking %v after ClearCombat", c.Name, c.BlockingTarget)
		}
	}
}

func TestZoneExitClearsCombatState(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	src := ZoneRef{Kind: ZoneBattlefield}
	dst := ZoneRef{Kind: ZoneGraveyard, Owner: g.Seats[0].ID}
	if err := g.MoveCardByID(src, dst, attacker); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	for _, c := range g.Seats[0].Graveyard.Cards {
		if c.InstanceID == attacker && c.AttackingTarget != uuid.Nil {
			t.Errorf("AttackingTarget not cleared on zone exit: %v", c.AttackingTarget)
		}
	}
}

func TestResolveCombatDamageUnblockedAppliesPower(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	defender := g.Seats[1]
	startLife := defender.Life
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// Skip declare_blockers (no blocker declared).
	advanceTo(t, g, StepCombatDamage)
	// AdvanceStep into combat_damage triggered the auto-resolve.
	if defender.Life != startLife-2 {
		t.Errorf("defender life: got %d, want %d (took 2 unblocked)",
			defender.Life, startLife-2)
	}
	if len(defender.LifeHistory) == 0 {
		t.Error("life history: expected one entry from combat damage")
	}
	// Combat declarations persist through the combat_damage step so
	// the client can keep its arrows drawn while damage shows up.
	stillAttacking := false
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			stillAttacking = true
		}
	}
	if !stillAttacking {
		t.Error("AttackingTarget cleared inside combat_damage; expected persistence until end_combat")
	}
	// One more advance moves the cursor into end_combat, which IS
	// where the clear should fire.
	advanceTo(t, g, StepEndCombat)
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			t.Error("AttackingTarget not cleared after entering end_combat")
		}
	}
}

func TestResolveCombatDamageBlockedSpared(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	blocker := pushCreatureToBattlefield(t, g, g.Seats[1])
	defender := g.Seats[1]
	startLife := defender.Life
	advanceTo(t, g, StepDeclareAttackers)
	_ = g.DeclareAttacker(attacker, defender.ID)
	advanceTo(t, g, StepDeclareBlockers)
	_ = g.DeclareBlocker(blocker, attacker)
	advanceTo(t, g, StepCombatDamage)
	if defender.Life != startLife {
		t.Errorf("blocked attacker dealt damage: life %d, want unchanged %d",
			defender.Life, startLife)
	}
}

func TestResolveCombatDamageRespectsCounters(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCreatureToBattlefield(t, g, g.Seats[0])
	// Slap a +1/+1 counter on it (base 2 → 3).
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attacker {
			g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 1}
		}
	}
	defender := g.Seats[1]
	startLife := defender.Life
	advanceTo(t, g, StepDeclareAttackers)
	_ = g.DeclareAttacker(attacker, defender.ID)
	advanceTo(t, g, StepCombatDamage)
	if defender.Life != startLife-3 {
		t.Errorf("counter-boosted attacker damage: life %d, want %d",
			defender.Life, startLife-3)
	}
}

func TestCurrentPowerClampsToZero(t *testing.T) {
	c := NewCard("Test", uuid.New())
	c.Power = 2
	c.Counters = map[string]int{"-1/-1": 5}
	if got := c.CurrentPower(); got != 0 {
		t.Errorf("CurrentPower with -1/-1 swing: got %d, want 0 (clamped)", got)
	}
}

func TestStartDealsOpeningHand(t *testing.T) {
	g := newActiveGame(t) // already started
	for _, p := range g.Seats {
		if p.Hand.Size() != OpeningHandSize {
			t.Errorf("seat %d hand size: got %d, want %d", p.Seat, p.Hand.Size(), OpeningHandSize)
		}
		if p.HandKept {
			t.Errorf("seat %d HandKept: got true, want false (window just opened)", p.Seat)
		}
	}
	if !g.MulligansOpen {
		t.Error("MulligansOpen: got false, want true after Start")
	}
}

func TestKeepHandClosesWindowWhenAllCommit(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for i, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand seat %d: %v", i, err)
		}
		if !p.HandKept {
			t.Errorf("seat %d HandKept: not set", i)
		}
		// Window stays open until the LAST seat commits.
		wantOpen := i < len(g.Seats)-1
		if g.MulligansOpen != wantOpen {
			t.Errorf("after %d/%d kept: MulligansOpen=%v, want %v",
				i+1, len(g.Seats), g.MulligansOpen, wantOpen)
		}
	}
}

func TestKeepHandIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	if err := g.KeepHand(p.ID); err != nil {
		t.Fatalf("first KeepHand: %v", err)
	}
	// Repeat — must not error and must not flip back.
	if err := g.KeepHand(p.ID); err != nil {
		t.Errorf("second KeepHand: got %v, want nil (idempotent)", err)
	}
	if !p.HandKept {
		t.Error("HandKept regressed to false on repeat call")
	}
}

func TestMulliganResetsKeptAndCountsTaken(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	if err := g.KeepHand(p.ID); err != nil {
		t.Fatalf("KeepHand: %v", err)
	}
	// Player changed their mind and called mulligan.
	if err := g.Mulligan(p.ID, OpeningHandSize); err != nil {
		t.Fatalf("Mulligan: %v", err)
	}
	if p.HandKept {
		t.Error("HandKept should be false after Mulligan (commitment reset)")
	}
	if p.MulligansTaken != 1 {
		t.Errorf("MulligansTaken: got %d, want 1", p.MulligansTaken)
	}
	// Two more mulligans → count = 3.
	_ = g.Mulligan(p.ID, OpeningHandSize)
	_ = g.Mulligan(p.ID, OpeningHandSize)
	if p.MulligansTaken != 3 {
		t.Errorf("MulligansTaken: got %d, want 3", p.MulligansTaken)
	}
}

func TestKeepHandRejectsEliminated(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	// Concede in a 2-player game ends the game; KeepHand on lobby/
	// ended state should fail with ErrGameNotActive — that's the
	// outermost guard. Use a 4-player game so the game stays active.
	g4 := newFourPlayerActiveGame(t)
	p4 := g4.Seats[0]
	if err := g4.Concede(p4.ID); err != nil {
		t.Fatalf("Concede (4p): %v", err)
	}
	if err := g4.KeepHand(p4.ID); err != ErrPlayerEliminated {
		t.Errorf("KeepHand on eliminated: got %v, want ErrPlayerEliminated", err)
	}
}

func TestConcedeMarksPlayerEliminated(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	p := g.Seats[2]
	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !p.Eliminated {
		t.Error("Eliminated flag not set")
	}
	if g.State != StateActive {
		t.Errorf("state: got %q, want active (3 players still alive)", g.State)
	}
}

func TestConcedeIsRejectedWhenAlreadyEliminated(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	p := g.Seats[2]
	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("first Concede: %v", err)
	}
	if err := g.Concede(p.ID); err != ErrPlayerEliminated {
		t.Errorf("second Concede: got %v, want ErrPlayerEliminated", err)
	}
}

func TestConcedeAdvancesPastEliminatedActiveSeat(t *testing.T) {
	// Active seat == 0 at start. Conceding seat 0 must move the
	// cursor to seat 1 so play continues.
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("active seat: got %d, want 1", g.Turn.ActiveSeat)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Errorf("priority holder: got %d, want 1", g.Turn.PriorityHolder)
	}
	if g.Turn.Step != StepUntap {
		t.Errorf("step after pass-past: got %q, want untap", g.Turn.Step)
	}
}

func TestConcedeEndsGameWhenOneSurvivor(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Three concessions in a 4-player game leaves one survivor.
	for i := 0; i < 3; i++ {
		if err := g.Concede(g.Seats[i].ID); err != nil {
			t.Fatalf("Concede seat %d: %v", i, err)
		}
	}
	if g.State != StateEnded {
		t.Errorf("state: got %q, want ended", g.State)
	}
	if g.Seats[3].Eliminated {
		t.Error("survivor wrongly marked eliminated")
	}
}

func TestConcedeFromLobbyRejected(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Cmdr %d", i+1)))
	}
	// Game not started.
	if err := g.Concede(g.Seats[0].ID); err != ErrGameNotActive {
		t.Errorf("Concede in lobby: got %v, want ErrGameNotActive", err)
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
	// As of S08, Start deals an opening hand of OpeningHandSize.
	if p.Hand.Size() != OpeningHandSize {
		t.Fatalf("opening hand: got %d, want %d", p.Hand.Size(), OpeningHandSize)
	}

	if err := g.Mulligan(p.ID, 6); err != nil {
		t.Fatalf("Mulligan: %v", err)
	}
	if p.Hand.Size() != 6 {
		t.Errorf("after mulligan: hand size %d, want 6", p.Hand.Size())
	}
	// 99 library cards minus 6 redrawn = 93 remaining. The opening
	// hand was returned to the library before the mulligan redraw.
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

func TestSetMonarchAndInitiative(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	if err := g.SetMonarch(p0.ID); err != nil {
		t.Fatalf("SetMonarch: %v", err)
	}
	if g.Monarch != p0.ID {
		t.Errorf("monarch: got %v, want %v", g.Monarch, p0.ID)
	}
	// Reassigning to a different seat overwrites.
	if err := g.SetMonarch(p1.ID); err != nil {
		t.Fatalf("SetMonarch reassign: %v", err)
	}
	if g.Monarch != p1.ID {
		t.Errorf("monarch reassign: got %v, want %v", g.Monarch, p1.ID)
	}
	// uuid.Nil clears.
	if err := g.SetMonarch(uuid.Nil); err != nil {
		t.Fatalf("SetMonarch clear: %v", err)
	}
	if g.Monarch != uuid.Nil {
		t.Errorf("monarch cleared: got %v, want nil", g.Monarch)
	}
	// Unknown player is rejected.
	if err := g.SetMonarch(uuid.New()); err != ErrPlayerNotFound {
		t.Errorf("monarch unknown: got %v, want ErrPlayerNotFound", err)
	}

	// Initiative behaves the same.
	if err := g.SetInitiative(p0.ID); err != nil {
		t.Fatalf("SetInitiative: %v", err)
	}
	if g.Initiative != p0.ID {
		t.Errorf("initiative: got %v, want %v", g.Initiative, p0.ID)
	}
}

func TestSetGoaded(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	c := Card{InstanceID: uuid.New(), Name: "Atog", TypeLine: "Creature — Atog", Owner: p0.ID, Controller: p0.ID, Power: 1, Toughness: 2}
	g.Battlefield.PushTop(c)

	if err := g.SetGoaded(c.InstanceID, p1.ID); err != nil {
		t.Fatalf("SetGoaded: %v", err)
	}
	if g.Battlefield.Cards[0].GoadedBy != p1.ID {
		t.Errorf("goaded_by: got %v, want %v", g.Battlefield.Cards[0].GoadedBy, p1.ID)
	}
	// Clearing.
	_ = g.SetGoaded(c.InstanceID, uuid.Nil)
	if g.Battlefield.Cards[0].GoadedBy != uuid.Nil {
		t.Errorf("goaded cleared: got %v, want nil", g.Battlefield.Cards[0].GoadedBy)
	}
	// Unknown card.
	if err := g.SetGoaded(uuid.New(), p1.ID); err != ErrCardNotFound {
		t.Errorf("unknown card: got %v, want ErrCardNotFound", err)
	}
}

func TestSetGoadedClearsOnZoneExit(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	c := Card{InstanceID: uuid.New(), Name: "Atog", TypeLine: "Creature — Atog", Owner: p0.ID, Controller: p0.ID, Power: 1, Toughness: 2}
	g.Battlefield.PushTop(c)
	_ = g.SetGoaded(c.InstanceID, p1.ID)

	if _, err := MoveCard(g.Battlefield, p0.Graveyard, c.InstanceID); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if p0.Graveyard.Cards[0].GoadedBy != uuid.Nil {
		t.Errorf("goaded persisted across zone exit: got %v", p0.Graveyard.Cards[0].GoadedBy)
	}
}

func TestSetPoisonAndEnergy(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]

	if err := g.SetPoison(p0.ID, 4); err != nil {
		t.Fatalf("SetPoison: %v", err)
	}
	if p0.Poison != 4 {
		t.Errorf("poison: got %d, want 4", p0.Poison)
	}
	// Negative clamps to 0 (set semantics, not increment).
	_ = g.SetPoison(p0.ID, -3)
	if p0.Poison != 0 {
		t.Errorf("poison clamped: got %d, want 0", p0.Poison)
	}
	if err := g.SetEnergy(p0.ID, 12); err != nil {
		t.Fatalf("SetEnergy: %v", err)
	}
	if p0.Energy != 12 {
		t.Errorf("energy: got %d, want 12", p0.Energy)
	}
	// Unknown player.
	if err := g.SetPoison(uuid.New(), 1); err != ErrPlayerNotFound {
		t.Errorf("poison unknown: got %v, want ErrPlayerNotFound", err)
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
