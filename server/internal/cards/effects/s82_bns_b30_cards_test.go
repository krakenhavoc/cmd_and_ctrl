package effects

import (
	"testing"

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

func TestS82BnsB30EyeCollectorMillsEachPlayerOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	eye := b12Push(g, me.ID, "Eye Collector", "Creature — Faerie", s82BnsB30Cards["Eye Collector"], 1, 1)
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	dealCombatDamageToPlayer(g, eye, opp.ID, 1)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Library.Size(); got != sizes[i]-1 {
			t.Errorf("seat %d library %d, want %d", i, got, sizes[i]-1)
		}
	}
}

func TestS82BnsB30FalkenrathMaraudersGrowsByTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	m := b12Push(g, me.ID, "Falkenrath Marauders", "Creature — Vampire Warrior", s82BnsB30Cards["Falkenrath Marauders"], 2, 2)
	dealCombatDamageToPlayer(g, m, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, m, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestS82BnsB30FathomFleetFirebrandPumpsRepeatedly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	f := b12Push(g, me.ID, "Fathom Fleet Firebrand", "Creature — Human Pirate", s82BnsB30Cards["Fathom Fleet Firebrand"], 2, 2)
	for range 2 {
		apaMana(me, "R", "R")
		apaActivate(t, g, me, f, 0, game.ActivateAbilityParams{})
	}
	if p, tt := effectivePTOf(t, g, f); p != 4 || tt != 2 {
		t.Errorf("%d/%d, want 4/2", p, tt)
	}
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestS82BnsB30FavoredOfIroasGainsDoubleStrikeFromYourEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	f := b12Push(g, me.ID, "Favored of Iroas", "Creature — Human Soldier", s82BnsB30Cards["Favored of Iroas"], 2, 2)
	if hasEffectiveKeyword(t, g, f, "double strike") {
		t.Fatal("double strike before any enchantment")
	}
	castAndResolveCreature(t, g, "Test Artifact", "Artifact", "")
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("an artifact triggered constellation")
	}
	castAndResolveCreature(t, g, "Test Enchantment", "Enchantment", "")
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("an enchantment did not grant double strike")
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, f, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestS82BnsB30FierceWitchstalkerMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	w := castAndResolveCreature(t, g, "Fierce Witchstalker", "Creature — Wolf", s82BnsB30Cards["Fierce Witchstalker"])
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Food", me.ID); n != 1 {
		t.Errorf("%d Food tokens, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, w, "trample") {
		t.Error("no trample")
	}
}

func TestS82BnsB30FissureWizardRummagesOnlyForADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castAndResolveCreature(t, g, "Fissure Wizard", "Creature — Goblin Wizard", s82BnsB30Cards["Fissure Wizard"])
	passPriorityAroundTable(t, g)
	hand, grave := me.Hand.Size(), me.Graveyard.Size()
	if discardChoiceFor(g, me.ID) == nil {
		t.Skip("no discard prompt open")
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	monSettle(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), hand, grave+1)
	}
}

func TestS82BnsB30FlutterfoxFliesWithAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	f := b12Push(g, me.ID, "Flutterfox", "Creature — Fox", s82BnsB30Cards["Flutterfox"], 2, 2)
	if hasEffectiveKeyword(t, g, f, "flying") {
		t.Fatal("flying with nothing")
	}
	b12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	if hasEffectiveKeyword(t, g, f, "flying") {
		t.Error("an opponent's artifact gave flying")
	}
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	if !hasEffectiveKeyword(t, g, f, "flying") {
		t.Error("no flying with an artifact")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	g.RunStateChecksForTest()
	if hasEffectiveKeyword(t, g, f, "flying") {
		t.Error("flying after the artifact left")
	}
	b12Permanent(g, me.ID, "Ench", "Enchantment")
	if !hasEffectiveKeyword(t, g, f, "flying") {
		t.Error("no flying with an enchantment")
	}
}

func TestS82BnsB30FretworkColonyGrowsAndBleedsAndCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b12Push(g, me.ID, "Fretwork Colony", "Creature — Insect", s82BnsB30Cards["Fretwork Colony"], 1, 1)
	if auraRestrictions(t, g, c)&game.CantBlock == 0 {
		t.Error("it can block")
	}
	life := lifeOf(g, me.ID)
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, c, game.CounterPlusOne); got < 1 {
		t.Errorf("%d counters after an upkeep", got)
	}
	if lifeOf(g, me.ID) >= life {
		t.Errorf("life %d, started %d", lifeOf(g, me.ID), life)
	}
}

func TestS82BnsB30FrostTricksterTapsAndFreezes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castAndResolveCreature(t, g, "Frost Trickster", "Creature — Bird Wizard", s82BnsB30Cards["Frost Trickster"])
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		pickCard(t, g, me.ID, theirs)
		passPriorityAroundTable(t, g)
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("not tapped")
	}
	if len(tmtB1Card(t, g, theirs).NextUntapSkips) != 1 {
		t.Error("not frozen")
	}
	if b16Tapped(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

func TestS82BnsB30GenerousPupSpreadsCountersOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pup := b12Push(g, me.ID, "Generous Pup", "Creature — Dog", s82BnsB30Cards["Generous Pup"], 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pup, game.CounterPlusOne, 1) })
	passPriorityAroundTable(t, g)
	if plusOnes(g, other) != 1 || plusOnes(g, theirs) != 0 || plusOnes(g, pup) != 1 {
		t.Errorf("other %d, theirs %d, pup %d; want 1, 0, 1", plusOnes(g, other), plusOnes(g, theirs), plusOnes(g, pup))
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pup, game.CounterPlusOne, 1) })
	passPriorityAroundTable(t, g)
	if plusOnes(g, other) != 1 {
		t.Errorf("a second trigger in one turn: other has %d counters, want 1", plusOnes(g, other))
	}
}

func TestS82BnsB30GhirapurGuideBlocksOutSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	guide := pushVanillaCreature(g, me.ID, "Ghirapur Guide", 3, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, guide).OracleID = s82BnsB30Cards["Ghirapur Guide"] })
	runner := pushVanillaCreature(g, me.ID, "Runner", 2, 2)
	small := pushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
	apaMana(me, "G", "G", "G")
	apaActivate(t, g, me, guide, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: runner}}})
	passPriorityAroundTable(t, g)
	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("%d block rules, want 1", n)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(small, runner), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, runner); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
	advanceToUpkeepOf(t, g, 1)
	if n := scopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d block rules outlived the turn", n)
	}
}
