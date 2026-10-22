package game

import (
	"testing"

	"github.com/google/uuid"
)

// attack_permanent_requirement_test.go pins "attacks <that permanent>
// if able" (#2567, ADR 0045 amendment 2026-10-22): the requirement is
// obeyed only by an attack on that permanent as the object it named,
// it is counted by CR 508.1d's maximisation like every other, it is held
// back until a turn after the one it was made in, and the duration
// "during its controller's next turn" ends at the cleanup of a later
// turn whose active player controls the creature. The cards are pinned
// in cards/effects/attack_this_permanent_test.go.

// walkerRef is the ObjectRef of a permanent as it is now.
func walkerRef(g *Game, id uuid.UUID) ObjectRef {
	return ObjectRef{ID: id, Epoch: findCard(g, id).ObjectEpoch}
}

// Obeyed only by attacking that permanent: the empty declaration and an
// attack on its controller are refused, an attack on it is accepted, and
// the sentence names it.
func TestAttackPermanentRequirementIsObeyedOnlyByThatPermanent(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 5)
	other := pushPlaneswalkerForTest(g, opp.ID, "Other Walker", 5)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{AttackPermanentIfAbleMod(walkerRef(g, walker), 0)}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.PassPriority())
	if re.Attacker != bear {
		t.Errorf("refusal names %s, want the creature", re.Attacker)
	}
	if got := re.Sentence(me.ID); got != "Kw Creature must attack Test Walker if able." {
		t.Errorf("sentence = %q", got)
	}
	requirementErr(t, g.DeclareAttacker(bear, opp.ID))
	requirementErr(t, g.DeclareAttacker(bear, other))
	if err := g.DeclareAttacker(bear, walker); err != nil {
		t.Fatalf("attack the named permanent: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the requirement obeyed: %v", err)
	}
}

// A permanent that is no longer the object named (it left and came
// back, CR 400.7) cannot be attacked "as it", so nothing is owed.
func TestAttackPermanentRequirementOwesNothingToANewObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 5)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	stale := walkerRef(g, walker)
	stale.Epoch++
	registerScopedEffectForTest(t, g, bear, []Mod{AttackPermanentIfAbleMod(stale, 0)}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with only a new object to attack: %v, want nothing owed", err)
	}
}

// Counted by CR 508.1d: under a one-attacker limit, a creature that must
// attack the walker and a creature with no requirement compete for the
// slot, and the declaration that sends the free one is refused.
func TestAttackPermanentRequirementCountsInTheMaximisation(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, opp, "Silent Arbiter")
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 5)
	bound := pushKeywordCreature(t, g, me, 2, 2)
	free := pushKeywordCreature(t, g, me, 2, 2)
	registerScopedEffectForTest(t, g, bound, []Mod{AttackPermanentIfAbleMod(walkerRef(g, walker), 0)}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)
	requirementErr(t, g.DeclareAttacker(free, opp.ID))
	if err := g.DeclareAttacker(bound, walker); err != nil {
		t.Fatalf("the bound creature into the one slot: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// AfterTurn holds the requirement back for the rest of the turn it was
// made in — "during that player's NEXT turn" — and lets it bind in a
// later one.
func TestAttackPermanentRequirementIsHeldBackUntilALaterTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 5)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{AttackPermanentIfAbleMod(walkerRef(g, walker), g.Turn.Seq)}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the turn it was made in: %v, want nothing owed yet", err)
	}
	advanceToSeatStep(t, g, 0, StepDeclareAttackers, true)
	requirementErr(t, func() error { _, err := g.AdvanceStep(); return err }())
}

// "During its controller's next turn": the turn it was made in does not
// end it even when it is the creature's controller's own, another
// player's turn does not end it, and the cleanup of its controller's
// next turn does.
func TestUntilEndOfItsControllersNextTurnDuration(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0]
	bear := pushKeywordCreature(t, g, owner, 2, 2)
	var d Duration
	g.WithWriteLock(func() {
		var ok bool
		d, ok = g.UntilEndOfItsControllersNextTurnDuration(bear)
		if !ok {
			t.Fatal("no duration for a creature on the battlefield")
		}
	})
	if d.Problem() != "" {
		t.Fatalf("problem: %s", d.Problem())
	}
	registerScopedEffectForTest(t, g, bear, []Mod{AddKeywordsMod("flying")}, d)
	live := func() bool { return len(g.ScopedEffects) > 0 }

	advanceToSeatStep(t, g, 1, StepUpkeep, false)
	if !live() {
		t.Fatal("the turn it was made in ended it, though it is its controller's own")
	}
	advanceToSeatStep(t, g, 0, StepUpkeep, false)
	if !live() {
		t.Fatal("another player's turn ended it")
	}
	advanceToSeatStep(t, g, 1, StepUpkeep, false)
	if live() {
		t.Error("it outlived its controller's next turn")
	}

	g.WithWriteLock(func() {
		if _, ok := g.UntilEndOfItsControllersNextTurnDuration(uuid.New()); ok {
			t.Error("a duration for a creature that is not there")
		}
	})
}

// The shapes a restore point is refused for.
func TestAttackPermanentShapesRefuseMalformedValues(t *testing.T) {
	for _, c := range []struct {
		name string
		m    Mod
		bad  bool
	}{
		{"sound", AttackPermanentIfAbleMod(ObjectRef{ID: uuid.New()}, 3), false},
		{"no permanent", Mod{Kind: ModAttackPermanentIfAble}, true},
		{"negative stamp", Mod{Kind: ModAttackPermanentIfAble, Objects: []ObjectRef{{ID: uuid.New()}}, AfterTurn: -1}, true},
		{"stamp on another kind", Mod{Kind: ModAddAttackRequirement, AfterTurn: 2}, true},
	} {
		if got := attackPermanentModProblem(c.m) != ""; got != c.bad {
			t.Errorf("%s: problem = %q, want bad=%v", c.name, attackPermanentModProblem(c.m), c.bad)
		}
	}
	if (Duration{Kind: UntilEndOfItsControllersNextTurn, AfterTurn: 2}).Problem() == "" {
		t.Error("an until-its-controller's-next-turn duration with no creature is unknown")
	}
	if (Duration{Kind: UntilEndOfTurn, AfterTurn: 2}).Problem() == "" {
		t.Error("a next-turn stamp on another kind is unknown")
	}
	if !UntilEndOfItsControllersNextTurn.Known() || (UntilEndOfItsControllersNextTurn + 1).Known() {
		t.Error("the new kind is the last known kind")
	}
	if blockRequirementModProblem(AttackPermanentIfAbleMod(ObjectRef{ID: uuid.New()}, 0)) != "" {
		t.Error("the requirement's Objects are refused as a block requirement's")
	}
}

// advanceToSeatStep walks AdvanceStep until `seat`'s `step`. With
// fresh set, the cursor must first leave the step it is on.
func advanceToSeatStep(t *testing.T, g *Game, seat int, step Step, fresh bool) {
	t.Helper()
	moved := !fresh
	for i := 0; i < 400; i++ {
		if moved && g.Turn.ActiveSeat == seat && g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
		moved = true
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}
