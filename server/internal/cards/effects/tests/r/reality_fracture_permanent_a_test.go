package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Fracture slice fra-permanent-a: the artifact, enchantment and
// aura cards.

const (
	rfPermAAjanisAnguishOracle    = "4a37f4a0-2869-4f81-b51b-4c3e7e0e0671"
	rfPermAFerocityOracle         = "747a1b36-eb4f-44e6-9e5a-bb1a0259a471"
	rfPermAGardenizeOracle        = "e74a2a4c-ca3c-49e8-9b0f-b4d1bb91fb54"
	rfPermAHuntersAxeOracle       = "d7d59fef-1401-464b-b1bb-ee5da92cde51"
	rfPermAKitesailOracle         = "eacfe895-5106-437c-b8c1-2ff03c107744"
	rfPermAMemoryTrapOracle       = "70616a8c-9b6c-408e-8f3e-2c348ec136d8"
	rfPermAMurmuringVolumeOracle  = "4baa7847-dfcc-4aa3-aa64-2deeaa3f013b"
	rfPermAPuppetCraftingOracle   = "d13c111a-ebfc-48ac-b734-f203c7c29d21"
	rfPermASolitaryCellOracle     = "66a22b35-3dc0-4a74-8474-e73293403de6"
	rfPermAEchoverseFulcrumOracle = "c04197ad-d2f3-499f-b863-bed0c869519c"
)

func rfPermAEquipment(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return effectstest.B12Push(g, owner, name, "Artifact — Equipment", oracle, 0, 0)
}

func TestMedicsKitesailPumpsGrantsFlyingAndGainsLifeOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	sail := rfPermAEquipment(g, me.ID, "Medic's Kitesail", rfPermAKitesailOracle)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, sail, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the equipped creature has no flying")
	}
	life := me.Life
	effectstest.B784AttackEach(t, g, [2]uuid.UUID{bear, g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

func TestHuntersAxeGrantsTheChosenKeywordOnAttack(t *testing.T) {
	for i, kw := range []string{"trample", "deathtouch"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.AuraBear(g, me.ID)
		axe := rfPermAEquipment(g, me.ID, "Hunter's Axe", rfPermAHuntersAxeOracle)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatMana(t, g, me, "{C}{C}")
		effectstest.EquipTo(t, g, me.ID, axe, bear)
		if p := effectstest.EffectivePower(t, g, bear); p != 4 {
			t.Fatalf("power = %d, want 4", p)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
			t.Fatal(err)
		}
		// The trigger is harvested as the step advances; the prompt it
		// queues then stops the advance, which is the point.
		_, _ = g.AdvanceStep()
		effectstest.AnswerOptionPick(t, g, me.ID, i)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.EffectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("choice %d: the attacker lacks %s", i, kw)
		}
		other := "trample"
		if kw == "trample" {
			other = "deathtouch"
		}
		if effectstest.EffectiveAbilitiesContain(t, g, bear, other) {
			t.Errorf("choice %d: the attacker also has %s", i, other)
		}
	}
}

func TestFerocityOfTheHuntBuffsAndReturnsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Ferocity of the Hunt", rfPermAFerocityOracle, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("no deathtouch")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 {
		t.Fatalf("%d Bears on the battlefield, want 1", len(ids))
	}
	if !effectstest.B16Tapped(t, g, ids[0]) {
		t.Error("the Bear returned untapped")
	}
}

func TestMemoryTrapExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	trap := effectstest.CastCatalogSpell(t, g, "Memory Trap", "Enchantment", rfPermAMemoryTrapOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(trap) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after the trap left, %d copies under its owner's control, want 1", n)
	}
}

func TestMemoryTrapOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := effects.Lookup(rfPermAMemoryTrapOracle)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

func TestSolitaryCellOnlyReachesManaValueThreeOrLess(t *testing.T) {
	spec, _ := effects.Lookup(rfPermASolitaryCellOracle)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Push(g, opp.ID, "Small", "Creature — Bear", "", 2, 2)
	big := effectstest.B12Push(g, opp.ID, "Big", "Creature — Bear", "", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case small:
				g.Battlefield.Cards[i].ManaCost = "{2}{G}"
			case big:
				g.Battlefield.Cards[i].ManaCost = "{3}{G}"
			}
		}
	})
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{small: true, big: false})
}

func TestSolitaryCellExilesAndDiscardsALegendaryCardToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 2, 2)
	cell := effectstest.CastCatalogSpell(t, g, "Solitary Cell", "Artifact", rfPermASolitaryCellOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	legend := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: legend, Name: "A Legend", TypeLine: "Legendary Creature — Elf", Owner: me.ID, Controller: me.ID})
	before := effectstest.HandCount(g, me)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, cell, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{legend}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, legend) {
		t.Error("the legendary card was not discarded")
	}
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("hand %d, want %d (discard one, draw one)", got, before)
	}
}

func TestMurmuringVolumeLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	book := effectstest.B12Push(g, me.ID, "Murmuring Volume", "Artifact — Book", rfPermAMurmuringVolumeOracle, 0, 0)
	spare := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	before := effectstest.HandCount(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, book, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, spare) {
		t.Error("the card was not discarded")
	}
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("hand %d, want %d", got, before)
	}
}

func TestPuppetCraftingMakesAnArtifactAFiveFiveConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Push(g, me.ID, "Mind Stone", "Artifact", "c97361b5-af16-4a7b-af85-a429dbaf4ad2", 0, 0)
	effectstest.AuraCast(t, g, "Puppet Crafting", rfPermAPuppetCraftingOracle, rock)
	types := effectstest.EffectiveTypes(t, g, rock)
	if !rfPermAContains(types, "Creature") || !rfPermAContains(types, "Artifact") {
		t.Errorf("types = %v, want Artifact and Creature", types)
	}
	if !rfPermAContains(effectstest.EffectiveSubtypes(t, g, rock), "Construct") {
		t.Error("no Construct subtype")
	}
	if p := effectstest.EffectivePower(t, g, rock); p != 5 {
		t.Errorf("power = %d, want 5", p)
	}
}

func TestPuppetCraftingRefusesAnAuraAndACreature(t *testing.T) {
	spec, _ := effects.Lookup(rfPermAPuppetCraftingOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.B12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	ench := effectstest.B12Push(g, me.ID, "Some Enchantment", "Enchantment", "", 0, 0)
	rock := effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Targets, map[uuid.UUID]bool{bear: false, aura: false, ench: true, rock: true})
}

func rfPermAContains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestEchoverseFulcrumLootsThenWipesTheBoard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	fulcrum := effectstest.CastCatalogSpell(t, g, "The Echoverse Fulcrum", "Legendary Artifact", rfPermAEchoverseFulcrumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the enters trigger drew but asked for no discard")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, fulcrum, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("creatures survived the wipe")
	}
	if g.Battlefield.Contains(fulcrum) {
		t.Error("the Fulcrum was not exiled as a cost")
	}
}

func TestGardenizeCountsDeathsAndAddsManaAtMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garden := effectstest.B12Push(g, me.ID, "Gardenize", "Enchantment", rfPermAGardenizeOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	opp := effectstest.B12Creature(g, g.Seats[1].ID, "Opp", "Creature — Bear", 1, 1)
	for _, id := range []uuid.UUID{a, b} {
		effectstest.KillCreature(t, g, me.ID, id)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.KillCreature(t, g, g.Seats[1].ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, garden, game.CounterCharge); n != 2 {
		t.Fatalf("charge counters = %d, want 2", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("mana = %d, want 2", n)
	}
}

func TestAjanisAnguishDealsXAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	life := opp.Life
	effectstest.LftCastKicked(t, g, "Ajani's Anguish", "Enchantment", rfPermAAjanisAnguishOracle, "{X}{R}", 3, nil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d", opp.Life, life-3)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("creatures you control lack trample")
	}
}
