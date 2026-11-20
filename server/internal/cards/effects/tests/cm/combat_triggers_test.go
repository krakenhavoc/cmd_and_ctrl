package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Bident of Thassa -----------------------------------------

func TestBidentDrawsPerCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bidentID := effectstest.PushPermanentForTest(g, me.ID, "Bident of Thassa", effectstest.BidentOracle, "Legendary Enchantment Artifact")
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	handBefore := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, a, b)
	if opp.Life != 40-4 {
		t.Fatalf("opponent life %d, want 36 (both bears connected)", opp.Life)
	}
	// Two "you may"s — one per creature that connected, each its own
	// trigger on the stack, asked as it resolves (CR 603.5).
	if prompts := effectstest.OwedMays(g, me.ID); prompts != 2 {
		t.Fatalf("Bident prompts: %d, want 2", prompts)
	}
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, bidentID) == nil {
		t.Fatalf("Bident triggers not on the stack")
	}
	if me.Hand.Size() != handBefore {
		t.Fatalf("draw happened before the trigger resolved")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Bident: hand delta %d, want 1 (one yes, one no)", got)
	}
}

func TestBidentIgnoresBlockedAttackerAndNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bidentID := effectstest.PushPermanentForTest(g, me.ID, "Bident of Thassa", effectstest.BidentOracle, "Legendary Enchantment Artifact")
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 5)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == bidentID {
			t.Errorf("Bident prompted on damage dealt to a blocking creature")
		}
	}

	// Non-combat damage to the opponent: no trigger either.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == bidentID {
			t.Errorf("Bident prompted on Lightning Bolt damage")
		}
	}
}

func TestBidentIgnoresOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	// Bident under the OPPONENT; my creature connects with them.
	bidentID := effectstest.PushPermanentForTest(g, opp.ID, "Bident of Thassa", effectstest.BidentOracle, "Legendary Enchantment Artifact")
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == bidentID {
			t.Errorf("opponent's Bident prompted on MY creature's combat damage")
		}
	}
}

// --- Coastal Piracy -------------------------------------------

func TestCoastalPiracyDrawsOnConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	piracyID := effectstest.PushPermanentForTest(g, me.ID, "Coastal Piracy", effectstest.CoastalPiracyOrcl, "Enchantment")
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	handBefore := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, a)
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, piracyID) == nil {
		t.Fatalf("Coastal Piracy trigger not on the stack")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Coastal Piracy: hand delta %d, want 1", got)
	}
}

// --- Scroll Thief ---------------------------------------------

func TestScrollThiefSelfOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	thiefID := effectstest.PushDiesCreatureForTest(g, me.ID, "Scroll Thief", effectstest.ScrollThiefOracle, "Creature — Merfolk Rogue", 1, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// Only the bear attacks: no Thief trigger.
	effectstest.AttackWith(t, g, opp.ID, bear)
	if effectstest.TriggerOnStack(g, thiefID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatalf("Scroll Thief triggered on another creature's damage")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Next turn cycle: the Thief itself connects.
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Past the draw step before taking the baseline, so the turn
	// draw doesn't pollute the delta.
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	handBefore := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, thiefID)
	trig := effectstest.TriggerOnStack(g, thiefID)
	if trig == nil {
		t.Fatalf("Scroll Thief did not trigger on its own combat damage")
	}
	if trig.Controller != me.ID {
		t.Errorf("trigger controller %s, want %s", trig.Controller, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Scroll Thief: hand delta %d, want 1", got)
	}
}

// --- Edric, Spymaster of Trest --------------------------------

func TestEdricLetsAttackingCreaturesControllerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker := g.Seats[0]
	edricOwner := g.Seats[2]
	victim := g.Seats[1]
	edricID := effectstest.PushDiesCreatureForTest(g, edricOwner.ID, "Edric, Spymaster of Trest", effectstest.EdricOracle,
		"Legendary Creature — Elf Rogue", 2, 2)
	bear := effectstest.PushVanillaCreature(g, attacker.ID, "Bear", 2, 2)
	attackerHand := attacker.Hand.Size()
	ownerHand := edricOwner.Hand.Size()

	// Seat 0's bear hits seat 1 — one of seat 2's opponents.
	effectstest.AttackWith(t, g, victim.ID, bear)

	g.RunStateChecksForTest()
	trig := effectstest.TriggerOnStack(g, edricID)
	if trig == nil {
		t.Fatalf("Edric trigger not on the stack")
	}
	// The ability is Edric's controller's (it's their permanent) …
	if trig.Controller != edricOwner.ID {
		t.Errorf("trigger controller %s, want Edric's controller %s", trig.Controller, edricOwner.ID)
	}
	// … and its "you may" goes to the BEAR's controller, as it resolves.
	prompt := effectstest.AwaitTriggerPrompt(t, g, attacker.ID)
	if prompt.Source != edricID {
		t.Fatalf("the attacker's prompt is from %s, want Edric %s", prompt.Source, edricID)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, attacker.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// … but the card goes to the attacker.
	if got := attacker.Hand.Size() - attackerHand; got != 1 {
		t.Errorf("attacker hand delta %d, want 1", got)
	}
	if got := edricOwner.Hand.Size() - ownerHand; got != 0 {
		t.Errorf("Edric's controller hand delta %d, want 0", got)
	}
}

func TestEdricIgnoresDamageToItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker := g.Seats[0]
	edricOwner := g.Seats[1]
	edricID := effectstest.PushDiesCreatureForTest(g, edricOwner.ID, "Edric, Spymaster of Trest", effectstest.EdricOracle,
		"Legendary Creature — Elf Rogue", 2, 2)
	bear := effectstest.PushVanillaCreature(g, attacker.ID, "Bear", 2, 2)

	// Hitting Edric's own controller is not "one of your opponents".
	effectstest.AttackWith(t, g, edricOwner.ID, bear)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == edricID {
			t.Errorf("Edric prompted on damage dealt to its own controller")
		}
	}
}
