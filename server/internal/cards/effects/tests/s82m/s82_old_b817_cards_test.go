package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b817_cards_test.go — Alliances, instants and sorceries, batch 817
// (#5634, S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB817Misinformation = "03dc8820-ec10-42c1-8200-d323cfcbf2b0"
	oldB817Omen           = "5935e9b3-b9b9-413c-ba17-103f5792131a"
	oldB817Bounty         = "77a43c24-398f-42c6-a0d3-b28b33a0ce12"
	oldB817Errand         = "977f0d8d-4359-44ae-8f22-2103ea909c40"
	oldB817Ritual         = "bccd8298-e056-4b91-b090-408c75467b92"
	oldB817Surge          = "bd3e55cf-ffca-402c-91ea-2420035a9dd1"
)

func TestOldB817CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB817Misinformation: "Misinformation", oldB817Omen: "Omen of Fire", oldB817Bounty: "Bounty of the Hunt",
		oldB817Errand: "Errand of Duty", oldB817Ritual: "Ritual of the Machine", oldB817Surge: "Surge of Strength",
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

func TestOldB817MisinformationPutsThreeCardsOnTopOfTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.GraveyardCardOf(g, opp, "A", "Sorcery", "", "")
	b := effectstest.GraveyardCardOf(g, opp, "B", "Sorcery", "", "")
	c := effectstest.GraveyardCardOf(g, opp, "C", "Sorcery", "", "")
	d := effectstest.GraveyardCardOf(g, opp, "D", "Sorcery", "", "")
	mine := effectstest.GraveyardCardOf(g, me, "Mine", "Sorcery", "", "")
	spell := effectstest.HandCardFull(me, "Misinformation", "Instant", "{B}", oldB817Misinformation, []string{"B"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b, c, d)}); err == nil {
		t.Fatal("four cards were accepted")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Fatal("a card from the caster's own graveyard was accepted")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(a, b, c)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatalf("the caster was not asked for the order: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{c, a, b}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if opp.Graveyard.Contains(a) || opp.Graveyard.Contains(b) || opp.Graveyard.Contains(c) || !opp.Graveyard.Contains(d) {
		t.Error("exactly the three targets should have left the graveyard")
	}
	n := len(opp.Library.Cards)
	if opp.Library.Cards[n-1].InstanceID != c {
		t.Error("the first card named is not on top of their library")
	}
	if !me.Graveyard.Contains(mine) || len(me.Library.Cards) != len(g.Seats[0].Library.Cards) {
		t.Error("the caster's own zones changed")
	}
}

func TestOldB817ErrandOfDutyMakesAKnightWithBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Errand of Duty", "Instant", oldB817Errand, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var knight *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Knight" && g.Battlefield.Cards[i].Controller == me.ID {
			knight = &g.Battlefield.Cards[i]
		}
	}
	if knight == nil {
		t.Fatal("no Knight token")
	}
	if !knight.IsToken() || !knight.HasColor("W") || !effectstest.HasKeywordOnBattlefield(t, g, knight.InstanceID, "banding") {
		t.Errorf("the token is wrong: %+v", *knight)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, knight.InstanceID); p != 1 || tough != 1 {
		t.Errorf("the Knight is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB817RitualOfTheMachineStealsForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	black := effectstest.B12Creature(g, opp.ID, "Black", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, black).Colors = []string{"B"} })
	artifact := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	spell := effectstest.HandCardFull(me, "Ritual of the Machine", "Sorcery", "{2}{B}{B}", oldB817Ritual, []string{"B"})
	for _, bad := range []uuid.UUID{black, artifact} {
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bad), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
			t.Fatal("an artifact or black creature was a legal target")
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, victim) != me.ID {
		t.Error("the caster does not control the creature")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, victim) != me.ID {
		t.Error("control ended with the turn; it has no duration")
	}
}

func TestOldB817SurgeOfStrengthDiscardsARedOrGreenCardAndPumpsByManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Four-drop", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).ManaCost = "{2}{G}{G}" })
	blue := effectstest.HandCardFull(me, "Blue", "Instant", "{U}", "", []string{"U"})
	green := effectstest.HandCardFull(me, "Green", "Instant", "{G}", "", []string{"G"})
	spell := effectstest.HandCardFull(me, "Surge of Strength", "Instant", "{R}{G}", oldB817Surge, []string{"R", "G"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear), DiscardIDs: []uuid.UUID{blue}}); err == nil {
		t.Fatal("a blue card paid the discard")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("cast without discarding")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardTarget(bear), DiscardIDs: []uuid.UUID{green}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(green) {
		t.Error("the green card was not discarded")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 6/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestOldB817OmenOfFireBouncesIslandsThenEachPlayerSacrificesPerWhitePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	plains := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	whiteA := effectstest.B12Creature(g, opp.ID, "Knight A", "Creature — Knight", 2, 2)
	whiteB := effectstest.B12Creature(g, opp.ID, "Knight B", "Creature — Knight", 2, 2)
	green := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{whiteA, whiteB} {
			effectstest.FindBattlefieldCardForTest(g, id).Colors = []string{"W"}
		}
		effectstest.FindBattlefieldCardForTest(g, green).Colors = []string{"G"}
	})
	effectstest.CastCatalogSpell(t, g, "Omen of Fire", "Instant", oldB817Omen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(island) || !opp.Hand.Contains(island) {
		t.Error("the Island was not returned to its owner's hand")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, whiteA, whiteB)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(whiteA) || g.Battlefield.Contains(whiteB) || !g.Battlefield.Contains(green) {
		t.Error("the opponent owed exactly the two white permanents they control")
	}
	if !g.Battlefield.Contains(plains) {
		t.Error("the caster controls no white permanent, so owes no sacrifice")
	}
}

func TestOldB817BountyOfTheHuntPutsCountersThenTakesThemBackAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	spell := effectstest.HandCardFull(me, "Bounty of the Hunt", "Instant", "{3}{G}{G}", oldB817Bounty, []string{"G"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1}}); err == nil {
		t.Fatal("cast without paying and without pitching")
	}
	pitch := effectstest.HandCardFull(me, "Pitch", "Instant", "{G}", "", []string{"G"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{effectstest.HandCardFull(me, "Blue", "Instant", "{U}", "", []string{"U"})}}); err == nil {
		t.Fatal("a blue card was pitched")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("pitching a green card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, game.CounterPlusOne) != 2 || effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatalf("counters %d and %d, want 2 and 1", effectstest.CountersOn(g, a, game.CounterPlusOne), effectstest.CountersOn(g, b, game.CounterPlusOne))
	}
	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	if n := effectstest.CountersOn(g, a, game.CounterPlusOne) + effectstest.CountersOn(g, b, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters by the end step, want all 3", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	// Two delayed triggers fire together at cleanup: the controller orders them.
	for i := 0; i < 4; i++ {
		_, _ = g.AdvanceStep()
		effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	}
	if n := effectstest.CountersOn(g, a, game.CounterPlusOne) + effectstest.CountersOn(g, b, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters are left after the cleanup step", n)
	}
}

func TestOldB817OmenOfFireSacrificesEvenWithNoIslandToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plains := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	white := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Knight", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, white).Colors = []string{"W"} })
	effectstest.CastCatalogSpell(t, g, "Omen of Fire", "Instant", oldB817Omen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plains) || !g.Battlefield.Contains(white) {
		t.Errorf("one white permanent means one sacrifice, here the Plains (Plains %v, Knight %v)", g.Battlefield.Contains(plains), g.Battlefield.Contains(white))
	}
}
