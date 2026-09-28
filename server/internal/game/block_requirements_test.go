package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// block_requirements_test.go pins CR 509.1c on the block declaration
// (#1597, ADR 0045 amendment of 2026-09-28): the verb refuses a
// declaration that makes a requirement unobeyable, the defending
// player's pass, finish_blocks and AdvanceStep refuse one that leaves
// an obeyable requirement unmet, and the maximum is counted over the
// whole declaration against menace, evasion, "can't block" and the
// whole-combat limits. The cards, the enumerator, the bot and the view
// are pinned in cards/effects, legal, aiseat and protocol.

// withBlockRequirement gives `id` one requirement of each kind listed,
// as a pinned data record (what Irresistible Prey writes).
func withBlockRequirement(t *testing.T, g *Game, id uuid.UUID, kinds ...BlockRequirementKind) {
	t.Helper()
	mods := make([]Mod, len(kinds))
	for i, k := range kinds {
		mods[i] = AddBlockRequirementMod(k)
	}
	registerScopedEffectForTest(t, g, id, mods, IndefiniteDuration())
}

// blockRequirementErr asserts err is a block_requirement refusal.
func blockRequirementErr(t *testing.T, err error) *BlockRefusedError {
	t.Helper()
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("err = %v, want ErrIllegalBlock", err)
	}
	var br *BlockRefusedError
	if !errors.As(err, &br) {
		t.Fatalf("refusal is %T, want *BlockRefusedError", err)
	}
	if br.Reason != BlockReasonRequirement {
		t.Fatalf("reason = %q, want %q (%v)", br.Reason, BlockReasonRequirement, err)
	}
	return br
}

// passToDefender hands priority from the active player to seat 1, the
// defender, so the next PassPriority is the defender's.
func passToDefender(t *testing.T, g *Game) {
	t.Helper()
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("priority holder %d, want the active seat", g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("active player's pass: %v", err)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Fatalf("priority holder %d after the active pass, want the defender", g.Turn.PriorityHolder)
	}
}

func block(t *testing.T, g *Game, pairs ...uuid.UUID) error {
	t.Helper()
	var decls []BlockDeclaration
	for i := 0; i+1 < len(pairs); i += 2 {
		decls = append(decls, BlockDeclaration{Blocker: pairs[i], Attacker: pairs[i+1]})
	}
	return g.DeclareBlockers(decls)
}

// TestLureMakesEveryAbleCreatureBlockIt — three creatures able to block
// a Lure'd attacker must all block it. Sending one elsewhere is refused
// at the verb; a declaration that leaves one out is accepted as a step
// on the way, and refused at the checkpoint — the pass, finish_blocks
// and AdvanceStep alike — naming the creature it left out.
func TestLureMakesEveryAbleCreatureBlockIt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lured := pushCombatant(t, g, me, "Lured Bear", 2, 2)
	other := pushCombatant(t, g, me, "Other Bear", 2, 2)
	withBlockRequirement(t, g, lured, BlockRequirementLure)
	b1 := pushCombatant(t, g, opp, "Wall One", 0, 4)
	b2 := pushCombatant(t, g, opp, "Wall Two", 0, 4)
	b3 := pushCombatant(t, g, opp, "Wall Three", 0, 4)
	declareAttacks(t, g, lured, other)

	// Blocking the other attacker gives one requirement up.
	br := blockRequirementErr(t, block(t, g, b1, other))
	if br.Blocker != b1 || br.Attacker != lured {
		t.Errorf("refusal names %s -> %s, want Wall One -> the Lure'd attacker", br.Blocker, br.Attacker)
	}
	if got := br.Sentence(opp.ID); got != "Wall One must block Lured Bear if able." {
		t.Errorf("sentence = %q", got)
	}
	// Two of three is a legal step.
	if err := block(t, g, b1, lured, b2, lured); err != nil {
		t.Fatalf("two of the three on the Lure'd attacker: %v", err)
	}
	passToDefender(t, g)
	br = blockRequirementErr(t, g.PassPriority())
	if br.Blocker != b3 {
		t.Errorf("the pass's refusal names %s, want the creature left out", br.Blocker)
	}
	blockRequirementErr(t, g.FinishBlocks(opp.ID))
	blockRequirementErr(t, func() error { _, err := g.AdvanceStep(); return err }())
	if g.Turn.Step != StepDeclareBlockers || g.BlockDeclarationStatusOf(opp.ID) != BlockDeclarationPending {
		t.Fatalf("a refused checkpoint moved the game on: step %s", g.Turn.Step)
	}
	if err := block(t, g, b3, lured); err != nil {
		t.Fatalf("the third blocker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every requirement obeyed: %v", err)
	}
	if g.BlockDeclarationStatusOf(opp.ID) != BlockDeclarationDeclared {
		t.Error("the accepted pass did not complete the declaration")
	}
}

// TestLureOnAMenaceAttackerRespectsMenace — CR 509.1c counts only what
// a restriction allows: one creature cannot block a menace attacker
// alone, so a lone potential blocker owes nothing, and with two the
// pair is owed — offered as one declaration, because the singles are
// refused.
func TestLureOnAMenaceAttackerRespectsMenace(t *testing.T) {
	t.Run("one blocker owes nothing", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushCombatant(t, g, me, "Menace Bear", 2, 2, "menace")
		withBlockRequirement(t, g, atk, BlockRequirementLure)
		pushCombatant(t, g, opp, "Lone Wall", 0, 4)
		declareAttacks(t, g, atk)
		// Nothing is legal, so the declaration completed at step start.
		if got := g.BlockDeclarationStatusOf(opp.ID); got != BlockDeclarationDeclared {
			t.Fatalf("status = %q: a defender with no legal block owes no Lure block", got)
		}
	})
	t.Run("two blockers owe the pair", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushCombatant(t, g, me, "Menace Bear", 2, 2, "menace")
		withBlockRequirement(t, g, atk, BlockRequirementLure)
		w1 := pushCombatant(t, g, opp, "Wall One", 0, 4)
		w2 := pushCombatant(t, g, opp, "Wall Two", 0, 4)
		declareAttacks(t, g, atk)
		passToDefender(t, g)
		blockRequirementErr(t, g.PassPriority())
		var owed []BlockDeclaration
		g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
		if len(owed) != 2 {
			t.Fatalf("witness = %+v, want both walls on the menace attacker", owed)
		}
		var br *BlockRefusedError
		if !errors.As(block(t, g, w1, atk), &br) || br.Reason != BlockReasonTooFewBlockers {
			t.Fatalf("a lone block on menace must still be too_few_blockers, got %v", br)
		}
		if err := block(t, g, w1, atk, w2, atk); err != nil {
			t.Fatalf("the pair: %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the pair declared: %v", err)
		}
	})
}

// TestBlocksEachCombatWithNoLegalBlockIsNotOwed — "blocks each combat
// if able" asks nothing of a creature that cannot block: every attacker
// flies, or it is tapped.
func TestBlocksEachCombatWithNoLegalBlockIsNotOwed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := pushCombatant(t, g, me, "Flyer", 2, 2, "flying")
	watch := pushCombatant(t, g, opp, "Watchdog", 1, 2)
	withBlockRequirement(t, g, watch, BlockRequirementBlocks)
	tapped := pushCombatant(t, g, opp, "Tapped Wall", 0, 4, "reach")
	withBlockRequirement(t, g, tapped, BlockRequirementBlocks)
	g.WithWriteLock(func() { findBattlefieldCard(g, tapped).Tapped = true })
	declareAttacks(t, g, flyer)
	if got := g.BlockDeclarationStatusOf(opp.ID); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q: nothing can block the flyer, so nothing is owed", got)
	}
}

// TestBlockRequirementYieldsToARestriction — a creature that must block
// but "can't block" owes nothing; and a whole-combat limit (Silent
// Arbiter) lets one creature block, so of two creatures that must block
// either is a legal answer while a creature with no requirement taking
// the one slot is refused.
func TestBlockRequirementYieldsToARestriction(t *testing.T) {
	t.Run("can't block", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushCombatant(t, g, me, "Bear", 2, 2)
		pacified := pushCombatant(t, g, opp, "Pacified", 2, 2)
		withBlockRequirement(t, g, pacified, BlockRequirementBlocks)
		registerScopedEffectForTest(t, g, pacified, []Mod{AddRestrictionsMod(CantBlock)}, IndefiniteDuration())
		pushCombatant(t, g, opp, "Free Wall", 0, 4)
		declareAttacks(t, g, atk)
		passToDefender(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("a creature that can't block owes no block: %v", err)
		}
	})
	t.Run("Silent Arbiter", func(t *testing.T) {
		stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushLimitSource(t, g, me, "Silent Arbiter")
		atk := pushCombatant(t, g, me, "Bear", 2, 2)
		must1 := pushCombatant(t, g, opp, "Must One", 1, 1)
		withBlockRequirement(t, g, must1, BlockRequirementBlocks)
		must2 := pushCombatant(t, g, opp, "Must Two", 1, 1)
		withBlockRequirement(t, g, must2, BlockRequirementBlocks)
		free := pushCombatant(t, g, opp, "Free Wall", 0, 4)
		declareAttacks(t, g, atk)
		blockRequirementErr(t, block(t, g, free, atk))
		if err := block(t, g, must2, atk); err != nil {
			t.Fatalf("either creature that must block is a legal answer: %v", err)
		}
		passToDefender(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("the limit allows one block, and one requirement is obeyed: %v", err)
		}
		_ = must1
	})
}

// TestMustBeBlockedAndExactlyOne — "must be blocked if able" is obeyed
// by any one blocker; "must be blocked by exactly one creature if able"
// by exactly one, so a second blocker on it is refused.
func TestMustBeBlockedAndExactlyOne(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mb := pushCombatant(t, g, me, "Gaea's Protector", 3, 5)
	withBlockRequirement(t, g, mb, BlockRequirementMustBeBlocked)
	one := pushCombatant(t, g, me, "War-Pride", 3, 3)
	withBlockRequirement(t, g, one, BlockRequirementExactlyOne)
	a := pushCombatant(t, g, opp, "Wall A", 0, 4)
	b := pushCombatant(t, g, opp, "Wall B", 0, 4)
	c := pushCombatant(t, g, opp, "Wall C", 0, 4)
	declareAttacks(t, g, mb, one)
	passToDefender(t, g)
	br := blockRequirementErr(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Gaea's Protector must be blocked if able." &&
		got != "War-Pride must be blocked by exactly one creature if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := block(t, g, a, one); err != nil {
		t.Fatalf("one blocker on the exactly-one attacker: %v", err)
	}
	br = blockRequirementErr(t, block(t, g, b, one))
	if br.Requirement.Kind != BlockRequirementExactlyOne {
		t.Errorf("a second blocker on it gives up %q, want exactlyOne", br.Requirement.Kind)
	}
	blockRequirementErr(t, g.PassPriority())
	if err := block(t, g, b, mb, c, mb); err != nil {
		t.Fatalf("two blockers on the must-be-blocked attacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both obeyed: %v", err)
	}
}

// TestBlockRequirementIsJudgedOnlyWhilePending — once a defender's
// declaration is complete, a requirement that arrives afterwards asks
// nothing of it, and one that is on an attacker another player defends
// asks nothing of this defender.
func TestBlockRequirementIsJudgedOnlyWhilePending(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Bear", 2, 2)
	pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, atk)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	withBlockRequirement(t, g, atk, BlockRequirementMustBeBlocked)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if e := g.blockRequirementsUnmetLocked(opp.ID); e != nil {
			t.Errorf("a requirement after completion re-opened the declaration: %v", e)
		}
	})
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep after completion: %v", err)
	}
}

// TestBlockRequirementSurvivesUndoAndSnapshot — the requirement is a
// data record, so an undo (Clone / RestoreFrom) and a persisted
// snapshot both keep enforcing it.
func TestBlockRequirementSurvivesUndoAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Lured Bear", 2, 2)
	withBlockRequirement(t, g, atk, BlockRequirementLure)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, atk)
	passToDefender(t, g)
	saved := g.Clone()

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	blockRequirementErr(t, restored.PassPriority())

	if err := block(t, g, w, atk); err != nil {
		t.Fatal(err)
	}
	g.RestoreFrom(saved)
	blockRequirementErr(t, g.PassPriority())
	if err := block(t, g, w, atk); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass after the undo and the block: %v", err)
	}
}

// TestBlockRequirementTableWithoutRequirementsIsUnchanged — the fast
// path: a table with no requirement declares, passes and advances as
// ever.
func TestBlockRequirementTableWithoutRequirementsIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Bear", 2, 2)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, atk)
	g.WithWriteLock(func() {
		if g.anyBlockRequirementLocked(opp.ID) {
			t.Error("a table with no requirement took the slow path")
		}
		if g.MustBlockForEffect() != nil {
			t.Error("must_block with no requirement")
		}
	})
	passToDefender(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("declining to block: %v", err)
	}
	_ = w
}

// TestBlockRequirementFlowPicksTheMaximum — the search counts: under a
// one-block limit, the one creature that must block takes the attacker
// where it obeys the most requirements (its own plus two Lures), and
// the slot spent anywhere else is refused.
func TestBlockRequirementFlowPicksTheMaximum(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, me, "Silent Arbiter")
	lure := pushCombatant(t, g, me, "Double Lure", 2, 2)
	withBlockRequirement(t, g, lure, BlockRequirementLure, BlockRequirementLure)
	plain := pushCombatant(t, g, me, "Plain", 2, 2)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	withBlockRequirement(t, g, w, BlockRequirementBlocks)
	declareAttacks(t, g, lure, plain)
	var owed []BlockDeclaration
	g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
	if len(owed) != 1 || owed[0].Blocker != w || owed[0].Attacker != lure {
		t.Fatalf("witness = %+v, want the Wall on the double Lure (three requirements)", owed)
	}
	blockRequirementErr(t, block(t, g, w, plain))
}
