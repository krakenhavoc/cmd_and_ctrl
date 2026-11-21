package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB17MoxPoisonAddsAnyColorAndTwoPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mox := effectstest.B12Push(g, me.ID, "Mox Poison", "Artifact", gapMoxPoison, 0, 0)
	effectstest.AdvanceToMain(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pm := effectstest.B12Push(g, me.ID, "Power Matrix", "Artifact", gapPowerMatrix, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, pm, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if p, to := effectstest.EffectivePTOf(t, g, bear); p != 3 || to != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, to)
	}
	for _, kw := range []string{"flying", "first strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("bear lacks %s", kw)
		}
	}
	if !effectstest.B16Tapped(t, g, pm) {
		t.Error("Power Matrix taps as its cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, pm, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Error("a tapped Power Matrix activated again")
	}
	if err := g.ActivateCatalogAbility(me.ID, pm, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(me.ID)}); err == nil {
		t.Error("a player is not a creature")
	}
}

func TestS82GapB17BlackViseHurtsOnlyTheChosenPlayerByHandMinusFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, chosen, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.CastCatalogSpell(t, g, "Black Vise", "Artifact", gapBlackVise, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, chosen.ID)

	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.D2dSetHand(other, 9)
	start := other.Life
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != start {
		t.Errorf("an unchosen opponent took %d", start-other.Life)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(chosen, 7)
	start = chosen.Life
	effectstest.PassPriorityAroundTable(t, g)
	if chosen.Life != start-3 {
		t.Errorf("hand of 7: life %d -> %d, want -3 (7 minus 4)", start, chosen.Life)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(chosen, 4)
	start = chosen.Life
	effectstest.PassPriorityAroundTable(t, g)
	if chosen.Life != start {
		t.Errorf("hand of 4 took %d, want 0", start-chosen.Life)
	}
}

func TestS82GapB17ErraticPortalBouncesUnlessThePayerPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	portal := effectstest.B12Push(g, me.ID, "Erratic Portal", "Artifact", gapErraticPort, 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)
	other := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID), Strict: true}); err == nil {
		t.Error("a player is not a creature")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}

	effectstest.FindBattlefieldCardForTest(g, portal).Tapped = false
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(opp, "C")
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	if !g.Battlefield.Contains(other) {
		t.Error("paying {1}: the creature should stay on the battlefield")
	}
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err == nil {
		t.Error("a tapped Portal activated without mana")
	}
}

func gapSootOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters["soot"]
}

func TestS82GapB17SmokestackGrowsOnlyOnYourUpkeepAndEatsPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stack := effectstest.B12Push(g, me.ID, "Smokestack", "Artifact", gapSmokestack, 0, 0)

	// Own upkeep: may add a counter; then the owner sacrifices one per counter.
	effectstest.B12Push(g, me.ID, "Fodder", "Artifact", "", 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	fodder := effectstest.B12Push(g, me.ID, "Second Fodder", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapSootOn(t, g, stack); got != 1 {
		t.Fatalf("soot counters = %d, want 1", got)
	}

	// An opponent's upkeep: no counter, but they sacrifice one permanent.
	a := effectstest.SeedBear(g, opp.ID)
	b := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Errorf("the opponent should have sacrificed exactly the chosen permanent (a=%v b=%v)", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
	if got := gapSootOn(t, g, stack); got != 1 {
		t.Errorf("soot counters = %d after an opponent's upkeep, want 1", got)
	}
}
