package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_triggers_test.go — S19 sub-PR 6: cast / opponent-draws
// triggers and the pay-unless prompt. Seat 0 is the active player
// in newCatalogGame; the "opponent" cards live under seat 1 so a
// seat-0 cast / draw is an opponent's from their point of view.

const (
	rhysticStudyOracle      = "53236dd7-845a-444c-96d5-f41ed7325d8f"
	smotheringTitheOracle   = "153376c9-dffd-458c-8ce3-a4c8269bc4e9"
	esperSentinelOracle     = "5def9f38-0a0b-4e8d-9f9d-29dcb46520b4"
	beastWhispererOracle    = "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a"
	consecratedSphinxOracle = "311a449d-dc74-46e6-9a47-6a597931f736"
	lightningBoltOracle     = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// answerPayUnless finds the pending pay_unless prompt addressed to
// chooserID and answers it (true = pay).
func answerPayUnless(t *testing.T, g *game.Game, chooserID uuid.UUID, pay bool) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != game.PendingChoicePayUnless || c.Chooser != chooserID {
			continue
		}
		snapshot := *c
		if err := g.ResolvePayUnless(c.ID, chooserID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		return &snapshot
	}
	t.Fatalf("no pay_unless prompt addressed to %s", chooserID)
	return nil
}

func hasPayUnlessFor(g *game.Game, chooserID uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooserID {
			return true
		}
	}
	return false
}

func countOnBattlefield(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- Rhystic Study --------------------------------------------

func TestRhysticStudyOpponentCastDeclineDraws(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	studyID := pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})

	// The Study's trigger sits above the Bolt (cast in response to
	// the cast event, higher Seq).
	item := triggerOnStack(g, studyID)
	if item == nil {
		t.Fatalf("Rhystic Study trigger not on the stack after an opponent cast")
	}
	if item.Controller != owner.ID {
		t.Errorf("trigger controller = %s, want the Study's controller %s", item.Controller, owner.ID)
	}
	passPriorityAroundTable(t, g)

	// Resolution asks the CASTER to pay, not the Study's owner.
	if !hasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster after the trigger resolved")
	}
	if owner.Hand.Size() != handBefore {
		t.Fatalf("Study drew before the payer answered")
	}
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("PayCost = %q, want {1}", prompt.PayCost)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Rhystic Study on decline: owner hand delta %d, want 1", got)
	}
}

func TestRhysticStudyOpponentPaysNoDraw(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	// Float the {1} now — castCatalogSpell walks steps to reach the
	// main phase, and pools empty on every step change.
	caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	answerPayUnless(t, g, caster.ID, true)

	if owner.Hand.Size() != handBefore {
		t.Errorf("Study drew even though the caster paid: delta %d", owner.Hand.Size()-handBefore)
	}
	if len(caster.ManaPool) != 0 {
		t.Errorf("payment did not leave the caster's pool: %d tokens left", len(caster.ManaPool))
	}
}

func TestRhysticStudyIgnoresControllersOwnCasts(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	studyID := pushPermanentForTest(g, caster.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if triggerOnStack(g, studyID) != nil {
		t.Errorf("Rhystic Study triggered on its own controller's cast")
	}
}

// --- Smothering Tithe -----------------------------------------

func TestSmotheringTitheOpponentDrawDeclineMakesTreasure(t *testing.T) {
	g := newCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	titheID := pushPermanentForTest(g, owner.ID, "Smothering Tithe", smotheringTitheOracle, "Enchantment")

	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	advanceToMain(t, g)

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	// The draw is a sandbox action outside any resolution — the
	// trigger has to land on the stack at that boundary.
	if triggerOnStack(g, titheID) == nil {
		t.Fatalf("Tithe trigger not on the stack after an opponent drew")
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer")
	}
	prompt := answerPayUnless(t, g, drawer.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("PayCost = %q, want {2}", prompt.PayCost)
	}
	if n := countOnBattlefield(g, "Treasure", owner.ID); n != 1 {
		t.Errorf("Treasures under the Tithe's controller: %d, want 1", n)
	}
}

func TestSmotheringTitheOpponentPaysNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Smothering Tithe", smotheringTitheOracle, "Enchantment")
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	// The mana goes in AFTER the walk — the pool empties at the end
	// of every step (CR 500.4).
	advanceToMain(t, g)
	drawer.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, drawer.ID, true)
	if n := countOnBattlefield(g, "Treasure", owner.ID); n != 0 {
		t.Errorf("Treasure created despite payment: %d", n)
	}
}

func TestSmotheringTitheIgnoresOwnDraws(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0]
	titheID := pushPermanentForTest(g, owner.ID, "Smothering Tithe", smotheringTitheOracle, "Enchantment")
	if err := g.DrawCard(owner.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if triggerOnStack(g, titheID) != nil || len(g.PendingTriggers) != 0 {
		t.Errorf("Tithe triggered on its own controller's draw")
	}
}

// --- Esper Sentinel -------------------------------------------

func TestEsperSentinelFirstNoncreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	sentinelID := pushDiesCreatureForTest(g, owner.ID, "Esper Sentinel", esperSentinelOracle,
		"Artifact Creature — Human Soldier", 1, 1)
	handBefore := owner.Hand.Size()

	// A creature spell first: not a noncreature spell, no trigger.
	castCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", colossalDreadmawOracle, nil)
	if triggerOnStack(g, sentinelID) != nil {
		t.Fatalf("Sentinel triggered on a creature spell")
	}
	passPriorityAroundTable(t, g)

	// First noncreature spell this turn: trigger.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if triggerOnStack(g, sentinelID) == nil {
		t.Fatalf("Sentinel did not trigger on the first noncreature spell")
	}
	passPriorityAroundTable(t, g)
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("PayCost = %q, want {1} (Sentinel's power)", prompt.PayCost)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Sentinel decline: owner hand delta %d, want 1", got)
	}

	// Second noncreature spell the same turn: no trigger.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if triggerOnStack(g, sentinelID) != nil {
		t.Errorf("Sentinel triggered on the second noncreature spell of the turn")
	}
}

func TestEsperSentinelXReadsPumpedPower(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	sentinelID := pushDiesCreatureForTest(g, owner.ID, "Esper Sentinel", esperSentinelOracle,
		"Artifact Creature — Human Soldier", 1, 1)
	// Two +1/+1 counters → power 3 → the ask is {3}.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(sentinelID, "+1/+1", 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{3}" {
		t.Errorf("PayCost = %q, want {3}", prompt.PayCost)
	}
}

// --- Beast Whisperer ------------------------------------------

func TestBeastWhispererDrawsOnOwnCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	whispererID := pushDiesCreatureForTest(g, caster.ID, "Beast Whisperer", beastWhispererOracle,
		"Creature — Elf Druid", 2, 3)

	// Noncreature: nothing.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if triggerOnStack(g, whispererID) != nil {
		t.Fatalf("Beast Whisperer triggered on an instant")
	}
	passPriorityAroundTable(t, g)
	handBefore := caster.Hand.Size()

	dreadmawID := castCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", colossalDreadmawOracle, nil)
	trig := triggerOnStack(g, whispererID)
	if trig == nil {
		t.Fatalf("Beast Whisperer did not trigger on a creature spell")
	}
	spell := g.StackMeta[dreadmawID]
	if spell == nil || !(trig.Seq > spell.Seq) {
		t.Errorf("trigger must sit above the creature spell (resolves first)")
	}
	passPriorityAroundTable(t, g)
	if got := caster.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Beast Whisperer: hand delta %d, want 1", got)
	}
	if !g.Battlefield.Contains(dreadmawID) {
		t.Errorf("the creature spell itself did not resolve")
	}
}

// --- Consecrated Sphinx ---------------------------------------

func TestConsecratedSphinxOpponentDrawYesDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	sphinxID := pushDiesCreatureForTest(g, owner.ID, "Consecrated Sphinx", consecratedSphinxOracle,
		"Creature — Sphinx", 4, 6)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	advanceToMain(t, g)
	handBefore := owner.Hand.Size()

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	if triggerOnStack(g, sphinxID) == nil {
		t.Fatalf("Sphinx trigger not on the stack after Yes")
	}
	passPriorityAroundTable(t, g)
	if got := owner.Hand.Size() - handBefore; got != 2 {
		t.Errorf("Sphinx: owner hand delta %d, want 2", got)
	}
	// The Sphinx's own draws must not re-trigger it.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Errorf("Sphinx re-triggered on its controller's own draws")
		}
	}
}

func TestConsecratedSphinxDeclineDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	pushDiesCreatureForTest(g, owner.ID, "Consecrated Sphinx", consecratedSphinxOracle,
		"Creature — Sphinx", 4, 6)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	advanceToMain(t, g)
	handBefore := owner.Hand.Size()

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	answerLatestTriggerPrompt(t, g, owner.ID, false)
	passPriorityAroundTable(t, g)
	if owner.Hand.Size() != handBefore {
		t.Errorf("Sphinx drew after a decline")
	}
}
