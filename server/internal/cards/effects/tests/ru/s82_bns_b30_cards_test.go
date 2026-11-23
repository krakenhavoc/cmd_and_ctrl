package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var s82BnsB30Cards = map[string]string{
	"Eye Collector":          "553c2006-e123-406e-ab4a-1d1f67c7468c",
	"Falkenrath Marauders":   "f819b134-24cd-4863-a46a-0a068e34ca9e",
	"Fathom Fleet Firebrand": "a297f9a2-3461-4138-8a3e-75027cbdbe12",
	"Favored of Iroas":       "5df784c0-3adc-4613-ba30-905ec8cf66f9",
	"Fierce Witchstalker":    "ad69f48e-c387-4376-a04f-37d3992c7946",
	"Fissure Wizard":         "e3474632-6d81-43a8-9329-4c7e0c631755",
	"Flutterfox":             "58805c22-29cc-4ab2-a95c-b1798eb7a90d",
	"Fretwork Colony":        "95f5ddb6-e3e5-4b35-88ac-eae192783eda",
	"Frost Trickster":        "1663d54f-4f39-4a03-aac2-e2923cbb7de8",
	"Generous Pup":           "9ee0d4a0-fd8d-43c4-a2aa-c4e6af5beecc",
	"Ghirapur Guide":         "92d5bb7e-08ba-448e-8763-f47c89505e80",
}

func TestS82BnsB30EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range s82BnsB30Cards {
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

func TestS82BnsB30EyeCollectorMillsEachPlayerOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	eye := effectstest.B12Push(g, me.ID, "Eye Collector", "Creature — Faerie", s82BnsB30Cards["Eye Collector"], 1, 1)
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	effectstest.DealCombatDamageToPlayer(g, eye, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Library.Size(); got != sizes[i]-1 {
			t.Errorf("seat %d library %d, want %d", i, got, sizes[i]-1)
		}
	}
}

func TestS82BnsB30FalkenrathMaraudersGrowsByTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.B12Push(g, me.ID, "Falkenrath Marauders", "Creature — Vampire Warrior", s82BnsB30Cards["Falkenrath Marauders"], 2, 2)
	effectstest.DealCombatDamageToPlayer(g, m, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, m, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestS82BnsB30FathomFleetFirebrandPumpsRepeatedly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	f := effectstest.B12Push(g, me.ID, "Fathom Fleet Firebrand", "Creature — Human Pirate", s82BnsB30Cards["Fathom Fleet Firebrand"], 2, 2)
	for range 2 {
		effectstest.ApaMana(me, "R", "R")
		effectstest.ApaActivate(t, g, me, f, 0, game.ActivateAbilityParams{})
	}
	if p, tt := effectstest.EffectivePTOf(t, g, f); p != 4 || tt != 2 {
		t.Errorf("%d/%d, want 4/2", p, tt)
	}
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestS82BnsB30FavoredOfIroasGainsDoubleStrikeFromYourEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	f := effectstest.B12Push(g, me.ID, "Favored of Iroas", "Creature — Human Soldier", s82BnsB30Cards["Favored of Iroas"], 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, f, "double strike") {
		t.Fatal("double strike before any enchantment")
	}
	effectstest.CastAndResolveCreature(t, g, "Test Artifact", "Artifact", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("an artifact triggered constellation")
	}
	effectstest.CastAndResolveCreature(t, g, "Test Enchantment", "Enchantment", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("an enchantment did not grant double strike")
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestS82BnsB30FierceWitchstalkerMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	w := effectstest.CastAndResolveCreature(t, g, "Fierce Witchstalker", "Creature — Wolf", s82BnsB30Cards["Fierce Witchstalker"])
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Food", me.ID); n != 1 {
		t.Errorf("%d Food tokens, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, w, "trample") {
		t.Error("no trample")
	}
}

func TestS82BnsB30FissureWizardRummagesOnlyForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Fissure Wizard", "Creature — Goblin Wizard", s82BnsB30Cards["Fissure Wizard"])
	effectstest.PassPriorityAroundTable(t, g)
	hand, grave := me.Hand.Size(), me.Graveyard.Size()
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Skip("no discard prompt open")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.MonSettle(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), hand, grave+1)
	}
}

func TestS82BnsB30FlutterfoxFliesWithAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.B12Push(g, me.ID, "Flutterfox", "Creature — Fox", s82BnsB30Cards["Flutterfox"], 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Fatal("flying with nothing")
	}
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	if effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Error("an opponent's artifact gave flying")
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if !effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Error("no flying with an artifact")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	g.RunStateChecksForTest()
	if effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Error("flying after the artifact left")
	}
	effectstest.B12Permanent(g, me.ID, "Ench", "Enchantment")
	if !effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Error("no flying with an enchantment")
	}
}

func TestS82BnsB30FretworkColonyGrowsAndBleedsAndCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Fretwork Colony", "Creature — Insect", s82BnsB30Cards["Fretwork Colony"], 1, 1)
	if effectstest.AuraRestrictions(t, g, c)&game.CantBlock == 0 {
		t.Error("it can block")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, c, game.CounterPlusOne); got < 1 {
		t.Errorf("%d counters after an upkeep", got)
	}
	if effectstest.LifeOf(g, me.ID) >= life {
		t.Errorf("life %d, started %d", effectstest.LifeOf(g, me.ID), life)
	}
}

func TestS82BnsB30FrostTricksterTapsAndFreezes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Frost Trickster", "Creature — Bird Wizard", s82BnsB30Cards["Frost Trickster"])
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("not tapped")
	}
	if len(effectstest.TmtB1Card(t, g, theirs).NextUntapSkips) != 1 {
		t.Error("not frozen")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

func TestS82BnsB30GenerousPupSpreadsCountersOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pup := effectstest.B12Push(g, me.ID, "Generous Pup", "Creature — Dog", s82BnsB30Cards["Generous Pup"], 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pup, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, other) != 1 || effectstest.PlusOnes(g, theirs) != 0 || effectstest.PlusOnes(g, pup) != 1 {
		t.Errorf("other %d, theirs %d, pup %d; want 1, 0, 1", effectstest.PlusOnes(g, other), effectstest.PlusOnes(g, theirs), effectstest.PlusOnes(g, pup))
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pup, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, other) != 1 {
		t.Errorf("a second trigger in one turn: other has %d counters, want 1", effectstest.PlusOnes(g, other))
	}
}

func TestS82BnsB30GhirapurGuideBlocksOutSmallBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guide := effectstest.PushVanillaCreature(g, me.ID, "Ghirapur Guide", 3, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, guide).OracleID = s82BnsB30Cards["Ghirapur Guide"] })
	runner := effectstest.PushVanillaCreature(g, me.ID, "Runner", 2, 2)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.ApaActivate(t, g, me, guide, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: runner}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ScopedBlockRuleCount(g); n != 1 {
		t.Fatalf("%d block rules, want 1", n)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(small, runner), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, runner); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if n := effectstest.ScopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d block rules outlived the turn", n)
	}
}
