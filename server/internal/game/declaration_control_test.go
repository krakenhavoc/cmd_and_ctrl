package game

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// declaration_control_test.go — #4154: "you choose which creatures
// attack / block" (Master Warcraft, Odric), driven through
// BeginDeclarationControlForEffect and the game verbs. Routing, the
// enumerator, the view and the cards are pinned in their own packages.

const warcraftName = "Master Warcraft"

// chooseDeclarations gives seat `declarer` a this-turn declaration
// control.
func chooseDeclarations(t *testing.T, g *Game, declarer int, attackers, blockers, thisCombat bool) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginDeclarationControlForEffect(DeclarationControl{
			Declarer:   g.Seats[declarer].ID,
			Source:     uuid.New(),
			SourceName: warcraftName,
			Attackers:  attackers,
			Blockers:   blockers,
			ThisCombat: thisCombat,
		})
	})
	if !ok {
		t.Fatalf("BeginDeclarationControlForEffect(seat %d) refused", declarer)
	}
}

func declaredAttackEvents(g *Game, card uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventAttack && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

func handoffStage(g *Game) AttackHandoffStage {
	if g.AttackHandoff == nil {
		return "<nil>"
	}
	return g.AttackHandoff.Stage
}

// The declarer chooses the attackers on the active player's behalf
// (the active player is "controlled" for that one choice), the active
// player then points them, and only then is the declaration locked in.
func TestDeclarationControlTheDeclarerChoosesTheAttackersAndTheActivePlayerTheirTargets(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, true, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	other := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := controllerOfSeat(g, 0); got != 1 {
		t.Fatalf("seat 0's declaration is made by seat %d, want the declarer (seat 1)", got)
	}
	if got := g.ControlStageOfPlayer(g.Seats[0].ID); got != ControlDeclaring {
		t.Errorf("ControlStageOfPlayer = %q, want %q", got, ControlDeclaring)
	}
	if got := g.ControlUntilOfPlayer(g.Seats[0].ID); got != ControlUntilDeclarations || got.PlaysForThem() {
		t.Errorf("ControlUntilOfPlayer = %q (plays for them %v), want declarations", got, got.PlaysForThem())
	}
	var live []PlayerControl
	g.ReadSnapshot(func() { live = g.LivePlayerControlsLocked() })
	if len(live) != 1 || live[0].Controller != g.Seats[1].ID || live[0].Controlled != g.Seats[0].ID || live[0].Stage != ControlDeclaring {
		t.Fatalf("live controls = %+v, want the one declaration", live)
	}
	if g.ControlsATurn(g.Seats[1].ID) {
		t.Error("the declarer reads as playing the active player's turn")
	}

	// The declarer stages one attacker, at a target of their choosing.
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatalf("the declarer's declaration: %v", err)
	}
	// The pass is the declarer's: it finishes their half, and nothing
	// is announced yet.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the declarer's finishing pass: %v", err)
	}
	if got := handoffStage(g); got != AttackHandoffTargets {
		t.Fatalf("handoff stage = %q, want %q", got, AttackHandoffTargets)
	}
	if n := len(declaredAttackEvents(g, bear)); n != 0 {
		t.Fatalf("%d EventAttack before the active player chose the target", n)
	}
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0 decided by seat %d in their own half", got)
	}
	if g.Turn.PriorityHolder != 0 || g.Turn.Step != StepDeclareAttackers {
		t.Fatalf("the declarer's pass moved the table: %+v", g.Turn)
	}
	// The active player can't add an attacker the declarer did not
	// choose, and points the chosen one elsewhere.
	if err := g.DeclareAttacker(other, g.Seats[2].ID); !errors.Is(err, ErrAttackersChosen) {
		t.Fatalf("the active player's own attacker: %v, want ErrAttackersChosen", err)
	}
	if err := g.DeclareAttacker(bear, g.Seats[3].ID); err != nil {
		t.Fatalf("the active player's re-point: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the active player's pass: %v", err)
	}
	if got := handoffStage(g); got != AttackHandoffChosen {
		t.Fatalf("handoff stage = %q after both halves, want %q", got, AttackHandoffChosen)
	}
	evs := declaredAttackEvents(g, bear)
	if len(evs) != 1 || evs[0].Target != g.Seats[3].ID || evs[0].Actor != g.Seats[0].ID {
		t.Fatalf("EventAttack = %+v, want one, at seat 3, by seat 0", evs)
	}
	if c := findCard(g, other); c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Errorf("the creature nobody chose is attacking=%v tapped=%v", c.AttackingTarget, c.Tapped)
	}
}

// Choosing no attackers ends the declaration: the active player may not
// attack on their own afterwards.
func TestDeclarationControlChoosingNoAttackersClosesTheDeclaration(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.PassPriority(); err != nil {
		t.Fatalf("the declarer's empty declaration: %v", err)
	}
	if got := handoffStage(g); got != AttackHandoffChosen {
		t.Fatalf("handoff stage = %q, want %q", got, AttackHandoffChosen)
	}
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0 still decided by seat %d", got)
	}
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); !errors.Is(err, ErrAttackersChosen) {
		t.Fatalf("an attack after the declarer chose none: %v, want ErrAttackersChosen", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[2].ID}}); !errors.Is(err, ErrAttackersChosen) {
		t.Fatalf("a bulk attack after the declarer chose none: %v, want ErrAttackersChosen", err)
	}
	advanceIntoStep(t, g, StepEndCombat)
}

// With nothing that could attack there is nothing to choose: the
// declaration is over as the step begins, and the active player is
// never handed to the declarer.
func TestDeclarationControlNothingToChooseEndsTheDeclarationAtOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	tapped := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	findCard(g, tapped).Tapped = true
	advanceIntoStep(t, g, StepDeclareAttackers)
	if got := handoffStage(g); got != AttackHandoffChosen {
		t.Fatalf("handoff stage = %q at the step's start, want %q", got, AttackHandoffChosen)
	}
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0 decided by seat %d with nothing to choose", got)
	}
}

// CR 508.1d still binds the declarer: their pass is refused while a
// creature that must attack is left home.
func TestDeclarationControlTheDeclarerObeysAttackRequirements(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	zurgo := mustAttackCreature(t, g, g.Seats[0], 1)
	advanceIntoStep(t, g, StepDeclareAttackers)

	requirementErr(t, g.PassPriority())
	if got := handoffStage(g); got != AttackHandoffDeclarer {
		t.Fatalf("a refused finish moved the handoff to %q", got)
	}
	if err := g.DeclareAttacker(zurgo, g.Seats[1].ID); err != nil {
		t.Fatalf("declaring the required attacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the finish after the requirement is met: %v", err)
	}
	if got := handoffStage(g); got != AttackHandoffTargets {
		t.Fatalf("handoff stage = %q, want %q", got, AttackHandoffTargets)
	}
}

// Exerting is a cost to attack, so it is the active player's choice,
// never the declarer's.
func TestDeclarationControlTheDeclarerCantExert(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: bear, Target: g.Seats[2].ID, Exert: true}, DeclareAttackersParams{}); !errors.Is(err, ErrDeclarerCantExert) {
		t.Fatalf("the declarer's exert: %v, want ErrDeclarerCantExert", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[2].ID, Exert: true}}); !errors.Is(err, ErrDeclarerCantExert) {
		t.Fatalf("the declarer's bulk exert: %v, want ErrDeclarerCantExert", err)
	}
}

// The attack tax is the active player's: the declarer's choice pays
// nothing, a pass is refused while a free target would avoid the cost,
// and the active player's confirmation pays it.
func TestDeclarationControlTheActivePlayerPaysTheAttackTaxOnConfirming(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	registerTimedTax(t, g, g.Seats[2].ID, "{2}")
	addMana(g.Seats[0], 2)

	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatalf("the declarer's choice at a taxed target: %v", err)
	}
	if n := g.Seats[0].ManaPool.Total(); n != 2 {
		t.Fatalf("the declarer's choice spent the active player's mana: %d left", n)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the declarer's finish: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, ErrAttackCostsUndecided) {
		t.Fatalf("the active player's pass with a cost a free target avoids: %v, want ErrAttackCostsUndecided", err)
	}
	if n := len(declaredAttackEvents(g, bear)); n != 0 {
		t.Fatalf("a refused pass announced the attack (%d)", n)
	}
	got, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[2].ID}})
	if err != nil {
		t.Fatalf("the active player's confirmation: %v", err)
	}
	if len(got) != 1 || got[0] != bear {
		t.Fatalf("confirmed %v, want the bear", got)
	}
	if n := g.Seats[0].ManaPool.Total(); n != 0 {
		t.Errorf("the confirmation left %d mana, want the {2} paid", n)
	}
	if got := handoffStage(g); got != AttackHandoffChosen {
		t.Fatalf("handoff stage = %q, want %q", got, AttackHandoffChosen)
	}
	if evs := declaredAttackEvents(g, bear); len(evs) != 1 || evs[0].Target != g.Seats[2].ID {
		t.Fatalf("EventAttack = %+v, want one at seat 2", evs)
	}
}

// A re-point to a free target avoids the cost, and the pass accepts it.
func TestDeclarationControlARepointAvoidsTheTax(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	registerTimedTax(t, g, g.Seats[2].ID, "{2}")
	addMana(g.Seats[0], 2)

	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(bear, g.Seats[3].ID); err != nil {
		t.Fatalf("the re-point: %v", err)
	}
	if n := g.Seats[0].ManaPool.Total(); n != 2 {
		t.Fatalf("the re-point charged something: %d left", n)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the accepting pass: %v", err)
	}
	if evs := declaredAttackEvents(g, bear); len(evs) != 1 || evs[0].Target != g.Seats[3].ID {
		t.Fatalf("EventAttack = %+v, want one at seat 3", evs)
	}
}

// A cost the active player won't pay, on a creature that can't attack
// anything for free, takes the creature out, and the declarer chooses
// again without it (the Odric ruling).
func TestDeclarationControlADeclinedCostGoesBackToTheDeclarer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	other := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, s := range []int{1, 2, 3} {
		registerTimedTax(t, g, g.Seats[s].ID, "{1}")
	}

	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	// Every target costs {1}: the pass declines it.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the declining pass: %v", err)
	}
	if got := handoffStage(g); got != AttackHandoffDeclarer {
		t.Fatalf("handoff stage = %q, want the declarer's again", got)
	}
	if c := findCard(g, bear); c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Fatalf("the declined creature is attacking=%v tapped=%v", c.AttackingTarget, c.Tapped)
	}
	if got := controllerOfSeat(g, 0); got != 1 {
		t.Fatalf("the declaration went back to seat %d, want the declarer", got)
	}
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); !errors.Is(err, ErrAttackDeclined) {
		t.Fatalf("choosing the declined creature again: %v, want ErrAttackDeclined", err)
	}
	if err := g.DeclareAttacker(other, g.Seats[3].ID); err != nil {
		t.Fatalf("choosing another: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	// The confirmation may leave out a creature only when it has no
	// free target; this one has none, so leaving it out declines again.
	if got, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: other, Target: g.Seats[3].ID}}); err == nil {
		// Unpaid: the pool is empty, so this is refused below.
		t.Fatalf("an unpaid confirmation was accepted: %v", got)
	}
	addMana(g.Seats[0], 1)
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: other, Target: g.Seats[3].ID}}); err != nil {
		t.Fatalf("the paid confirmation: %v", err)
	}
	if evs := declaredAttackEvents(g, other); len(evs) != 1 {
		t.Fatalf("EventAttack = %+v, want one", evs)
	}
	if n := len(declaredAttackEvents(g, bear)); n != 0 {
		t.Fatalf("the declined creature attacked (%d)", n)
	}
}

// Leaving out a chosen attacker that could attack for free is refused.
func TestDeclarationControlTheConfirmationKeepsEveryChosenAttacker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	other := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[2].ID}, {Attacker: other, Target: g.Seats[2].ID}}); err != nil {
		t.Fatalf("the declarer's bulk choice: %v", err)
	}
	if n := len(declaredAttackEvents(g, bear)); n != 0 {
		t.Fatal("the declarer's bulk choice was locked in")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[3].ID}}); !errors.Is(err, ErrChosenAttackerMustAttack) {
		t.Fatalf("leaving out a free attacker: %v, want ErrChosenAttackerMustAttack", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: g.Seats[3].ID}, {Attacker: other, Target: g.Seats[1].ID}}); err != nil {
		t.Fatalf("the full confirmation: %v", err)
	}
	if evs := declaredAttackEvents(g, other); len(evs) != 1 || evs[0].Target != g.Seats[1].ID {
		t.Fatalf("EventAttack = %+v, want one at seat 1", evs)
	}
}

// The step can't be skipped past an unfinished chosen declaration.
func TestDeclarationControlAdvanceStepWaitsForTheDeclaration(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrAttackDeclarationOpen) {
		t.Fatalf("AdvanceStep in the declarer's half: %v, want ErrAttackDeclarationOpen", err)
	}
}

// The declarer declares every other defending player's blocks; their
// own stay theirs.
func TestDeclarationControlTheDeclarerDeclaresEveryOtherDefendersBlocks(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, false, true, false)
	atk1 := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	atk2 := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	atk3 := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	wall1 := pushKeywordCreature(t, g, g.Seats[1], 0, 4)
	wall2 := pushKeywordCreature(t, g, g.Seats[2], 0, 4)
	wall3 := pushKeywordCreature(t, g, g.Seats[3], 0, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("a blockers-only record chose seat 0's attackers (seat %d)", got)
	}
	for i, a := range []uuid.UUID{atk1, atk2, atk3} {
		if err := g.DeclareAttacker(a, g.Seats[i+1].ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if got := controllerOfSeat(g, 1); got != 1 {
		t.Errorf("the declarer's own blocks are made by seat %d", got)
	}
	for _, s := range []int{2, 3} {
		if got := controllerOfSeat(g, s); got != 1 {
			t.Fatalf("seat %d's blocks are made by seat %d, want the declarer", s, got)
		}
		if got := g.ControlStageOfPlayer(g.Seats[s].ID); got != ControlDeclaring {
			t.Errorf("seat %d stage = %q", s, got)
		}
	}
	// Seat 2's wall blocks; seat 3's doesn't.
	if err := g.DeclareBlocker(wall2, atk2); err != nil {
		t.Fatalf("the declarer's block for seat 2: %v", err)
	}
	if err := g.FinishBlocks(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if got := controllerOfSeat(g, 2); got != 2 {
		t.Errorf("seat 2 still decided by seat %d after their declaration", got)
	}
	if err := g.FinishBlocks(g.Seats[3].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(wall1, atk1); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if c := findCard(g, wall3); c.BlockingTarget != uuid.Nil {
		t.Error("seat 3's creature blocks though the declarer chose no blocks")
	}
	if c := findCard(g, wall2); c.BlockingTarget != atk2 {
		t.Errorf("seat 2's wall blocks %v, want the attacker the declarer chose", c.BlockingTarget)
	}
}

// ADR 0167's ruling: the declaration effect takes precedence over a
// player controlling the active player's turn, and the turn's
// controller makes the active player's own half.
func TestDeclarationControlTakesPrecedenceOverTurnControl(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 3, 1)
	bear := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if got := controllerOfSeat(g, 1); got != 3 {
		t.Fatalf("setup: seat 1's turn is controlled by seat %d", got)
	}
	chooseDeclarations(t, g, 2, true, false, false)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := controllerOfSeat(g, 1); got != 2 {
		t.Fatalf("seat 1's attack declaration is made by seat %d, want the declarer", got)
	}
	var live []PlayerControl
	g.ReadSnapshot(func() { live = g.LivePlayerControlsLocked() })
	if len(live) != 1 || live[0].Until != ControlUntilDeclarations {
		t.Fatalf("live = %+v, want the declaration alone over seat 1", live)
	}
	if err := g.DeclareAttacker(bear, g.Seats[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if got := controllerOfSeat(g, 1); got != 3 {
		t.Fatalf("seat 1's own half is decided by seat %d, want the turn's controller", got)
	}
}

// A this-turn record lasts the turn; a this-combat one ends with its
// combat phase.
func TestDeclarationControlLifetimes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	chooseDeclarations(t, g, 1, false, true, true)
	chooseDeclarations(t, g, 2, true, true, false)
	if n := len(g.DeclarationControls); n != 2 {
		t.Fatalf("records = %d", n)
	}
	// Seat 2's record chooses this combat's attackers: nothing to
	// choose, so it finishes at once.
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepPostcombatMain)
	if n := len(g.DeclarationControls); n != 1 || g.DeclarationControls[0].ThisCombat {
		t.Fatalf("after combat: %+v, want the this-turn record alone", g.DeclarationControls)
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if n := len(g.DeclarationControls); n != 0 {
		t.Fatalf("next turn: %+v, want none", g.DeclarationControls)
	}
}

// A declarer who has left the game chooses nothing.
func TestDeclarationControlADeclarerWhoLeftChoosesNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, true, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0's declaration made by seat %d after the declarer left", got)
	}
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
}

// A table mid-declaration is a restore point.
func TestDeclarationControlSnapshotRoundTrip(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, true, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	back, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !reflect.DeepEqual(back.DeclarationControls, g.DeclarationControls) || !reflect.DeepEqual(back.AttackHandoff, g.AttackHandoff) {
		t.Fatalf("restored %+v / %+v, want %+v / %+v", back.DeclarationControls, back.AttackHandoff, g.DeclarationControls, g.AttackHandoff)
	}
	if err := back.DeclareAttacker(bear, back.Seats[3].ID); err != nil {
		t.Fatalf("the restored active player's re-point: %v", err)
	}
	if err := back.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if evs := declaredAttackEvents(back, bear); len(evs) != 1 || evs[0].Target != back.Seats[3].ID {
		t.Fatalf("restored EventAttack = %+v", evs)
	}
}

// Undo across the declarer's half restores it: the clone carries both.
func TestDeclarationControlCloneCarriesTheHandoff(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, true, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	before := g.Clone()
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if got := handoffStage(before); got != AttackHandoffDeclarer {
		t.Fatalf("the clone shares the live handoff: stage %q", got)
	}
	g.RestoreFrom(before)
	if got := handoffStage(g); got != AttackHandoffDeclarer || controllerOfSeat(g, 0) != 1 {
		t.Fatalf("after the undo: handoff %q, seat 0 decided by %d", got, controllerOfSeat(g, 0))
	}
}

// A declarer who leaves mid-way chooses nothing more: what they staged
// is taken back, and the active player declares for themselves.
func TestDeclarationControlADeclarerWhoLeavesMidwayHandsTheDeclarationBack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooseDeclarations(t, g, 1, true, false, false)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	other := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	registerTimedTax(t, g, g.Seats[2].ID, "{2}")
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0 decided by seat %d after the declarer left", got)
	}
	if err := g.DeclareAttacker(other, g.Seats[3].ID); err != nil {
		t.Fatalf("the active player's own declaration: %v", err)
	}
	if c := findCard(g, bear); c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Fatalf("the departed declarer's unpaid choice is still attacking=%v tapped=%v", c.AttackingTarget, c.Tapped)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if evs := declaredAttackEvents(g, other); len(evs) != 1 {
		t.Fatalf("EventAttack = %+v, want the active player's own attack", evs)
	}
}
