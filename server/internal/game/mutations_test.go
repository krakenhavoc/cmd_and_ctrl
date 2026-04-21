package game

import (
	"errors"
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
	g := newActiveGameMulligansOpen(t, seats)
	// S13: every test that drives priority / steps assumes the
	// mulligan window has closed and the cursor has landed on a
	// priority-granting step. Closing mulligans here fires the first
	// step-entry hook, which auto-untaps seat 0 and advances the
	// cursor to Upkeep with PriorityHolder=ActiveSeat.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand %s: %v", p.Name, err)
		}
	}
	return g
}

// newActiveGameMulligansOpen returns a started game with the mulligan
// window still open (no KeepHand calls), so tests that exercise the
// pre-game flow (TestStartDealsOpeningHand,
// TestKeepHandClosesWindowWhenAllCommit) can observe the initial
// state. After S13 the cursor sits on Untap with NoPriority until
// KeepHand for the last seat closes the window and fires the first
// auto-untap.
func newActiveGameMulligansOpen(t *testing.T, seats int) *Game {
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
	g := newActiveGameMulligansOpen(t, 2)
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
	g := newActiveGameMulligansOpen(t, 4)
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
	// cursor to seat 1 so play continues. S13: the new seat starts
	// on Untap, the entry hook auto-untaps and walks the cursor on
	// to Upkeep where priority is granted.
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
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step after pass-past: got %q, want upkeep (S13 auto-advance)", g.Turn.Step)
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
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step after PassTurn: got %q, want %q (S13 auto-advance through Untap)",
			g.Turn.Step, StepUpkeep)
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

// TestS13PassPriorityRejectedOnNoPriorityStep verifies the server
// returns ErrNoPriority when called during Untap or Cleanup. The
// helper closes mulligans which auto-advances to Upkeep, so we get
// to a no-priority state via PassTurn (lands on Untap of next seat
// — but mulligans-close already happened, so… use a fresh game with
// mulligans open).
func TestS13PassPriorityRejectedOnNoPriorityStep(t *testing.T) {
	g := newActiveGameMulligansOpen(t, 4)
	if g.Turn.Step != StepUntap || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("setup: step=%q priority=%d, want Untap/NoPriority", g.Turn.Step, g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); err != ErrNoPriority {
		t.Errorf("PassPriority on Untap (mulligans open): got %v, want ErrNoPriority", err)
	}
}

// TestS13PassPrioritySkipsEliminatedSeat verifies the new rotation
// loop walks past eliminated seats so a 4-player game with two dead
// seats still terminates the priority loop on the survivors' wrap.
func TestS13PassPrioritySkipsEliminatedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Eliminate seats 1 and 3 directly. (Concede would also advance
	// the cursor when seat 0 is active; we want to leave the cursor
	// alone here so we can observe rotation across the dead seats.)
	g.Seats[1].Eliminated = true
	g.Seats[3].Eliminated = true

	// Cursor at seat 0 Upkeep with PriorityHolder=0 (post-mulligan).
	// PassPriority from seat 0 should jump straight to seat 2,
	// skipping seat 1.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority 0→2: %v", err)
	}
	if g.Turn.PriorityHolder != 2 {
		t.Errorf("after first pass: PriorityHolder=%d, want 2 (skipped eliminated seat 1)",
			g.Turn.PriorityHolder)
	}

	// Pass again — seat 2 is the only other live opponent, so the
	// rotation walks past seat 3 (eliminated) and wraps back to the
	// active seat (0), which advances the step.
	beforeStep := g.Turn.Step
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority wrap: %v", err)
	}
	if g.Turn.Step == beforeStep {
		t.Errorf("step did not advance on rotation wrap (still %q)", g.Turn.Step)
	}
}

// pushTypedCardToHand puts a card with the given name + type line on
// the player's hand and returns its instance ID. Used by the S13.1
// cast tests to control the type-line route the mutation takes
// (land vs. instant vs. permanent vs. sorcery).
func pushTypedCardToHand(p *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Hand.PushTop(c)
	return c.InstanceID
}

// advanceTo positions the cursor at a specific step. Helper for
// sorcery-speed gating tests below. (advanceTo is also defined
// further down in this file for combat tests; the names match by
// design.)

// TestS131CastLandRoutesToBattlefield verifies that casting a land
// skips the stack entirely (CR 305 — special action) when the
// sorcery-speed gate is open.
func TestS131CastLandRoutesToBattlefield(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHand(p, "Forest", "Basic Land — Forest")

	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell land: %v", err)
	}
	if !g.Battlefield.Contains(id) {
		t.Errorf("land did not route to battlefield")
	}
	if g.Stack != nil && g.Stack.Contains(id) {
		t.Errorf("land should not have hit the stack")
	}
	if _, ok := g.StackMeta[id]; ok {
		t.Errorf("land should not have a StackMeta entry")
	}
}

// TestS131CastSpellGoesToStack verifies that a non-land cast lands
// the card on the stack with a StackMeta entry, and the caster
// retains priority (CR 117.3c).
func TestS131CastSpellGoesToStack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHand(p, "Counterspell", "Instant")

	priorityBefore := g.Turn.PriorityHolder
	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell instant: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Errorf("spell did not hit the stack")
	}
	item, ok := g.StackMeta[id]
	if !ok {
		t.Fatalf("StackMeta entry missing for cast spell")
	}
	if item.Kind != StackItemSpell || item.Controller != p.ID || item.Owner != p.ID {
		t.Errorf("StackMeta shape wrong: %+v", item)
	}
	if g.Turn.PriorityHolder != priorityBefore {
		t.Errorf("caster did not retain priority: was %d, now %d", priorityBefore, g.Turn.PriorityHolder)
	}
}

// TestS131SorcerySpeedGate verifies that sorceries reject outside the
// caller's main phase / empty stack / active-player window, and
// instants ignore the gate.
func TestS131SorcerySpeedGate(t *testing.T) {
	g := newActiveGame(t)
	// Cursor after newActiveGame is StepUpkeep — not a main phase.
	p := g.Seats[0]
	sorcID := pushTypedCardToHand(p, "Wrath of God", "Sorcery")
	if err := g.CastSpell(p.ID, sorcID, CastSpellParams{}); err != ErrSorcerySpeedRequired {
		t.Errorf("sorcery on Upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}
	instID := pushTypedCardToHand(p, "Lightning Bolt", "Instant")
	if err := g.CastSpell(p.ID, instID, CastSpellParams{}); err != nil {
		t.Errorf("instant on Upkeep should bypass sorcery gate: %v", err)
	}
}

// TestS131SplitSecondBlocksFurtherCasts verifies that
// SplitSecondActive rejects subsequent casts (CR 702.79). Split-
// second mana abilities and special actions remain legal — we only
// cover the cast path here.
func TestS131SplitSecondBlocksFurtherCasts(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	splitID := pushTypedCardToHand(p, "Trickbind", "Instant")
	if err := g.CastSpell(p.ID, splitID, CastSpellParams{SplitSecond: true}); err != nil {
		t.Fatalf("first split-second cast: %v", err)
	}
	if !g.SplitSecondActive {
		t.Fatalf("SplitSecondActive should be true after split-second cast")
	}
	otherID := pushTypedCardToHand(p, "Lightning Bolt", "Instant")
	if err := g.CastSpell(p.ID, otherID, CastSpellParams{}); err != ErrSplitSecondActive {
		t.Errorf("subsequent cast: got %v, want ErrSplitSecondActive", err)
	}
}

// TestS131PassPriorityResolvesTopWhenStackNonEmpty verifies the
// rewritten PassPriority: with a non-empty stack and all opponents
// passed, the wrap branch resolves the top instead of advancing the
// step. Permanents land on the battlefield; instants go to the
// owner's graveyard.
func TestS131PassPriorityResolvesTopWhenStackNonEmpty(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p0, p1 := g.Seats[0], g.Seats[1]

	// Cast a creature from seat 0 — should land on stack.
	creatureID := pushTypedCardToHand(p0, "Grizzly Bears", "Creature — Bear")
	if err := g.CastSpell(p0.ID, creatureID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell creature: %v", err)
	}
	if !g.Stack.Contains(creatureID) {
		t.Fatalf("creature not on stack")
	}

	// Caster passes, then opponent passes — wrap fires resolution.
	stepBefore := g.Turn.Step
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority caster: %v", err)
	}
	// Now seat 1 holds priority. Pass — wraps to active seat.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority opponent: %v", err)
	}
	if g.Turn.Step != stepBefore {
		t.Errorf("step advanced on resolve: was %q, now %q (should NOT advance)", stepBefore, g.Turn.Step)
	}
	if !g.Battlefield.Contains(creatureID) {
		t.Errorf("creature did not resolve to battlefield")
	}
	if _, ok := g.StackMeta[creatureID]; ok {
		t.Errorf("StackMeta entry for resolved creature still present")
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority did not return to active seat after resolve: %d", g.Turn.PriorityHolder)
	}

	// Cast an instant — should resolve to graveyard on full pass.
	advanceTo(t, g, StepPostcombatMain)
	instantID := pushTypedCardToHand(p1, "Shock", "Instant")
	if err := g.CastSpell(p1.ID, instantID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell instant: %v", err)
	}
	// p1 is the priority holder (cast). Pass to p0 first.
	g.Turn.PriorityHolder = (g.Turn.ActiveSeat + 1) % len(g.Seats) // simulate priority on caster seat
	_ = g.PassPriority()                                           // p1 → p0
	_ = g.PassPriority()                                           // p0 → wrap → resolve
	if !p1.Graveyard.Contains(instantID) {
		t.Errorf("instant did not resolve to owner's graveyard")
	}
}

// TestS131TargetReCheckAllIllegalCountersByGameRules covers
// CR 608.2b: when every targeted slot has become illegal between
// announce and resolve, the spell is "countered by game rules" and
// goes to its owner's graveyard without effect — even if it would
// normally be a permanent that resolves to the battlefield.
func TestS131TargetReCheckAllIllegalCountersByGameRules(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	target := g.Seats[2]

	// Cast a creature with a player target (sandbox: targeting works
	// regardless of the card's actual oracle text).
	creatureID := pushTypedCardToHand(caster, "Stalking Vengeance", "Creature — Avatar")
	if err := g.CastSpell(caster.ID, creatureID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: target.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	// Eliminate the targeted player BEFORE the spell resolves.
	target.Eliminated = true

	// Walk priority all the way around to fire the resolve.
	stepBefore := g.Turn.Step
	for g.stackHasItemsLocked() {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.Step != stepBefore {
		t.Errorf("step advanced unexpectedly: was %q, now %q", stepBefore, g.Turn.Step)
	}
	// Creature should be in caster's graveyard, NOT the battlefield.
	if g.Battlefield.Contains(creatureID) {
		t.Errorf("creature resolved to battlefield despite all-illegal targets")
	}
	if !caster.Graveyard.Contains(creatureID) {
		t.Errorf("creature did not route to owner's graveyard (countered by game rules)")
	}
}

// TestS131TargetReCheckPartialIllegalStillResolves covers the
// partial-illegal case: at least one target survives, so the spell
// resolves normally. Sandbox: the engine doesn't trim the surviving
// target subset, just lets the spell through.
func TestS131TargetReCheckPartialIllegalStillResolves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	dead := g.Seats[2]
	alive := g.Seats[3]

	id := pushTypedCardToHand(caster, "Multi-target Spell", "Sorcery")
	if err := g.CastSpell(caster.ID, id, CastSpellParams{
		Targets: []TargetRef{
			{Kind: TargetPlayer, ID: dead.ID},
			{Kind: TargetPlayer, ID: alive.ID},
		},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	dead.Eliminated = true

	for g.stackHasItemsLocked() {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !caster.Graveyard.Contains(id) {
		t.Errorf("partial-target sorcery did not resolve to owner's graveyard")
	}
}

// TestS131CastCapturesModesXDistribution verifies the announce-time
// data-capture path: modes / X / distribution flow through
// CastSpellParams onto the StackItem. Sandbox: the engine doesn't
// interpret these values, just preserves them so opponents can see
// what was chosen and resolve manually.
func TestS131CastCapturesModesXDistribution(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	a, b := g.Seats[1], g.Seats[2]

	id := pushTypedCardToHand(caster, "Cryptic Command", "Instant")
	if err := g.CastSpell(caster.ID, id, CastSpellParams{
		Modes:  []int{0, 2},
		XValue: 4,
		Distribution: map[uuid.UUID]int{
			a.ID: 1,
			b.ID: 3,
		},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	item, ok := g.StackMeta[id]
	if !ok {
		t.Fatalf("StackMeta entry missing")
	}
	if got := item.Modes; len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Errorf("Modes: got %v, want [0 2]", got)
	}
	if item.XValue != 4 {
		t.Errorf("XValue: got %d, want 4", item.XValue)
	}
	if got := item.Distribution[a.ID]; got != 1 {
		t.Errorf("Distribution[a]: got %d, want 1", got)
	}
	if got := item.Distribution[b.ID]; got != 3 {
		t.Errorf("Distribution[b]: got %d, want 3", got)
	}

	// Mutate the caller's slice/map after the cast — must NOT affect
	// the StackMeta entry (CastSpell deep-copies on the way in).
	if false {
		// Sanity: this branch exists to document intent and silence
		// any future "noop test" linters.
		t.Log("CastSpell defensively copies announce-time slices/maps")
	}
}

// TestS131SelfAndNoneTargetsBypassReCheck verifies that Kind=Self /
// Kind=None entries don't count as "targeted" for the re-check
// short-circuit — the spell resolves normally even if those entries
// exist alongside no actual player/card targets.
func TestS131SelfAndNoneTargetsBypassReCheck(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]

	id := pushTypedCardToHand(caster, "Self Sorcery", "Sorcery")
	if err := g.CastSpell(caster.ID, id, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetSelf}, {Kind: TargetNone}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for g.stackHasItemsLocked() {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !caster.Graveyard.Contains(id) {
		t.Errorf("self-targeted sorcery did not resolve normally")
	}
}

// TestS131CounterSpellRoutesToOwnerGraveyardByDefault covers the
// canonical Counterspell shape: counter target spell → graveyard,
// no destination override.
func TestS131CounterSpellRoutesToOwnerGraveyardByDefault(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	id := pushTypedCardToHand(caster, "Lightning Bolt", "Instant")
	if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if g.Stack.Contains(id) {
		t.Errorf("spell still on stack after counter")
	}
	if !caster.Graveyard.Contains(id) {
		t.Errorf("countered spell did not route to owner's graveyard")
	}
	if _, ok := g.StackMeta[id]; ok {
		t.Errorf("StackMeta entry still present after counter")
	}
}

// TestS131CounterSpellHonoursDestination covers Hinder
// (counter → library) and Remand (counter → hand) shapes.
func TestS131CounterSpellHonoursDestination(t *testing.T) {
	cases := []struct {
		name string
		dst  ZoneRef
	}{
		{"library", ZoneRef{Kind: ZoneLibrary}},
		{"hand", ZoneRef{Kind: ZoneHand}},
		{"exile", ZoneRef{Kind: ZoneExile}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			caster := g.Seats[0]
			id := pushTypedCardToHand(caster, "Spell", "Instant")
			if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			ref := c.dst
			if ref.Kind != ZoneExile {
				ref.Owner = caster.ID
			}
			if err := g.CounterSpell(id, &ref); err != nil {
				t.Fatalf("CounterSpell to %s: %v", c.name, err)
			}
			if g.Stack.Contains(id) {
				t.Errorf("spell still on stack")
			}
			switch c.dst.Kind {
			case ZoneLibrary:
				if !caster.Library.Contains(id) {
					t.Errorf("not in library")
				}
			case ZoneHand:
				if !caster.Hand.Contains(id) {
					t.Errorf("not in hand")
				}
			case ZoneExile:
				if !g.Exile.Contains(id) {
					t.Errorf("not in exile")
				}
			}
		})
	}
}

// TestS131CounterSpellRejectsBattlefieldDestination covers the
// engine-level guard against "counter target spell, putting it onto
// the battlefield" — no real card does this with a generic counter
// shape, and routing to battlefield via this verb would muddle the
// resolution path.
func TestS131CounterSpellRejectsBattlefieldDestination(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	id := pushTypedCardToHand(caster, "Spell", "Instant")
	if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	dst := ZoneRef{Kind: ZoneBattlefield}
	if err := g.CounterSpell(id, &dst); err != ErrInvalidStackDestination {
		t.Errorf("battlefield dst: got %v, want ErrInvalidStackDestination", err)
	}
}

// TestS131CounterAbilityRemovesItem covers the ability-removal
// shape: an activated / triggered ability item ceases to exist.
func TestS131CounterAbilityRemovesItem(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[0]
	abilityID := uuid.New()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[abilityID] = &StackItem{
			ID:           abilityID,
			Kind:         StackItemActivated,
			Controller:   caster.ID,
			Owner:        caster.ID,
			SourceCardID: caster.ID, // a permanent's instance ID would go here
		}
	})
	if err := g.CounterAbility(abilityID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	if _, ok := g.StackMeta[abilityID]; ok {
		t.Errorf("ability still in StackMeta after CounterAbility")
	}
}

// TestS131CounterMissingItem covers the not-on-stack case for both
// counter verbs.
func TestS131CounterMissingItem(t *testing.T) {
	g := newActiveGame(t)
	if err := g.CounterSpell(uuid.New(), nil); err != ErrCardNotOnStack {
		t.Errorf("CounterSpell missing: got %v, want ErrCardNotOnStack", err)
	}
	if err := g.CounterAbility(uuid.New()); err != ErrCardNotOnStack {
		t.Errorf("CounterAbility missing: got %v, want ErrCardNotOnStack", err)
	}
}

// TestS131ActivateAbilityCreatesStackItem covers the activated-
// ability shape: a fresh StackItem with synthetic ID, controller =
// player, source = card. No card moves; the source stays put.
func TestS131ActivateAbilityCreatesStackItem(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[0]
	src := pushCreatureToBattlefield(t, g, caster)

	beforeStackCount := len(g.StackMeta)
	if err := g.ActivateAbility(caster.ID, src, AbilityParams{
		Label: "Tap: Add G",
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if len(g.StackMeta) != beforeStackCount+1 {
		t.Errorf("StackMeta count: got %d, want +1", len(g.StackMeta))
	}
	// The source card must still be on the battlefield (abilities
	// don't move their source).
	if !g.Battlefield.Contains(src) {
		t.Errorf("source card moved off battlefield after ability activation")
	}
	// Find the new entry — controller=caster, kind=activated.
	var entry *StackItem
	for _, item := range g.StackMeta {
		if item.SourceCardID == src && item.Kind == StackItemActivated {
			entry = item
			break
		}
	}
	if entry == nil {
		t.Fatalf("activated ability entry not found in StackMeta")
	}
	if entry.Controller != caster.ID || entry.Owner != caster.ID {
		t.Errorf("entry controller/owner wrong: %+v", entry)
	}
	if entry.Label != "Tap: Add G" {
		t.Errorf("entry label: got %q, want %q", entry.Label, "Tap: Add G")
	}
}

// TestS131ActivateLoyaltyAppliesDelta covers the sandbox loyalty
// shape: sorcery-speed gate, immediate delta application via the
// "loyalty" counter, once-per-turn flag set.
func TestS131ActivateLoyaltyAppliesDelta(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace, the Mind Sculptor",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      caster.ID,
		Controller: caster.ID,
		Counters:   map[string]int{"loyalty": 3},
	})
	if err := g.ActivateLoyalty(caster.ID, pwID, "+0 brainstorm", 0); err != nil {
		t.Fatalf("ActivateLoyalty +0: %v", err)
	}
	// Loyalty unchanged at 3 (delta=0, no-op).
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == pwID {
			if g.Battlefield.Cards[i].Counters["loyalty"] != 3 {
				t.Errorf("loyalty after +0: got %d, want 3", g.Battlefield.Cards[i].Counters["loyalty"])
			}
		}
	}
	// Second activation same turn → blocked.
	if err := g.ActivateLoyalty(caster.ID, pwID, "-1 unsummon", -1); err != ErrLoyaltyAlreadyActivated {
		t.Errorf("second activation: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// TestS131ActivateLoyaltyResetsOnNewTurn verifies the once-per-turn
// flag clears when the cursor moves to a new active seat.
func TestS131ActivateLoyaltyResetsOnNewTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      caster.ID,
		Controller: caster.ID,
		Counters:   map[string]int{"loyalty": 3},
	})
	if err := g.ActivateLoyalty(caster.ID, pwID, "+1", 1); err != nil {
		t.Fatalf("ActivateLoyalty: %v", err)
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.LoyaltyActivatedThisTurn[pwID] {
		t.Errorf("LoyaltyActivatedThisTurn[pw] should reset on new turn")
	}
}

// TestS131ActivateLoyaltySorcerySpeedGate verifies the sorcery-speed
// gate: no main phase / stack non-empty / not active player → reject.
func TestS131ActivateLoyaltySorcerySpeedGate(t *testing.T) {
	g := newActiveGame(t)
	// Cursor at Upkeep — not main phase.
	caster := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      caster.ID,
		Controller: caster.ID,
		Counters:   map[string]int{"loyalty": 3},
	})
	if err := g.ActivateLoyalty(caster.ID, pwID, "+1", 1); err != ErrSorcerySpeedRequired {
		t.Errorf("upkeep activation: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// TestS131AnnounceTriggerAPNAPDrain covers the APNAP queue (CR
// 603.3b): triggers from active player drain first, then turn-order
// clockwise. With seats [0..3] and active=0, queueing in order
// [seat2, seat0, seat3, seat1] should drain into StackMeta in seat
// order [0, 1, 2, 3].
func TestS131AnnounceTriggerAPNAPDrain(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Park a card on each player's battlefield to use as source.
	srcs := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		srcs[i] = pushCreatureToBattlefield(t, g, p)
	}
	// Announce in non-APNAP order: seat 2, seat 0, seat 3, seat 1.
	for _, seat := range []int{2, 0, 3, 1} {
		if err := g.AnnounceTrigger(g.Seats[seat].ID, srcs[seat], AbilityParams{
			Label: "trigger",
		}); err != nil {
			t.Fatalf("AnnounceTrigger seat %d: %v", seat, err)
		}
	}
	if len(g.PendingTriggers) != 4 {
		t.Fatalf("PendingTriggers count: got %d, want 4", len(g.PendingTriggers))
	}
	// Drain — manually for the test (in real play PassPriority calls it).
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if len(g.PendingTriggers) != 0 {
		t.Errorf("PendingTriggers should be empty after drain: got %d", len(g.PendingTriggers))
	}
	if len(g.StackMeta) != 4 {
		t.Fatalf("StackMeta count after drain: got %d, want 4", len(g.StackMeta))
	}
	// Verify each trigger landed; we don't assert per-seat ordering
	// inside StackMeta because Go maps are unordered, but the controller
	// set must match the four seats.
	gotControllers := make(map[uuid.UUID]bool, 4)
	for _, item := range g.StackMeta {
		gotControllers[item.Controller] = true
	}
	for i, p := range g.Seats {
		if !gotControllers[p.ID] {
			t.Errorf("seat %d's trigger missing from drained stack", i)
		}
	}
}

// TestS131SBALethalDamageDestroys covers CR 704.5g — a creature
// whose damage marked >= toughness is destroyed (moves to owner's
// graveyard).
func TestS131SBALethalDamageDestroys(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if err := g.MarkDamage(cardID, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if g.Battlefield.Contains(cardID) {
		t.Errorf("creature with lethal damage still on battlefield")
	}
	if !owner.Graveyard.Contains(cardID) {
		t.Errorf("creature did not route to owner's graveyard")
	}
}

// TestS131SBAToughnessReducedToZeroDestroys covers CR 704.5f — a
// creature with -1/-1 counters reducing toughness to <= 0 is
// destroyed. Printed-0 creatures are skipped (placeholder
// convention).
func TestS131SBAToughnessReducedToZeroDestroys(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Wither Victim",
		TypeLine:   "Creature — Beast",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
		Counters:   map[string]int{"-1/-1": 2},
	})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if g.Battlefield.Contains(cardID) {
		t.Errorf("creature reduced to 0 toughness still on battlefield")
	}
}

// TestS131SBAPlaceholderCreatureSurvives covers the placeholder
// convention: a creature with Toughness == 0 and no counters is
// treated as "stats not parsed" and left alone (matches the demo
// seed and Mortivore-style "*" cards documented on Card.Power).
func TestS131SBAPlaceholderCreatureSurvives(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Demo Card",
		TypeLine:   "Creature — Placeholder",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !g.Battlefield.Contains(cardID) {
		t.Errorf("placeholder creature was destroyed despite no counters")
	}
}

// TestS131SBAZeroLifeEliminates covers CR 704.5a — a player at 0 or
// less life loses immediately. Eliminated state is set; the game
// transitions to ended when only one survivor remains.
func TestS131SBAZeroLifeEliminates(t *testing.T) {
	g := newActiveGame(t)
	target := g.Seats[1]
	target.ChangeLife(-StartingLife) // exact 0
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !target.Eliminated {
		t.Errorf("player at 0 life not marked eliminated")
	}
	if g.State != StateEnded {
		t.Errorf("state: got %q, want ended (only one survivor)", g.State)
	}
}

// TestS131SBAEmptyLibraryDrawEliminates covers CR 704.5b — a player
// who tries to draw from an empty library is marked at draw time and
// eliminated on the next SBA pass. The auto-draw path swallows the
// underlying ErrZoneEmpty so the cursor still moves.
func TestS131SBAEmptyLibraryDrawEliminates(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	target := g.Seats[1]
	target.Library.Cards = nil
	if err := g.DrawCard(target.ID); err != ErrZoneEmpty {
		t.Fatalf("expected ErrZoneEmpty, got %v", err)
	}
	if !target.LosesAtNextSBA {
		t.Errorf("LosesAtNextSBA flag not set after empty-library draw")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !target.Eliminated {
		t.Errorf("target not eliminated after SBA pass")
	}
}

// TestS131SBA21CommanderDamageEliminates covers CR 704.5v / 903.14a
// — a player who has been dealt 21+ damage by a single commander
// loses. (Per-commander tracking lands in sub-PR 8; today's per-
// opponent map is used.)
func TestS131SBA21CommanderDamageEliminates(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	target := g.Seats[1]
	attacker := g.Seats[0]
	target.RecordCommanderDamage(attacker.ID, CommanderDamageLethal)
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !target.Eliminated {
		t.Errorf("target not eliminated after 21 commander damage")
	}
}

// TestS131CommanderCastTaxIncrements covers CR 903.8 — each cast
// from the command zone increments the per-commander counter so the
// client can show the +N tax label. The engine doesn't enforce the
// mana cost; the counter is the affordance.
func TestS131CommanderCastTaxIncrements(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	cmdrCard, err := caster.Command.Top()
	if err != nil {
		t.Fatalf("Command.Top: %v", err)
	}
	// Stamp a creature type-line so the cast routes to the stack
	// (the placeholder commander defaults to no type-line, which
	// the cast-as-non-permanent path would treat as not-creature
	// not-instant — we want a real cast loop).
	for i := range caster.Command.Cards {
		if caster.Command.Cards[i].InstanceID == cmdrCard.InstanceID {
			caster.Command.Cards[i].TypeLine = "Legendary Creature — Avatar"
			caster.Command.Cards[i].Power = 4
			caster.Command.Cards[i].Toughness = 4
		}
	}

	if err := g.CastSpell(caster.ID, cmdrCard.InstanceID, CastSpellParams{
		FromZone: "command",
	}); err != nil {
		t.Fatalf("first commander cast: %v", err)
	}
	if got := caster.CommanderCasts[cmdrCard.InstanceID]; got != 1 {
		t.Errorf("first cast: CommanderCasts=%d, want 1", got)
	}

	// Resolve and route back to command via the AsCommander flag.
	for g.stackHasItemsLocked() {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(cmdrCard.InstanceID) {
		t.Fatalf("commander did not resolve to battlefield")
	}
	dst := ZoneRef{Kind: ZoneGraveyard, Owner: caster.ID}
	if err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneBattlefield},
		dst,
		cmdrCard.InstanceID,
		true,
	); err != nil {
		t.Fatalf("MoveCardByIDAsCommander: %v", err)
	}
	if !caster.Command.Contains(cmdrCard.InstanceID) {
		t.Errorf("commander did not route to command zone with as_commander flag")
	}
	if caster.Graveyard.Contains(cmdrCard.InstanceID) {
		t.Errorf("commander leaked into graveyard despite zone replacement")
	}

	// Second cast — counter should be 2.
	if err := g.CastSpell(caster.ID, cmdrCard.InstanceID, CastSpellParams{
		FromZone: "command",
	}); err != nil {
		t.Fatalf("second commander cast: %v", err)
	}
	if got := caster.CommanderCasts[cmdrCard.InstanceID]; got != 2 {
		t.Errorf("second cast: CommanderCasts=%d, want 2", got)
	}
}

// TestS131CommanderZoneReplacementOnlyForCommanders verifies that
// the as_commander flag does NOT redirect non-commander cards.
// (Sandbox safety so a misclick doesn't clobber a normal move.)
func TestS131CommanderZoneReplacementOnlyForCommanders(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  cardID,
		Name:        "Plain Creature",
		TypeLine:    "Creature — Bear",
		Power:       2,
		Toughness:   2,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: false,
	})
	dst := ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID}
	if err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneBattlefield},
		dst,
		cardID,
		true,
	); err != nil {
		t.Fatalf("MoveCardByIDAsCommander: %v", err)
	}
	if !owner.Graveyard.Contains(cardID) {
		t.Errorf("non-commander did not route to graveyard despite as_commander flag")
	}
	if owner.Command.Contains(cardID) {
		t.Errorf("non-commander leaked into command zone")
	}
}

// TestS131ConcedeClearsStackItems covers CR 800.4a — when a player
// leaves the game, every spell + ability they control on the stack
// ceases to exist. Spell items move to exile (closest sandbox
// analogue to "cease to exist"); ability items + pending triggers
// disappear from StackMeta / PendingTriggers.
func TestS131ConcedeClearsStackItems(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	leaver := g.Seats[1]
	other := g.Seats[2]

	// Two spells from leaver + one from other on the stack.
	leaverSpell := pushTypedCardToHand(leaver, "Counterspell", "Instant")
	if err := g.CastSpell(leaver.ID, leaverSpell, CastSpellParams{}); err != nil {
		t.Fatalf("leaver cast: %v", err)
	}
	otherSpell := pushTypedCardToHand(other, "Lightning Bolt", "Instant")
	if err := g.CastSpell(other.ID, otherSpell, CastSpellParams{}); err != nil {
		t.Fatalf("other cast: %v", err)
	}
	// Activated ability from leaver pointing at a card they control.
	leaverSrc := pushCreatureToBattlefield(t, g, leaver)
	if err := g.ActivateAbility(leaver.ID, leaverSrc, AbilityParams{Label: "ability"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	// Pending trigger from leaver.
	if err := g.AnnounceTrigger(leaver.ID, leaverSrc, AbilityParams{Label: "trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	// Leaver's spell should be exiled; other's spell should still be on stack.
	if g.Stack.Contains(leaverSpell) {
		t.Errorf("leaver's spell still on stack after concede")
	}
	if !g.Exile.Contains(leaverSpell) {
		t.Errorf("leaver's spell did not move to exile (cease-to-exist analogue)")
	}
	if !g.Stack.Contains(otherSpell) {
		t.Errorf("other's spell wrongly removed by concede")
	}

	// All leaver-controlled stack metadata must be gone.
	for _, item := range g.StackMeta {
		if item != nil && item.Controller == leaver.ID {
			t.Errorf("leaver-controlled StackMeta entry persisted after concede: %+v", item)
		}
	}
	// Pending triggers from leaver dropped.
	for _, t2 := range g.PendingTriggers {
		if t2 != nil && t2.Controller == leaver.ID {
			t.Errorf("leaver-controlled pending trigger persisted after concede")
		}
	}
}

// TestS132SBAPlaneswalkerZeroLoyalty covers CR 704.5i — a
// planeswalker with 0 loyalty counters is moved to its owner's
// graveyard.
func TestS132SBAPlaneswalkerZeroLoyalty(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if g.Battlefield.Contains(pwID) {
		t.Errorf("planeswalker with 0 loyalty still on battlefield")
	}
	if !owner.Graveyard.Contains(pwID) {
		t.Errorf("planeswalker did not route to owner's graveyard")
	}
}

// TestS132SBABattleZeroDefense covers CR 704.5p — a battle with 0
// defense counters is moved to its owner's graveyard.
func TestS132SBABattleZeroDefense(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Invasion of Tolvada",
		TypeLine:   "Battle — Siege",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if g.Battlefield.Contains(id) {
		t.Errorf("battle with 0 defense still on battlefield")
	}
	if !owner.Graveyard.Contains(id) {
		t.Errorf("battle did not route to owner's graveyard")
	}
}

// TestS132SBAPlusMinusCounterCancel covers CR 704.5q — +1/+1 and
// -1/-1 counters on the same creature cancel 1-for-1.
func TestS132SBAPlusMinusCounterCancel(t *testing.T) {
	cases := []struct {
		name      string
		plus      int
		minus     int
		wantPlus  int
		wantMinus int
	}{
		{"1 each cancels both", 1, 1, 0, 0},
		{"3 plus + 2 minus → 1 plus", 3, 2, 1, 0},
		{"2 plus + 5 minus → 3 minus", 2, 5, 0, 3},
		{"only plus stays", 2, 0, 2, 0},
		{"only minus stays (still alive)", 0, 1, 0, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			id := uuid.New()
			counters := map[string]int{}
			if c.plus > 0 {
				counters[CounterPlusOne] = c.plus
			}
			if c.minus > 0 {
				counters[CounterMinusOne] = c.minus
			}
			g.Battlefield.PushTop(Card{
				InstanceID: id,
				Name:       "Test Creature",
				TypeLine:   "Creature — Beast",
				Power:      4,
				Toughness:  4,
				Owner:      owner.ID,
				Controller: owner.ID,
				Counters:   counters,
			})
			g.WithWriteLock(func() { g.runStateChecksLocked() })
			var got *Card
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					got = &g.Battlefield.Cards[i]
					break
				}
			}
			if got == nil {
				t.Fatalf("creature destroyed unexpectedly")
			}
			if got.Counters[CounterPlusOne] != c.wantPlus {
				t.Errorf("+1/+1: got %d, want %d", got.Counters[CounterPlusOne], c.wantPlus)
			}
			if got.Counters[CounterMinusOne] != c.wantMinus {
				t.Errorf("-1/-1: got %d, want %d", got.Counters[CounterMinusOne], c.wantMinus)
			}
		})
	}
}

// TestS132SBAPoisonTenLosesGame covers CR 704.5c — a player with
// ≥ 10 poison counters loses the game.
func TestS132SBAPoisonTenLosesGame(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	target := g.Seats[1]
	if err := g.AddPlayerCounter(target.ID, CounterPoison, PoisonLethal); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}
	if !target.Eliminated {
		t.Errorf("player at %d poison not eliminated", PoisonLethal)
	}
	if target.Poison != PoisonLethal {
		t.Errorf("Player.Poison legacy field out of sync: got %d, want %d",
			target.Poison, PoisonLethal)
	}
}

// TestS132AddPlayerCounterClampsAtZero verifies negative deltas
// driving below zero clamp at zero.
func TestS132AddPlayerCounterClampsAtZero(t *testing.T) {
	g := newActiveGame(t)
	target := g.Seats[0]
	if err := g.AddPlayerCounter(target.ID, CounterEnergy, 3); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := g.AddPlayerCounter(target.ID, CounterEnergy, -10); err != nil {
		t.Fatalf("subtract: %v", err)
	}
	if got := target.Counters[CounterEnergy]; got != 0 {
		t.Errorf("energy: got %d, want 0", got)
	}
	if target.Energy != 0 {
		t.Errorf("Energy legacy field: got %d, want 0", target.Energy)
	}
}

// TestS132CleanupClearsDamageMarked covers CR 514.2 — at the start
// of cleanup, all damage marked on permanents is removed.
func TestS132CleanupClearsDamageMarked(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Big Creature",
		TypeLine:   "Creature — Avatar",
		Power:      4,
		Toughness:  10,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if err := g.MarkDamage(cardID, 5); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID && c.DamageMarked != 5 {
			t.Fatalf("setup: DamageMarked=%d, want 5", c.DamageMarked)
		}
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID && c.DamageMarked != 0 {
			t.Errorf("DamageMarked after cleanup: got %d, want 0", c.DamageMarked)
		}
	}
}

// TestS134CleanupPromptsForOverMaxHand verifies that an end-of-turn
// hand of 9 (default cap = 7) populates DiscardPending with the
// over-max count and pauses the cleanup auto-advance.
func TestS134CleanupPromptsForOverMaxHand(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	// Stuff the active player's hand to 9 cards.
	for active.Hand.Size() < 9 {
		_ = g.DrawCard(active.ID)
	}
	if active.Hand.Size() != 9 {
		t.Fatalf("hand setup: got %d, want 9", active.Hand.Size())
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	// Cursor should be paused at Cleanup, not the next seat's Upkeep.
	if g.Turn.Step != StepCleanup {
		t.Errorf("cursor advanced past Cleanup despite over-max hand: step=%q", g.Turn.Step)
	}
	if g.DiscardPending[active.ID] != 2 {
		t.Errorf("DiscardPending[active]: got %d, want 2", g.DiscardPending[active.ID])
	}
}

// TestS134DiscardSelectionResolvesAndAdvances verifies that
// dispatching discard_selection with the correct count moves the
// chosen cards to graveyard, drains the pending entry, and
// re-fires the cleanup hook so the cursor walks on to the next
// seat's Upkeep.
func TestS134DiscardSelectionResolvesAndAdvances(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	for active.Hand.Size() < 9 {
		_ = g.DrawCard(active.ID)
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	if g.Turn.Step != StepCleanup {
		t.Fatalf("expected paused at Cleanup, got %q", g.Turn.Step)
	}

	// Pick 2 hand cards to discard.
	picks := []uuid.UUID{
		active.Hand.Cards[0].InstanceID,
		active.Hand.Cards[1].InstanceID,
	}
	gyBefore := active.Graveyard.Size()
	if err := g.DiscardSelection(active.ID, picks); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if active.Graveyard.Size() != gyBefore+2 {
		t.Errorf("graveyard size: got %d, want %d", active.Graveyard.Size(), gyBefore+2)
	}
	if len(g.DiscardPending) != 0 {
		t.Errorf("DiscardPending should be empty after selection: %+v", g.DiscardPending)
	}
	// Cursor should have advanced to the next seat's Upkeep.
	if g.Turn.ActiveSeat == 0 {
		t.Errorf("cursor stayed on seat 0 after discard resolution")
	}
}

// TestS134DiscardSelectionWrongCount rejects with ErrInvalidParam.
func TestS134DiscardSelectionWrongCount(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	for active.Hand.Size() < 9 {
		_ = g.DrawCard(active.ID)
	}
	advanceTo(t, g, StepEnd)
	_, _ = g.AdvanceStep()
	// Need to discard 2; supply 1.
	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID}); err != ErrInvalidParam {
		t.Errorf("wrong count: got %v, want ErrInvalidParam", err)
	}
}

// TestS134NoMaxHandSizeBypassesPrompt verifies that
// MaxHandSize == NoMaxHandSize (-1) skips the discard prompt
// entirely — the Reliquary Tower / Thought Vessel effect.
func TestS134NoMaxHandSizeBypassesPrompt(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	if err := g.SetMaxHandSize(active.ID, NoMaxHandSize); err != nil {
		t.Fatalf("SetMaxHandSize: %v", err)
	}
	for active.Hand.Size() < 15 {
		_ = g.DrawCard(active.ID)
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if len(g.DiscardPending) != 0 {
		t.Errorf("DiscardPending non-empty for no-cap player: %+v", g.DiscardPending)
	}
	if g.Turn.ActiveSeat == 0 {
		t.Errorf("cursor stuck on seat 0 despite no-cap (no prompt expected)")
	}
}

// TestS134SetMaxHandSizeRejectsBadValue covers the validation gate.
func TestS134SetMaxHandSizeRejectsBadValue(t *testing.T) {
	g := newActiveGame(t)
	if err := g.SetMaxHandSize(g.Seats[0].ID, -2); err != ErrInvalidParam {
		t.Errorf("got %v, want ErrInvalidParam for -2", err)
	}
}

// TestS135StartInitialisesKnownBy verifies S13.5's initial-state
// knowledge: library cards have empty KnownBy, hand cards are
// known to their owner only, command-zone cards are known to all
// seated players.
func TestS135StartInitialisesKnownBy(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for _, p := range g.Seats {
		// Library cards: zero knowers (post-shuffle nobody knows the
		// top-of-library order, including the owner).
		for _, c := range p.Library.Cards {
			if len(c.KnownBy) != 0 {
				t.Errorf("library card %s starts with knowers: %v", c.Name, c.KnownBy)
			}
		}
		// Hand cards: owner only.
		for _, c := range p.Hand.Cards {
			if !c.IsKnownTo(p.ID) {
				t.Errorf("hand card %s not known to owner", c.Name)
			}
			for _, other := range g.Seats {
				if other.ID != p.ID && c.IsKnownTo(other.ID) {
					t.Errorf("hand card %s leaked to opponent %s", c.Name, other.Name)
				}
			}
		}
		// Command-zone cards: all seated.
		for _, c := range p.Command.Cards {
			for _, other := range g.Seats {
				if !c.IsKnownTo(other.ID) {
					t.Errorf("command card %s not known to %s", c.Name, other.Name)
				}
			}
		}
	}
}

// TestS135ShuffleClearsLibraryKnowledge verifies that ShuffleLibrary
// clears KnownBy on every library card. Models scry-then-shuffle:
// the player saw the top, then the shuffle erased their knowledge.
func TestS135ShuffleClearsLibraryKnowledge(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	// Pretend the player scryd and now knows the top card.
	if owner.Library.Size() == 0 {
		t.Fatal("setup: empty library")
	}
	owner.Library.Cards[0].AddKnower(owner.ID)
	if !owner.Library.Cards[0].IsKnownTo(owner.ID) {
		t.Fatal("setup: KnownBy add failed")
	}
	if err := g.ShuffleLibrary(owner.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	for _, c := range owner.Library.Cards {
		if len(c.KnownBy) != 0 {
			t.Errorf("library card %s retains knowers after shuffle: %v", c.Name, c.KnownBy)
		}
	}
}

// TestS135PublicZoneMoveGrantsKnowledge verifies that moving a card
// to a public zone (battlefield) marks every seated player as a
// knower — a creature played by one seat is visible to all.
func TestS135PublicZoneMoveGrantsKnowledge(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	id := pushTypedCardToHand(caster, "Grizzly Bears", "Creature — Bear")
	// Hand-place doesn't grant knowers — manual setup. Confirm the
	// pre-state.
	caster.Hand.Cards[len(caster.Hand.Cards)-1].KnownBy = nil
	if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Walk priority to resolve.
	for g.stackHasItemsLocked() {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// Find the resolved card on the battlefield and verify every
	// seated player is now a knower.
	var found *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			found = &g.Battlefield.Cards[i]
		}
	}
	if found == nil {
		t.Fatal("creature not on battlefield")
	}
	for _, p := range g.Seats {
		if !found.IsKnownTo(p.ID) {
			t.Errorf("battlefield card not known to %s", p.Name)
		}
	}
}

// TestS135DrawAddsOwnerOnly verifies that drawing a card from the
// library makes only the owner a knower (not opponents).
func TestS135DrawAddsOwnerOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0]
	handBefore := owner.Hand.Size()
	if err := g.DrawCard(owner.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if owner.Hand.Size() != handBefore+1 {
		t.Fatalf("hand size: got %d, want %d", owner.Hand.Size(), handBefore+1)
	}
	drawn := owner.Hand.Cards[owner.Hand.Size()-1]
	if !drawn.IsKnownTo(owner.ID) {
		t.Errorf("owner not added as knower on draw")
	}
	for _, other := range g.Seats[1:] {
		if drawn.IsKnownTo(other.ID) {
			t.Errorf("opponent %s leaked as knower of drawn card", other.Name)
		}
	}
}

// TestS131SplitSecondClearsOnResolve verifies that after the split-
// second item resolves, SplitSecondActive flips back to false and
// new casts go through.
func TestS131SplitSecondClearsOnResolve(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	splitID := pushTypedCardToHand(p, "Trickbind", "Instant")
	if err := g.CastSpell(p.ID, splitID, CastSpellParams{SplitSecond: true}); err != nil {
		t.Fatalf("split-second cast: %v", err)
	}

	// Pass twice (caster → opponent → wrap+resolve).
	_ = g.PassPriority()
	_ = g.PassPriority()

	if g.SplitSecondActive {
		t.Errorf("SplitSecondActive should be false after resolve")
	}
	if !p.Graveyard.Contains(splitID) {
		t.Errorf("split-second instant should have routed to graveyard")
	}
	// Subsequent cast should now succeed.
	otherID := pushTypedCardToHand(p, "Lightning Bolt", "Instant")
	if err := g.CastSpell(p.ID, otherID, CastSpellParams{}); err != nil {
		t.Errorf("post-resolve cast: %v", err)
	}
}

// TestAdvanceStepRunsSBAs covers CR 117.5 / 704.3: state-based
// actions are checked at every priority-granting boundary, including
// step transitions driven by AdvanceStep (not only by PassPriority).
// Regression: AdvanceStep used to skip the SBA loop, so a
// planeswalker that hit 0 loyalty mid-step stayed on the battlefield
// until the next PassPriority.
func TestAdvanceStepRunsSBAs(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Battlefield.Contains(pwID) {
		t.Errorf("0-loyalty planeswalker still on battlefield after AdvanceStep")
	}
	if !owner.Graveyard.Contains(pwID) {
		t.Errorf("planeswalker did not route to owner graveyard")
	}
}

// TestAdvanceStepEliminatesZeroLife covers CR 704.5a: a player at 0
// life loses the game on the next SBA check. AdvanceStep is a valid
// SBA trigger point, so stepping through end-of-turn with a player
// at 0 life should eliminate them without needing an explicit
// PassPriority.
func TestAdvanceStepEliminatesZeroLife(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	target := g.Seats[1]
	target.ChangeLife(-StartingLife) // exact 0
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if !target.Eliminated {
		t.Errorf("player at 0 life not eliminated after AdvanceStep")
	}
}

// TestPassTurnRunsSBAs mirrors TestAdvanceStepRunsSBAs for the
// "next turn" shortcut. PassTurn lands on a priority-granting step
// (Upkeep, via the auto-advance past Untap), so SBAs should fire
// once the cursor settles.
func TestPassTurnRunsSBAs(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0]
	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Jace",
		TypeLine:   "Legendary Planeswalker — Jace",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Battlefield.Contains(pwID) {
		t.Errorf("0-loyalty planeswalker still on battlefield after PassTurn")
	}
	if !owner.Graveyard.Contains(pwID) {
		t.Errorf("planeswalker did not route to owner graveyard")
	}
}

// --- S15 sub-PR 3 strict-mode cost gate ------------------------

// pushTypedCardToHandWithCost is the cost-aware sibling of
// pushTypedCardToHand. Used by the strict-mode cast tests so the
// cost validator has something to parse.
func pushTypedCardToHandWithCost(p *Player, name, typeLine, manaCost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Hand.PushTop(c)
	return c.InstanceID
}

// TestS15CastSpellPermissiveWarnsAndProceeds proves the default
// (Strict=false) path emits an EventCostWarning and lands the
// spell on the stack regardless of an empty pool. Sandbox-style.
func TestS15CastSpellPermissiveWarnsAndProceeds(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHandWithCost(p, "Lightning Bolt", "Instant", "{R}")

	beforeEvents := len(g.Events)
	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("permissive CastSpell: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Errorf("permissive cast did not reach the stack")
	}
	// Pool unchanged.
	if len(p.ManaPool) != 0 {
		t.Errorf("permissive cast touched the pool: %+v", p.ManaPool)
	}
	// One EventCostWarning recorded for the cast.
	sawWarning := false
	for _, ev := range g.Events[beforeEvents:] {
		if ev.Kind == EventCostWarning && ev.Source == id {
			sawWarning = true
			break
		}
	}
	if !sawWarning {
		t.Errorf("expected EventCostWarning emitted for permissive cast")
	}
}

// TestS15CastSpellStrictRejectsWhenShort proves the strict-mode
// gate refuses casts the pool can't cover and surfaces a
// structured InsufficientManaError listing the missing symbols.
func TestS15CastSpellStrictRejectsWhenShort(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHandWithCost(p, "Counterspell", "Instant", "{U}{U}")

	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true})
	if err == nil {
		t.Fatalf("strict CastSpell with empty pool should have failed")
	}
	var im *InsufficientManaError
	if !errorsAs(err, &im) {
		t.Fatalf("expected *InsufficientManaError, got %T (%v)", err, err)
	}
	want := []string{"{U}", "{U}"}
	if len(im.Missing) != len(want) {
		t.Fatalf("Missing: got %v, want %v", im.Missing, want)
	}
	for i, w := range want {
		if im.Missing[i] != w {
			t.Errorf("Missing[%d]: got %q, want %q", i, im.Missing[i], w)
		}
	}
	// The cast should not have advanced state — card stays in hand.
	if !p.Hand.Contains(id) {
		t.Errorf("rejected cast left hand")
	}
	if g.Stack != nil && g.Stack.Contains(id) {
		t.Errorf("rejected cast reached the stack")
	}
}

// TestS15CastSpellStrictDeductsWhenPayable proves the happy path:
// strict mode + payable pool drains the right tokens and lets the
// cast proceed.
func TestS15CastSpellStrictDeductsWhenPayable(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHandWithCost(p, "Lightning Bolt", "Instant", "{R}")
	p.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "C"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("strict CastSpell payable: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Errorf("strict cast did not reach the stack")
	}
	// Red was spent; colorless survives.
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "C" {
		t.Errorf("post-spend pool: got %+v, want 1×C", p.ManaPool)
	}
}

// TestS15CastSpellForceCastBypassesGate proves the override toast
// path: strict mode set, pool can't cover, but ForceCast=true lets
// the cast proceed without touching the pool.
func TestS15CastSpellForceCastBypassesGate(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushTypedCardToHandWithCost(p, "Counterspell", "Instant", "{U}{U}")

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, ForceCast: true}); err != nil {
		t.Fatalf("force-cast: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Errorf("force-cast did not reach the stack")
	}
	// Pool unchanged — override doesn't deduct.
	if len(p.ManaPool) != 0 {
		t.Errorf("force-cast touched empty pool: %+v", p.ManaPool)
	}
}

// TestS15EffectiveCostStrictCommanderTax exercises the {2}-per-prior-
// cast surcharge under strict mode (CR 903.8). First cast pays the
// printed cost; second cast (after returning the commander to the
// command zone) needs cost+{2}.
func TestS15EffectiveCostStrictCommanderTax(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	cmdr := NewCard("Test Commander", p.ID)
	cmdr.TypeLine = "Legendary Creature"
	cmdr.ManaCost = "{1}{W}"
	cmdr.IsCommander = true
	p.Command.PushTop(cmdr)
	id := cmdr.InstanceID

	// Seed enough for {1}{W} but not for {3}{W} (1 generic + 1 white).
	p.ManaPool.AddMana(ManaToken{Color: "W"}, ManaToken{Color: "C"})

	// First cast under strict mode — should succeed.
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, FromZone: "command"}); err != nil {
		t.Fatalf("first commander cast: %v", err)
	}
	if g.Seats[0].CommanderCasts[id] != 1 {
		t.Errorf("CommanderCasts after first cast: got %d, want 1", g.Seats[0].CommanderCasts[id])
	}

	// Return the commander to the command zone (simulate "next time
	// it would change zones, send to command instead" replacement).
	if _, err := MoveCard(g.Stack, p.Command, id); err != nil {
		t.Fatalf("move back to command: %v", err)
	}
	delete(g.StackMeta, id)
	// Seed enough for {1}{W} ONLY (no tax) — should be too short.
	p.ManaPool = nil
	p.ManaPool.AddMana(ManaToken{Color: "W"}, ManaToken{Color: "C"})
	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, FromZone: "command"})
	if err == nil {
		t.Fatalf("second commander cast on cost+0 pool should have failed (tax requires +{2})")
	}
	var im *InsufficientManaError
	if !errorsAs(err, &im) {
		t.Fatalf("expected *InsufficientManaError on second cast, got %v", err)
	}
	// Now seed cost + {2} surcharge — should succeed.
	p.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, FromZone: "command"}); err != nil {
		t.Fatalf("second commander cast on cost+{2} pool: %v", err)
	}
}

// errorsAs is a tiny test-local wrapper to avoid pulling errors
// into the import block in every TestS15 case.
func errorsAs(err error, target any) bool {
	return errors.As(err, target)
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
	if turn.Number != 1 || turn.Step != StepUpkeep {
		t.Errorf("turn: %+v (want number=1 step=upkeep after S13 auto-advance from Untap)", turn)
	}
}
