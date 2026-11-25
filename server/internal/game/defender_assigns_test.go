package game

import (
	"testing"

	"github.com/google/uuid"
)

// defender_assigns_test.go — #6192 (ADR 0224 amendment 2026-11-25):
//
//   - Defensive Formation: "Rather than the attacking player, you assign
//     the combat damage of each creature attacking you. You can divide that
//     creature's combat damage as you choose among any of the creatures
//     blocking it." CR 510.1c's division of a blocked attacker's damage,
//     handed to the player it attacks, as CR 702.22j hands it over for
//     banding.
//   - Cunning Giant: "If this creature is unblocked, you may have it
//     assign its combat damage to a creature defending player controls."
//     CR 510.1b's unblocked assignment, with one creature as the other
//     answer.

const (
	formationOracle = "test-defensive-formation"
	giantOracle     = "test-cunning-giant"
)

func withAssignHooks(t *testing.T) {
	t.Helper()
	prevF, prevG := CatalogYouAssignAttackersDamage, CatalogUnblockedAssignsToACreature
	CatalogYouAssignAttackersDamage = func(key string) bool { return key == formationOracle }
	CatalogUnblockedAssignsToACreature = func(key string) bool { return key == giantOracle }
	t.Cleanup(func() { CatalogYouAssignAttackersDamage, CatalogUnblockedAssignsToACreature = prevF, prevG })
}

func setAssignOracle(g *Game, id uuid.UUID, oracle string) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].OracleID = oracle
		}
	}
}

func pushFormation(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Defensive Formation", owner.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = formationOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The issue's acceptance: with Defensive Formation, a 4-power attacker
// blocked by two creatures has its damage divided by the defending player,
// as they choose and with no trample over.
func TestDefensiveFormationDefenderDividesTheDamage(t *testing.T) {
	withAssignHooks(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushFormation(t, g, opp)
	beast := pushCombatant(t, g, me, "Trampling Beast", 4, 4, "trample")
	a := pushCombatant(t, g, opp, "Bear A", 2, 2)
	b := pushCombatant(t, g, opp, "Bear B", 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, a, beast, b, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil {
		t.Fatal("no prompt for the attacker's damage")
	}
	f := choice.DamageAssignment
	if choice.Chooser != opp.ID || f.AttackerID != beast || !f.Banding || f.AllowTrample || f.DefenderAssignsBy != "Defensive Formation" {
		t.Fatalf("prompt = chooser %s, frame %+v; want the defending player dividing for Defensive Formation", choice.Chooser, *f)
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: a, Amount: 2}, {BlockerID: b, Amount: 2}}, 0); err == nil {
		t.Fatal("the attacking player answered the defending player's division")
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: a, Amount: 4}, {BlockerID: b, Amount: 0}}, 0); err != nil {
		t.Fatalf("the defending player's division: %v", err)
	}
	if findBattlefieldCard(g, a) != nil {
		t.Error("Bear A took 4 and is still on the battlefield")
	}
	if got := damageOn(g, b); got != 0 {
		t.Errorf("Bear B has %d damage, want 0", got)
	}
	if opp.Life != StartingLife {
		t.Errorf("the defender lost %d life, want none", StartingLife-opp.Life)
	}
}

// Only creatures attacking the Formation's controller: the attacking
// player's own Formation hands nothing over.
func TestDefensiveFormationIsOnlyForCreaturesAttackingYou(t *testing.T) {
	withAssignHooks(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushFormation(t, g, me)
	beast := pushCombatant(t, g, me, "Beast", 4, 4)
	a := pushCombatant(t, g, opp, "Bear A", 2, 2)
	b := pushCombatant(t, g, opp, "Bear B", 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, a, beast, b, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil || choice.Chooser != me.ID || choice.DamageAssignment.Banding {
		t.Fatalf("prompt = %+v, want the attacker's controller's own division", choice)
	}
}

// Cunning Giant unblocked: its controller may put all of its damage on one
// creature the defending player controls…
func TestCunningGiantAssignsToACreature(t *testing.T) {
	withAssignHooks(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCombatant(t, g, me, "Cunning Giant", 4, 4)
	setAssignOracle(g, giant, giantOracle)
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 5)
	declareAttacks(t, g, giant)
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil || choice.Chooser != me.ID {
		t.Fatalf("prompt = %+v, want the Giant's controller asked", choice)
	}
	f := choice.DamageAssignment
	if !f.OneCreature || !f.AsThoughUnblocked || len(f.BlockerIDs) != 2 {
		t.Fatalf("frame = %+v, want one-creature over the defender's two creatures", *f)
	}
	// Divided between two creatures is not "a creature".
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 2}, {BlockerID: wall, Amount: 2}}, 0); err == nil {
		t.Fatal("a division among two creatures was accepted")
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 2}, {BlockerID: wall, Amount: 0}}, 2); err == nil {
		t.Fatal("a split between a creature and the player was accepted")
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 4}, {BlockerID: wall, Amount: 0}}, 0); err != nil {
		t.Fatalf("all of it to the Bear: %v", err)
	}
	if findBattlefieldCard(g, bear) != nil {
		t.Error("the Bear took 4 and is still on the battlefield")
	}
	if opp.Life != StartingLife {
		t.Errorf("the defender lost %d life, want none", StartingLife-opp.Life)
	}
}

// …or to the player, as it would anyway.
func TestCunningGiantMayStillHitThePlayer(t *testing.T) {
	withAssignHooks(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCombatant(t, g, me, "Cunning Giant", 4, 4)
	setAssignOracle(g, giant, giantOracle)
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)
	declareAttacks(t, g, giant)
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 0}}, 4); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != StartingLife-4 {
		t.Errorf("the defender's life = %d, want %d", opp.Life, StartingLife-4)
	}
	if got := damageOn(g, bear); got != 0 {
		t.Errorf("the Bear has %d damage, want 0", got)
	}
}

// With no creature to choose, or once it is blocked, there is no prompt.
func TestCunningGiantAsksOnlyWhenUnblockedWithACreatureToHit(t *testing.T) {
	withAssignHooks(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCombatant(t, g, me, "Cunning Giant", 4, 4)
	setAssignOracle(g, giant, giantOracle)
	declareAttacks(t, g, giant)
	advanceIntoStep(t, g, StepCombatDamage)
	if c := damageChoice(g); c != nil {
		t.Fatalf("a prompt with no creature to hit: %+v", *c)
	}
	if opp.Life != StartingLife-4 {
		t.Errorf("the defender's life = %d, want %d", opp.Life, StartingLife-4)
	}
}
