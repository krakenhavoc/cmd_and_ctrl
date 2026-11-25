package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// blocks_by_piles_test.go — #6200: Camouflage's piles (CR 509.1, 700.3a,
// 700.3d, 101.4) and Raging River's labelled piles (CR 509.1b).

// pileAttack has seat 0 declare `attackers` at seat 1 and stops in the
// declare attackers step, before the declare blockers step begins.
func pileAttack(t *testing.T, g *Game, attackers ...uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
}

// answerPile answers the open choose_cards prompt with `picks`.
func answerPile(t *testing.T, g *Game, picks ...uuid.UUID) {
	t.Helper()
	c := openPromptOfKind(g, PendingChoiceChooseCards)
	if c == nil {
		t.Fatal("no pile prompt is open")
	}
	if err := g.ResolveChooseCards(c.ID, c.Chooser, picks); err != nil {
		t.Fatalf("answering the pile: %v", err)
	}
}

// pileBlockOf is the attacker `blocker` blocks, or uuid.Nil.
func pileBlockOf(g *Game, blocker uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, blocker); c != nil {
			out = c.BlockingTarget
		}
	})
	return out
}

// TestCamouflagePilesBlockTheAttackerTheyAreAssignedToAtRandom — two
// attackers, so two piles; each pile lands on a different attacker at
// random and every creature in it blocks that attacker. The defender
// can't declare blockers themselves, and nobody gets priority until the
// piles are in (CR 509.1, 509.2).
func TestCamouflagePilesBlockTheAttackerTheyAreAssignedToAtRandom(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "Attacker One", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "Attacker Two", 3, 3)
	w1 := pushCombatant(t, g, g.Seats[1], "Wall One", 0, 4)
	w2 := pushCombatant(t, g, g.Seats[1], "Wall Two", 0, 4)
	w3 := pushCombatant(t, g, g.Seats[1], "Wall Three", 0, 4)
	pileAttack(t, g, a1, a2)
	g.WithWriteLock(func() {
		if !g.BlocksByPilesThisTurnForEffect(uuid.Nil, g.Seats[0].ID, "Camouflage") {
			t.Fatal("no record was registered")
		}
	})
	advanceIntoStep(t, g, StepDeclareBlockers)

	def := g.Seats[1].ID
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationPending {
		t.Fatalf("defender status = %q, want pending while the piles are asked", got)
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("priority holder = %d, want parked while the piles are asked", g.Turn.PriorityHolder)
	}
	if err := g.DeclareBlockers([]BlockDeclaration{{Blocker: w1, Attacker: a1}}); !errors.Is(err, ErrBlocksDecidedByPiles) {
		t.Fatalf("a declared block = %v, want ErrBlocksDecidedByPiles", err)
	}
	first := openPromptOfKind(g, PendingChoiceChooseCards)
	if first == nil || first.Chooser != def || len(first.ChooseCards) != 3 {
		t.Fatalf("first pile prompt = %+v, want the defender's three creatures", first)
	}
	answerPile(t, g, w1, w2)
	second := openPromptOfKind(g, PendingChoiceChooseCards)
	if second == nil || len(second.ChooseCards) != 1 || second.ChooseCards[0] != w3 {
		t.Fatalf("second pile prompt = %+v, want only Wall Three (CR 700.3a: each creature in one pile)", second)
	}
	answerPile(t, g, w3)

	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("defender status = %q, want declared once the piles are in", got)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("priority holder = %d, want the active player (CR 509.2)", g.Turn.PriorityHolder)
	}
	pileOne := pileBlockOf(g, w1)
	if pileOne != a1 && pileOne != a2 {
		t.Fatalf("Wall One blocks %v, want one of the attackers", pileOne)
	}
	if pileBlockOf(g, w2) != pileOne {
		t.Fatalf("Wall Two blocks %v, want the same attacker as Wall One (one pile)", pileBlockOf(g, w2))
	}
	other := a1
	if pileOne == a1 {
		other = a2
	}
	if pileBlockOf(g, w3) != other {
		t.Fatalf("Wall Three blocks %v, want the other attacker (a different pile)", pileBlockOf(g, w3))
	}
}

// TestCamouflageAssignmentIsRandom — over many seeded games the first pile
// lands on each attacker at least once: the assignment is drawn, not the
// battlefield order.
func TestCamouflageAssignmentIsRandom(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 24 && len(seen) < 2; i++ {
		g := newActiveGame(t)
		a1 := pushCombatant(t, g, g.Seats[0], "Attacker One", 2, 2)
		a2 := pushCombatant(t, g, g.Seats[0], "Attacker Two", 3, 3)
		w1 := pushCombatant(t, g, g.Seats[1], "Wall One", 0, 4)
		pileAttack(t, g, a1, a2)
		// Test games share one fixed RNG key; a different source names a
		// different stream.
		g.WithWriteLock(func() { g.BlocksByPilesThisTurnForEffect(uuid.New(), g.Seats[0].ID, "Camouflage") })
		advanceIntoStep(t, g, StepDeclareBlockers)
		answerPile(t, g, w1) // the second pile is empty: nothing is left to put in it
		switch pileBlockOf(g, w1) {
		case a1:
			seen["one"] = true
		case a2:
			seen["two"] = true
		default:
			t.Fatal("Wall One blocks nothing")
		}
	}
	if len(seen) != 2 {
		t.Fatalf("the first pile always landed on the same attacker: %v", seen)
	}
}

// TestCamouflagePileCreaturesThatCantBlockDont — a creature in a pile that
// can't block its pile's attacker (a flyer, here, every pile) doesn't, and
// an empty set of piles blocks nothing (CR 700.3d).
func TestCamouflagePileCreaturesThatCantBlockDont(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "Bird", 1, 1, "flying")
	w1 := pushCombatant(t, g, g.Seats[1], "Wall One", 0, 4)
	pileAttack(t, g, a1)
	g.WithWriteLock(func() { g.BlocksByPilesThisTurnForEffect(uuid.Nil, g.Seats[0].ID, "Camouflage") })
	advanceIntoStep(t, g, StepDeclareBlockers)
	answerPile(t, g, w1)
	if pileBlockOf(g, w1) != uuid.Nil {
		t.Fatal("Wall One blocked a flyer from its pile")
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationDeclared {
		t.Fatalf("defender status = %q, want declared", got)
	}
}

// TestCamouflageExtraBlockerMayGoInTwoPiles — "Creatures those players
// control that can block additional creatures may likewise be put into
// additional piles": a creature that can block one more is offered for
// the second pile too, and blocks both attackers.
func TestCamouflageExtraBlockerMayGoInTwoPiles(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "Attacker One", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "Attacker Two", 3, 3)
	guard := pushCombatant(t, g, g.Seats[1], "Palace Guard", 1, 4)
	pileAttack(t, g, a1, a2)
	g.WithWriteLock(func() {
		findCard(g, guard).effective.BlocksAnyNumber = true
		g.BlocksByPilesThisTurnForEffect(uuid.Nil, g.Seats[0].ID, "Camouflage")
	})
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.WithWriteLock(func() { findCard(g, guard).effective.BlocksAnyNumber = true })
	answerPile(t, g, guard)
	second := openPromptOfKind(g, PendingChoiceChooseCards)
	if second == nil || len(second.ChooseCards) != 1 {
		t.Fatalf("second pile prompt = %+v, want the guard offered again", second)
	}
	answerPile(t, g, guard)
	var blocked []uuid.UUID
	g.WithWriteLock(func() { blocked = findCard(g, guard).BlockedAttackers() })
	if len(blocked) != 2 {
		t.Fatalf("the guard blocks %v, want both attackers", blocked)
	}
}

// TestCamouflageIsOverAtEndOfTurn — the record is "this turn": swept at
// cleanup (CR 514.2).
func TestCamouflageIsOverAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() { g.BlocksByPilesThisTurnForEffect(uuid.Nil, g.Seats[0].ID, "Camouflage") })
	advanceIntoStep(t, g, StepEnd)
	advanceIntoStep(t, g, StepUpkeep)
	var on bool
	g.WithWriteLock(func() { on = g.BlocksByPilesForEffect() })
	if on {
		t.Fatal("Camouflage's record outlived its turn")
	}
}

// TestRagingRiverLabelsLimitBlockersToTheChosenPile — the defender puts
// Wall One left and Wall Two right; the attacker labelled "left" can be
// blocked by Wall One and by a flyer, not by Wall Two (CR 509.1b).
func TestRagingRiverLabelsLimitBlockersToTheChosenPile(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "Attacker One", 2, 2)
	w1 := pushCombatant(t, g, g.Seats[1], "Wall One", 0, 4)
	w2 := pushCombatant(t, g, g.Seats[1], "Wall Two", 0, 4)
	bird := pushCombatant(t, g, g.Seats[1], "Bird", 1, 1, "flying")
	pileAttack(t, g, a1)
	g.WithWriteLock(func() { g.QueueLeftRightPilesForEffect(uuid.Nil, g.Seats[0].ID) })
	split := openPromptOfKind(g, PendingChoiceChooseCards)
	if split == nil || split.Chooser != g.Seats[1].ID || len(split.ChooseCards) != 2 {
		t.Fatalf("split prompt = %+v, want the defender's two creatures without flying", split)
	}
	answerPile(t, g, w1)
	label := openPromptOfKind(g, PendingChoiceOptionPick)
	if label == nil || label.Chooser != g.Seats[0].ID {
		t.Fatalf("label prompt = %+v, want the attacking player", label)
	}
	if err := g.ResolveOptionPick(label.ID, label.Chooser, 0); err != nil {
		t.Fatalf("choosing left: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	var refused, wallOK, birdOK bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		atk := findBattlefieldCard(g, a1)
		wallOK = g.CanBlockLocked(atk, findBattlefieldCard(g, w1))
		refused = !g.CanBlockLocked(atk, findBattlefieldCard(g, w2))
		birdOK = g.CanBlockLocked(atk, findBattlefieldCard(g, bird))
	})
	if !wallOK {
		t.Error("Wall One, in the left pile, can't block the attacker labelled left")
	}
	if !refused {
		t.Error("Wall Two, in the right pile, can block the attacker labelled left")
	}
	if !birdOK {
		t.Error("a creature with flying can't block the attacker labelled left")
	}
	if err := g.DeclareBlockers([]BlockDeclaration{{Blocker: w2, Attacker: a1}}); !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("Wall Two's block = %v, want refused", err)
	}
	if err := g.DeclareBlockers([]BlockDeclaration{{Blocker: w1, Attacker: a1}}); err != nil {
		t.Fatalf("Wall One's block: %v", err)
	}
}
