package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gogo_mysterious_mime_test.go — Gogo, Mysterious Mime (#3774, deck
// #3655). A beginning-of-combat "you may become a copy of another
// target creature you control until end of turn, except its name is
// Gogo. If you do, you both get +2/+0 and haste and attack this turn if
// able."

const gogoOracle = "6fdcb563-9279-4365-8a78-26e75fbfd917"

func gogoInPlay(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Gogo, Mysterious Mime", "Legendary Creature — Wizard", gogoOracle, 2, 2)
}

// gogoToTheTrigger walks to the beginning of combat, picks `target` for
// Gogo's trigger and lets it resolve up to the "you may".
func gogoToTheTrigger(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, target)
}

func TestGogoRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gogoOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered=%v completeness=%v caveats=%v", ok, spec.Completeness, spec.Caveats)
	}
}

// Yes: Gogo is the Giant (named Gogo), both get +2/+0 and haste, and
// both are required to attack.
func TestGogoBecomesACopyThenBothGetPumpedAndMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	gogo := gogoInPlay(g, me.ID)
	giant := b12Push(g, me.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	gogoToTheTrigger(t, g, me, giant)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c := cardByID(g, gogo)
	if c.Name != "Gogo, Mysterious Mime" {
		t.Errorf("the copy is named %q, want Gogo, Mysterious Mime", c.Name)
	}
	if !containsString(effectiveSubtypes(t, g, gogo), "Giant") {
		t.Errorf("Gogo did not become a Giant: %v", effectiveSubtypes(t, g, gogo))
	}
	if got := effectivePower(t, g, gogo); got != 5 {
		t.Errorf("Gogo's power %d, want 5 (a 3/3 copy plus 2)", got)
	}
	if got := effectivePower(t, g, giant); got != 5 {
		t.Errorf("the Giant's power %d, want 5", got)
	}
	for name, id := range map[string]uuid.UUID{"Gogo": gogo, "the Giant": giant} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", name)
		}
	}

	// Both are required to attack.
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with both home = %v, want an attack-requirement refusal", err)
	}
	if err := g.DeclareAttacker(gogo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Gogo): %v", err)
	}
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != giant {
		t.Fatalf("pass with the Giant home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Giant): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both attacking: %v", err)
	}
}

// All of it ends with the turn: Gogo is Gogo again, the pump is gone,
// and nothing is required.
func TestGogoEverythingEndsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gogo := gogoInPlay(g, me.ID)
	giant := b12Push(g, me.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	gogoToTheTrigger(t, g, me, giant)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	// Both are required to attack, so the turn cannot move on until they do.
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToDeclareAttackersOf(t, g, seat)
	for _, id := range []uuid.UUID{gogo, giant} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))

	if got := effectivePower(t, g, gogo); got != 2 {
		t.Errorf("Gogo's power %d next turn, want its own 2", got)
	}
	if got := effectivePower(t, g, giant); got != 3 {
		t.Errorf("the Giant's power %d next turn, want 3", got)
	}
	if containsString(effectiveSubtypes(t, g, gogo), "Giant") {
		t.Error("Gogo is still a Giant")
	}
	if hasRequirementRecord(g) {
		t.Error("\"this turn\" outlived the turn")
	}
}

// No: nothing happens.
func TestGogoDecliningChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gogo := gogoInPlay(g, me.ID)
	giant := b12Push(g, me.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	gogoToTheTrigger(t, g, me, giant)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, gogo); got != 2 {
		t.Errorf("Gogo's power %d, want 2", got)
	}
	if got := effectivePower(t, g, giant); got != 3 {
		t.Errorf("the Giant's power %d, want 3", got)
	}
	if hasRequirementRecord(g) {
		t.Error("a declined Gogo still forced an attack")
	}
}

// "Another": Gogo alone has no legal target, so the trigger never goes
// on the stack and nothing is asked (CR 603.3d).
func TestGogoAloneIsNeverAsked(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	gogo := gogoInPlay(g, me.ID)
	// An opponent's creature is not "a creature you control".
	b12Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	if latestConfirmFor(g, me.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("Gogo asked about a target with no legal one")
	}
	if got := effectivePower(t, g, gogo); got != 2 {
		t.Errorf("Gogo's power %d, want 2", got)
	}
	if hasRequirementRecord(g) {
		t.Error("a trigger with no target still forced an attack")
	}
}

// "If you do": Gogo flickered while the trigger waited is a new object
// (CR 400.7). The copy does not begin, so the other creature gets
// nothing either.
func TestGogoThatCameBackAsANewObjectDoesNothingAtAll(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gogo := gogoInPlay(g, me.ID)
	giant := b12Push(g, me.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	gogoToTheTrigger(t, g, me, giant)
	flickerInResponse(t, g, gogo)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		answerMayChoice(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}

	if got := effectivePower(t, g, giant); got != 3 {
		t.Errorf("the Giant's power %d, want 3 (no copy, so no \"if you do\")", got)
	}
	if hasEffectiveKeyword(t, g, giant, "haste") {
		t.Error("the Giant gained haste though the copy never began")
	}
	if hasRequirementRecord(g) {
		t.Error("a copy that never began still forced an attack")
	}
	if got := effectivePower(t, g, gogo); got != 2 {
		t.Errorf("the new Gogo's power %d, want its own 2", got)
	}
}

// The target leaving before the trigger resolves: it fizzles (CR
// 608.2b), with no prompt and no effect on Gogo.
func TestGogoTargetGoneBeforeResolutionDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gogo := gogoInPlay(g, me.ID)
	giant := b12Push(g, me.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	gogoToTheTrigger(t, g, me, giant)
	b25Destroy(g, giant)
	passPriorityAroundTable(t, g)

	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked about copying a creature that is gone")
	}
	if got := effectivePower(t, g, gogo); got != 2 {
		t.Errorf("Gogo's power %d, want 2", got)
	}
	if hasRequirementRecord(g) {
		t.Error("a fizzled trigger still forced an attack")
	}
}
