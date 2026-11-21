package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b17_cards_test.go — Mystery Booster 2 artifacts, GAP batch 17 (#5948).

const (
	gapMoxPoison   = "8d1a45b1-08c4-441d-86d6-f44061230c0b"
	gapPowerMatrix = "4355507e-e3a4-4c49-a1f7-344150213a2f"
	gapBlackVise   = "de7839fb-7040-48ab-a6d4-d1952972943d"
	gapErraticPort = "4a85cd7e-625b-4043-b7f1-999def26eb36"
	gapSmokestack  = "fda61506-ed79-47fd-badd-639185297913"
)

func TestS82GapB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapMoxPoison: "Mox Poison", gapPowerMatrix: "Power Matrix", gapBlackVise: "Black Vise",
		gapErraticPort: "Erratic Portal", gapSmokestack: "Smokestack",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB17MoxPoisonAddsAnyColorAndTwoPoison(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mox := b12Push(g, me.ID, "Mox Poison", "Artifact", gapMoxPoison, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap Mox Poison: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool = %+v, want one mana", me.ManaPool)
	}
	if got := me.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison counters = %d, want 2", got)
	}
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Mox activated again")
	}
	if got := me.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison counters after the refused tap = %d, want 2", got)
	}
}

func TestS82GapB17PowerMatrixPumpsAndGrants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pm := b12Push(g, me.ID, "Power Matrix", "Artifact", gapPowerMatrix, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, pm, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if p, to := effectivePTOf(t, g, bear); p != 3 || to != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, to)
	}
	for _, kw := range []string{"flying", "first strike", "trample"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("bear lacks %s", kw)
		}
	}
	if !b16Tapped(t, g, pm) {
		t.Error("Power Matrix taps as its cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, pm, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Error("a tapped Power Matrix activated again")
	}
	if err := g.ActivateCatalogAbility(me.ID, pm, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(me.ID)}); err == nil {
		t.Error("a player is not a creature")
	}
}

func TestS82GapB17BlackViseHurtsOnlyTheChosenPlayerByHandMinusFour(t *testing.T) {
	g := newCatalogGame(t)
	me, chosen, other := g.Seats[0], g.Seats[1], g.Seats[2]
	castCatalogSpell(t, g, "Black Vise", "Artifact", gapBlackVise, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, me.ID, chosen.ID)

	advanceToUpkeepOf(t, g, 2)
	d2dSetHand(other, 9)
	start := other.Life
	passPriorityAroundTable(t, g)
	if other.Life != start {
		t.Errorf("an unchosen opponent took %d", start-other.Life)
	}

	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(chosen, 7)
	start = chosen.Life
	passPriorityAroundTable(t, g)
	if chosen.Life != start-3 {
		t.Errorf("hand of 7: life %d -> %d, want -3 (7 minus 4)", start, chosen.Life)
	}

	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(chosen, 4)
	start = chosen.Life
	passPriorityAroundTable(t, g)
	if chosen.Life != start {
		t.Errorf("hand of 4 took %d, want 0", start-chosen.Life)
	}
}

func TestS82GapB17ErraticPortalBouncesUnlessThePayerPays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	portal := b12Push(g, me.ID, "Erratic Portal", "Artifact", gapErraticPort, 0, 0)
	bear := seedBear(g, opp.ID)
	other := seedBear(g, opp.ID)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID), Strict: true}); err == nil {
		t.Error("a player is not a creature")
	}
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}

	findBattlefieldCardForTest(g, portal).Tapped = false
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: cardRefs(other), Strict: true}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	passPriorityAroundTable(t, g)
	b06AddMana(opp, "C")
	answerPayUnless(t, g, opp.ID, true)
	if !g.Battlefield.Contains(other) {
		t.Error("paying {1}: the creature should stay on the battlefield")
	}
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: cardRefs(other), Strict: true}); err == nil {
		t.Error("a tapped Portal activated without mana")
	}
}

func gapSootOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters["soot"]
}

func TestS82GapB17SmokestackGrowsOnlyOnYourUpkeepAndEatsPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stack := b12Push(g, me.ID, "Smokestack", "Artifact", gapSmokestack, 0, 0)

	// Own upkeep: may add a counter; then the owner sacrifices one per counter.
	b12Push(g, me.ID, "Fodder", "Artifact", "", 0, 0)
	advanceToUpkeepOf(t, g, 0)
	fodder := b12Push(g, me.ID, "Second Fodder", "Artifact", "", 0, 0)
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if got := gapSootOn(t, g, stack); got != 1 {
		t.Fatalf("soot counters = %d, want 1", got)
	}

	// An opponent's upkeep: no counter, but they sacrifice one permanent.
	a := seedBear(g, opp.ID)
	b := seedBear(g, opp.ID)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Errorf("the opponent should have sacrificed exactly the chosen permanent (a=%v b=%v)", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
	if got := gapSootOn(t, g, stack); got != 1 {
		t.Errorf("soot counters = %d after an opponent's upkeep, want 1", got)
	}
}
