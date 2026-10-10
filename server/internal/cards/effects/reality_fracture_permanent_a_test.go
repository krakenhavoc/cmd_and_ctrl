package effects

import (
	"testing"

	"github.com/google/uuid"

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
	return b12Push(g, owner, name, "Artifact — Equipment", oracle, 0, 0)
}

func TestMedicsKitesailPumpsGrantsFlyingAndGainsLifeOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	sail := rfPermAEquipment(g, me.ID, "Medic's Kitesail", rfPermAKitesailOracle)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, sail, bear)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the equipped creature has no flying")
	}
	life := me.Life
	b784AttackEach(t, g, [2]uuid.UUID{bear, g.Seats[1].ID})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

func TestHuntersAxeGrantsTheChosenKeywordOnAttack(t *testing.T) {
	for i, kw := range []string{"trample", "deathtouch"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := auraBear(g, me.ID)
		axe := rfPermAEquipment(g, me.ID, "Hunter's Axe", rfPermAHuntersAxeOracle)
		advanceToMain(t, g)
		floatMana(t, g, me, "{C}{C}")
		equipTo(t, g, me.ID, axe, bear)
		if p := effectivePower(t, g, bear); p != 4 {
			t.Fatalf("power = %d, want 4", p)
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
			t.Fatal(err)
		}
		// The trigger is harvested as the step advances; the prompt it
		// queues then stops the advance, which is the point.
		_, _ = g.AdvanceStep()
		answerOptionPick(t, g, me.ID, i)
		passPriorityAroundTable(t, g)
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("choice %d: the attacker lacks %s", i, kw)
		}
		other := "trample"
		if kw == "trample" {
			other = "deathtouch"
		}
		if effectiveAbilitiesContain(t, g, bear, other) {
			t.Errorf("choice %d: the attacker also has %s", i, other)
		}
	}
}

func TestFerocityOfTheHuntBuffsAndReturnsTheCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Ferocity of the Hunt", rfPermAFerocityOracle, bear)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
	if !effectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("no deathtouch")
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	ids := battlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 {
		t.Fatalf("%d Bears on the battlefield, want 1", len(ids))
	}
	if !b16Tapped(t, g, ids[0]) {
		t.Error("the Bear returned untapped")
	}
}

func TestMemoryTrapExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	trap := castCatalogSpell(t, g, "Memory Trap", "Enchantment", rfPermAMemoryTrapOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(trap) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after the trap left, %d copies under its owner's control, want 1", n)
	}
}

func TestMemoryTrapOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := Lookup(rfPermAMemoryTrapOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

func TestSolitaryCellOnlyReachesManaValueThreeOrLess(t *testing.T) {
	spec, _ := Lookup(rfPermASolitaryCellOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Push(g, opp.ID, "Small", "Creature — Bear", "", 2, 2)
	big := b12Push(g, opp.ID, "Big", "Creature — Bear", "", 4, 4)
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
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{small: true, big: false})
}

func rfPermAExpectLegal(t *testing.T, g *game.Game, caster uuid.UUID, spec *game.TargetSpec, want map[uuid.UUID]bool) {
	t.Helper()
	if spec == nil {
		t.Fatal("no target clause")
	}
	var cards []uuid.UUID
	g.ReadSnapshot(func() {
		cards = g.LegalTargetsForEffect(game.SourceChooser(caster), spec).Cards
	})
	in := map[uuid.UUID]bool{}
	for _, id := range cards {
		in[id] = true
	}
	for id, legal := range want {
		if in[id] != legal {
			t.Errorf("target %v legal = %v, want %v", id, in[id], legal)
		}
	}
}

func TestSolitaryCellExilesAndDiscardsALegendaryCardToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 2, 2)
	cell := castCatalogSpell(t, g, "Solitary Cell", "Artifact", rfPermASolitaryCellOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	legend := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: legend, Name: "A Legend", TypeLine: "Legendary Creature — Elf", Owner: me.ID, Controller: me.ID})
	before := handCount(g, me)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, cell, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{legend}})
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, legend) {
		t.Error("the legendary card was not discarded")
	}
	if got := handCount(g, me); got != before {
		t.Errorf("hand %d, want %d (discard one, draw one)", got, before)
	}
}

func TestMurmuringVolumeLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	book := b12Push(g, me.ID, "Murmuring Volume", "Artifact — Book", rfPermAMurmuringVolumeOracle, 0, 0)
	spare := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	before := handCount(g, me)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, book, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}})
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, spare) {
		t.Error("the card was not discarded")
	}
	if got := handCount(g, me); got != before {
		t.Errorf("hand %d, want %d", got, before)
	}
}

func TestPuppetCraftingMakesAnArtifactAFiveFiveConstruct(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "c97361b5-af16-4a7b-af85-a429dbaf4ad2", 0, 0)
	auraCast(t, g, "Puppet Crafting", rfPermAPuppetCraftingOracle, rock)
	types := effectiveTypes(t, g, rock)
	if !rfPermAContains(types, "Creature") || !rfPermAContains(types, "Artifact") {
		t.Errorf("types = %v, want Artifact and Creature", types)
	}
	if !rfPermAContains(effectiveSubtypes(t, g, rock), "Construct") {
		t.Error("no Construct subtype")
	}
	if p := effectivePower(t, g, rock); p != 5 {
		t.Errorf("power = %d, want 5", p)
	}
}

func TestPuppetCraftingRefusesAnAuraAndACreature(t *testing.T) {
	spec, _ := Lookup(rfPermAPuppetCraftingOracle)
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := b12Push(g, me.ID, "Some Aura", "Enchantment — Aura", "", 0, 0)
	ench := b12Push(g, me.ID, "Some Enchantment", "Enchantment", "", 0, 0)
	rock := b12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	rfPermAExpectLegal(t, g, me.ID, spec.Targets, map[uuid.UUID]bool{bear: false, aura: false, ench: true, rock: true})
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	fulcrum := castCatalogSpell(t, g, "The Echoverse Fulcrum", "Legendary Artifact", rfPermAEchoverseFulcrumOracle, nil)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the enters trigger drew but asked for no discard")
	}
	discardFromHand(t, g, me.ID)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, fulcrum, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("creatures survived the wipe")
	}
	if g.Battlefield.Contains(fulcrum) {
		t.Error("the Fulcrum was not exiled as a cost")
	}
}

func TestGardenizeCountsDeathsAndAddsManaAtMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garden := b12Push(g, me.ID, "Gardenize", "Enchantment", rfPermAGardenizeOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	opp := b12Creature(g, g.Seats[1].ID, "Opp", "Creature — Bear", 1, 1)
	for _, id := range []uuid.UUID{a, b} {
		killCreature(t, g, me.ID, id)
		passPriorityAroundTable(t, g)
	}
	killCreature(t, g, g.Seats[1].ID, opp)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, garden, game.CounterCharge); n != 2 {
		t.Fatalf("charge counters = %d, want 2", n)
	}
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := len(me.ManaPool); n != 2 {
		t.Errorf("mana = %d, want 2", n)
	}
}

func TestAjanisAnguishDealsXAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	life := opp.Life
	lftCastKicked(t, g, "Ajani's Anguish", "Enchantment", rfPermAAjanisAnguishOracle, "{X}{R}", 3, nil, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d", opp.Life, life-3)
	}
	if !effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("creatures you control lack trample")
	}
}
