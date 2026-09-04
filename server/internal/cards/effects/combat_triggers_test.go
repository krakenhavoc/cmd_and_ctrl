package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_triggers_test.go — S19 sub-PR 7: "deals combat damage to a
// player" triggers. Seat 0 is the active player; its creatures
// attack seat 1. Combat damage lands in the combat damage step
// (AdvanceStep into it resolves damage + runs SBAs + drains
// triggers), so the trigger is on the stack — or its prompt queued
// — as soon as the step is entered.

const (
	bidentOracle      = "e1afaef7-9fa3-4662-a95f-adfb0da9fd11"
	coastalPiracyOrcl = "8a05ec32-7b0c-4f23-a4f7-413301c2a70a"
	scrollThiefOracle = "637c5583-4683-4ae4-8b4e-f5da42a772c7"
	edricOracle       = "9a1de7e4-9930-4db3-a8f3-d146d0abf38b"
	// Lightning Bolt, for the non-combat-damage control case. Named
	// distinctly from the sub-PR 6 test file's constant so the two
	// branches merge without a collision.
	boltOracleCombat = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// advanceTo walks AdvanceStep until the cursor is on step.
func advanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40 && g.Turn.Step != step; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s: %v", step, err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("never reached %s (at %s)", step, g.Turn.Step)
	}
}

// attackWith declares every given creature as attacking defender
// and advances into the combat damage step.
func attackWith(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", id, err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
}

// pushVanillaCreature seeds a non-catalog creature that can attack
// (EnteredBattlefieldAt is zero, so no summoning sickness).
func pushVanillaCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, name, "", "Creature — Test", power, toughness)
}

// --- Bident of Thassa -----------------------------------------

func TestBidentDrawsPerCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bidentID := pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	handBefore := me.Hand.Size()

	attackWith(t, g, opp.ID, a, b)
	if opp.Life != 40-4 {
		t.Fatalf("opponent life %d, want 36 (both bears connected)", opp.Life)
	}
	// Two "you may" prompts — one per creature that connected.
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID && c.Source == bidentID {
			prompts++
		}
	}
	if prompts != 2 {
		t.Fatalf("Bident prompts: %d, want 2", prompts)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if triggerOnStack(g, bidentID) == nil {
		t.Fatalf("accepted Bident trigger not on the stack")
	}
	if me.Hand.Size() != handBefore {
		t.Fatalf("draw happened before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Bident: hand delta %d, want 1 (one yes, one no)", got)
	}
}

func TestBidentIgnoresBlockedAttackerAndNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bidentID := pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	a := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 5)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == bidentID {
			t.Errorf("Bident prompted on damage dealt to a blocking creature")
		}
	}

	// Non-combat damage to the opponent: no trigger either.
	advanceTo(t, g, game.StepPostcombatMain)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == bidentID {
			t.Errorf("Bident prompted on Lightning Bolt damage")
		}
	}
}

func TestBidentIgnoresOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	// Bident under the OPPONENT; my creature connects with them.
	bidentID := pushPermanentForTest(g, opp.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	a := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	attackWith(t, g, opp.ID, a)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == bidentID {
			t.Errorf("opponent's Bident prompted on MY creature's combat damage")
		}
	}
}

// --- Coastal Piracy -------------------------------------------

func TestCoastalPiracyDrawsOnConnect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	piracyID := pushPermanentForTest(g, me.ID, "Coastal Piracy", coastalPiracyOrcl, "Enchantment")
	a := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	handBefore := me.Hand.Size()

	attackWith(t, g, opp.ID, a)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if triggerOnStack(g, piracyID) == nil {
		t.Fatalf("Coastal Piracy trigger not on the stack after Yes")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Coastal Piracy: hand delta %d, want 1", got)
	}
}

// --- Scroll Thief ---------------------------------------------

func TestScrollThiefSelfOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	thiefID := pushDiesCreatureForTest(g, me.ID, "Scroll Thief", scrollThiefOracle, "Creature — Merfolk Rogue", 1, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	handBefore := me.Hand.Size()

	// Only the bear attacks: no Thief trigger.
	attackWith(t, g, opp.ID, bear)
	if triggerOnStack(g, thiefID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatalf("Scroll Thief triggered on another creature's damage")
	}
	passPriorityAroundTable(t, g)

	// Next turn cycle: the Thief itself connects.
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	// Past the draw step before taking the baseline, so the turn
	// draw doesn't pollute the delta.
	advanceTo(t, g, game.StepPrecombatMain)
	handBefore = me.Hand.Size()
	attackWith(t, g, opp.ID, thiefID)
	trig := triggerOnStack(g, thiefID)
	if trig == nil {
		t.Fatalf("Scroll Thief did not trigger on its own combat damage")
	}
	if trig.Controller != me.ID {
		t.Errorf("trigger controller %s, want %s", trig.Controller, me.ID)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Scroll Thief: hand delta %d, want 1", got)
	}
}

// --- Edric, Spymaster of Trest --------------------------------

func TestEdricLetsAttackingCreaturesControllerDraw(t *testing.T) {
	g := newCatalogGame(t)
	attacker := g.Seats[0]
	edricOwner := g.Seats[2]
	victim := g.Seats[1]
	edricID := pushDiesCreatureForTest(g, edricOwner.ID, "Edric, Spymaster of Trest", edricOracle,
		"Legendary Creature — Elf Rogue", 2, 2)
	bear := pushVanillaCreature(g, attacker.ID, "Bear", 2, 2)
	attackerHand := attacker.Hand.Size()
	ownerHand := edricOwner.Hand.Size()

	// Seat 0's bear hits seat 1 — one of seat 2's opponents.
	attackWith(t, g, victim.ID, bear)

	// The prompt goes to the BEAR's controller, not Edric's.
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == edricID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no Edric prompt queued")
	}
	if prompt.Chooser != attacker.ID {
		t.Fatalf("Edric prompt chooser = %s, want the attacking creature's controller %s", prompt.Chooser, attacker.ID)
	}
	answerLatestTriggerPrompt(t, g, attacker.ID, true)
	trig := triggerOnStack(g, edricID)
	if trig == nil {
		t.Fatalf("Edric trigger not on the stack after Yes")
	}
	// The ability is Edric's controller's (it's their permanent) …
	if trig.Controller != edricOwner.ID {
		t.Errorf("trigger controller %s, want Edric's controller %s", trig.Controller, edricOwner.ID)
	}
	passPriorityAroundTable(t, g)
	// … but the card goes to the attacker.
	if got := attacker.Hand.Size() - attackerHand; got != 1 {
		t.Errorf("attacker hand delta %d, want 1", got)
	}
	if got := edricOwner.Hand.Size() - ownerHand; got != 0 {
		t.Errorf("Edric's controller hand delta %d, want 0", got)
	}
}

func TestEdricIgnoresDamageToItsOwnController(t *testing.T) {
	g := newCatalogGame(t)
	attacker := g.Seats[0]
	edricOwner := g.Seats[1]
	edricID := pushDiesCreatureForTest(g, edricOwner.ID, "Edric, Spymaster of Trest", edricOracle,
		"Legendary Creature — Elf Rogue", 2, 2)
	bear := pushVanillaCreature(g, attacker.ID, "Bear", 2, 2)

	// Hitting Edric's own controller is not "one of your opponents".
	attackWith(t, g, edricOwner.ID, bear)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == edricID {
			t.Errorf("Edric prompted on damage dealt to its own controller")
		}
	}
}
