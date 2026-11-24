package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b056_test.go — Commander 2021, batch 56 (#5504): +1/+1 and other
// counters. Each test plays the real catalog entry and asserts what its
// text says.

const (
	oldB056SunDroplet     = "1820af5c-9cc2-4b77-b4ca-86084442f087"
	oldB056InfernoProject = "36d641d4-d3f5-4394-81d7-3209955a1628"
	oldB056BoldPlagiarist = "85277284-1eba-4f87-bf7b-abf1dcbb7691"
	oldB056SequenceEngine = "87c1bd51-6a85-4a14-af7b-a5b02ab69d73"
	oldB056GeometricNexus = "935bcf49-ec8b-4ad6-b12d-77dc7cccf1d4"
	oldB056Willowdusk     = "94723d61-3e29-4180-91bc-df37c4086a66"
	oldB056StartingLife   = 40
)

func oldB056Seats(g *game.Game) (*game.Player, *game.Player) {
	i := g.Turn.ActiveSeat
	return g.Seats[i], g.Seats[(i+1)%len(g.Seats)]
}

func oldB056Fractals(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == "Fractal" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// oldB056AdvanceToDropletPrompt walks the turn forward, resolving what is on
// the stack, until Sun Droplet asks its controller (it triggers at EVERY
// player's upkeep).
func oldB056AdvanceToDropletPrompt(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 300; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestConfirmFor(g, me.ID) != nil {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("Sun Droplet never asked")
}

func TestOldB056EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB056SunDroplet:     "Sun Droplet",
		oldB056InfernoProject: "Inferno Project",
		oldB056BoldPlagiarist: "Bold Plagiarist",
		oldB056SequenceEngine: "Sequence Engine",
		oldB056GeometricNexus: "Geometric Nexus",
		oldB056Willowdusk:     "Willowdusk, Essence Seer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", name, spec.Completeness)
		}
	}
}

// --- Sun Droplet ------------------------------------------------------------

func TestOldB056SunDropletCountsDamageButNotLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB056Seats(g)
	droplet := effectstest.PushCatalogPermanent(g, me.ID, "Sun Droplet", "Artifact", oldB056SunDroplet, false)
	hitter := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.LoseLife(t, g, me, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, droplet, "charge"); n != 0 {
		t.Fatalf("%d charge counters after plain life loss, want 0", n)
	}
	effectstest.DealCombatDamageToPlayer(g, hitter, me.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, droplet, "charge"); n != 4 {
		t.Errorf("%d charge counters after 4 damage, want 4", n)
	}
	// Damage to another player does not count.
	effectstest.DealCombatDamageToPlayer(g, hitter, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, droplet, "charge"); n != 4 {
		t.Errorf("%d charge counters after damage to someone else, want still 4", n)
	}
}

func TestOldB056SunDropletUpkeepRemovesACounterForOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	droplet := effectstest.PushCatalogPermanent(g, me.ID, "Sun Droplet", "Artifact", oldB056SunDroplet, false)
	effectstest.AddCountersForTest(g, droplet, "charge", 2)
	oldB056AdvanceToDropletPrompt(t, g, me)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, droplet, "charge"); n != 2 || me.Life != oldB056StartingLife {
		t.Fatalf("declining changed things: %d counters, life %d", n, me.Life)
	}
}

func TestOldB056SunDropletUpkeepYesRemovesOneAndGainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	droplet := effectstest.PushCatalogPermanent(g, me.ID, "Sun Droplet", "Artifact", oldB056SunDroplet, false)
	effectstest.AddCountersForTest(g, droplet, "charge", 2)
	oldB056AdvanceToDropletPrompt(t, g, me)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, droplet, "charge"); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	if me.Life != oldB056StartingLife+1 {
		t.Errorf("life %d, want %d", me.Life, oldB056StartingLife+1)
	}
}

func TestOldB056SunDropletAsksNothingWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sun Droplet", "Artifact", oldB056SunDroplet, false)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to remove a counter that is not there")
	}
}

// --- Inferno Project --------------------------------------------------------

func TestOldB056InfernoProjectEntersWithTheTotalManaValueOfInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB056Seats(g)
	effectstest.B17GraveyardCard(me, "Shock", "Instant", "{R}")
	effectstest.B17GraveyardCard(me, "Fireball", "Sorcery", "{X}{R}")
	effectstest.B17GraveyardCard(me, "Rolling Thunder", "Sorcery", "{3}{R}{R}")
	effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.B17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	effectstest.AdvanceToMain(t, g)
	project := effectstest.CastFromHandForTest(t, g, me, "Inferno Project", "Creature — Elemental", "{6}{R}", oldB056InfernoProject, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// 1 + 1 (X is 0 in a graveyard) + 5 = 7.
	if n := effectstest.CountersOn(g, project, game.CounterPlusOne); n != 7 {
		t.Errorf("%d +1/+1 counters, want 7", n)
	}
	if p, tough := effectstest.PtOf(t, g, project); p != 7 || tough != 7 {
		t.Errorf("Inferno Project is %d/%d, want 7/7", p, tough)
	}
}

func TestOldB056InfernoProjectDiesWithNoInstantsOrSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Inferno Project", TypeLine: "Creature — Elemental", ManaCost: "{6}{R}",
		OracleID: oldB056InfernoProject, PrintedPTKnown: true, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) {
		t.Error("a 0/0 Inferno Project stayed on the battlefield")
	}
}

// --- Bold Plagiarist --------------------------------------------------------

func TestOldB056BoldPlagiaristCopiesCountersAnOpponentPutsOnTheirOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB056Seats(g)
	plagiarist := effectstest.B12Push(g, me.ID, "Bold Plagiarist", "Creature — Vampire Rogue", oldB056BoldPlagiarist, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, theirs, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, plagiarist, game.CounterPlusOne); n != 2 {
		t.Errorf("the Plagiarist has %d +1/+1 counters, want 2", n)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, theirs, "shield", 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, plagiarist, "shield"); n != 1 {
		t.Errorf("the Plagiarist has %d shield counters, want 1", n)
	}
}

func TestOldB056BoldPlagiaristIgnoresCountersOnCreaturesTheOpponentDoesNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB056Seats(g)
	plagiarist := effectstest.B12Push(g, me.ID, "Bold Plagiarist", "Creature — Vampire Rogue", oldB056BoldPlagiarist, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// I put counters on my own creature: not an opponent's placement.
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, mine, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	// An opponent puts counters on MY creature, and on the Plagiarist: not
	// "a creature they control".
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, mine, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.AddCounterByForEffect(opp.ID, plagiarist, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, plagiarist, game.CounterPlusOne); n != 1 {
		t.Errorf("the Plagiarist has %d +1/+1 counters, want only the 1 an opponent put on it directly", n)
	}
}

// --- Sequence Engine --------------------------------------------------------

func TestOldB056SequenceEngineExilesAMatchingCreatureCardForAFractalOfThatSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB056Seats(g)
	engine := effectstest.PushCatalogPermanent(g, me.ID, "Sequence Engine", "Artifact", oldB056SequenceEngine, false)
	three := effectstest.B17GraveyardCard(opp, "Ogre", "Creature — Ogre", "{2}{R}")
	two := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardTargets(two)}); err == nil {
		t.Fatal("targeted a card whose mana value is not X")
	}
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardTargets(three)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(three) || opp.Graveyard.Contains(three) {
		t.Error("the creature card was not exiled from the opponent's graveyard")
	}
	fractals := oldB056Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if p, tough := effectstest.PtOf(t, g, fractals[0]); p != 3 || tough != 3 {
		t.Errorf("the Fractal is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.TappedNow(g, engine) {
		t.Error("the Engine did not tap")
	}
}

func TestOldB056SequenceEngineRefusesANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	engine := effectstest.PushCatalogPermanent(g, me.ID, "Sequence Engine", "Artifact", oldB056SequenceEngine, false)
	spell := effectstest.B17GraveyardCard(me, "Shock", "Instant", "{1}{R}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardTargets(spell)}); err == nil {
		t.Fatal("targeted an instant card")
	}
}

// --- Geometric Nexus --------------------------------------------------------

func TestOldB056GeometricNexusCountsSpellManaValuesThenMakesAFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	nexus := effectstest.PushCatalogPermanent(g, me.ID, "Geometric Nexus", "Artifact", oldB056GeometricNexus, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Test Spell", "Instant", "{2}{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Test Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, nexus, "charge"); n != 3 {
		t.Fatalf("%d charge counters, want 3 (the creature spell does not count)", n)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nexus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, nexus, "charge"); n != 0 {
		t.Errorf("%d charge counters left, want 0", n)
	}
	fractals := oldB056Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if p, tough := effectstest.PtOf(t, g, fractals[0]); p != 3 || tough != 3 {
		t.Errorf("the Fractal is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB056GeometricNexusWithNoCountersMakesADyingFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	nexus := effectstest.PushCatalogPermanent(g, me.ID, "Geometric Nexus", "Artifact", oldB056GeometricNexus, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nexus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB056Fractals(g, me.ID)); got != 0 {
		t.Errorf("%d Fractals survived with X = 0, want a 0/0 that dies", got)
	}
}

// --- Willowdusk, Essence Seer -----------------------------------------------

func oldB056Willow(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB056Seats(g)
	willow := effectstest.B12Push(g, me.ID, "Willowdusk, Essence Seer", "Legendary Creature — Dryad Druid", oldB056Willowdusk, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	return g, me, willow, bear
}

func TestOldB056WillowdyskUsesTheGreaterOfLifeGainedAndLifeLost(t *testing.T) {
	g, me, willow, bear := oldB056Willow(t)
	effectstest.GainLifeNow(t, g, willow, me, 5)
	effectstest.LoseLife(t, g, me, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(willow)}); err == nil {
		t.Fatal("Willowdusk targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 5 {
		t.Errorf("%d counters, want 5 (gained 5, lost 2)", n)
	}
}

func TestOldB056WillowduskLifeLostCanBeTheGreater(t *testing.T) {
	g, me, willow, bear := oldB056Willow(t)
	effectstest.GainLifeNow(t, g, willow, me, 1)
	effectstest.LoseLife(t, g, me, 7)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 7 {
		t.Errorf("%d counters, want 7 (gained 1, lost 7)", n)
	}
}

func TestOldB056WillowduskPutsNothingOnAQuietTurn(t *testing.T) {
	g, me, willow, bear := oldB056Willow(t)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters on a turn with no life change, want 0", n)
	}
}
