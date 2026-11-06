package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_b1_cards_test.go — Foundations batch 1 (#4465): the A–C cards that
// needed no new seam. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	fdnB1AdamantWillOracle     = "467e22c3-6107-40ff-afc7-5960710c970b"
	fdnB1AngelicEdictOracle    = "36de8401-ee4e-413f-91ad-06924e39c857"
	fdnB1BakeIntoAPieOracle    = "1b9ec782-0ba1-41f1-bc39-d3302494ecb3"
	fdnB1BigfinBouncerOracle   = "4e473298-0b67-42b8-90a8-3bcdd73da473"
	fdnB1BurrogBefuddlerOracle = "73b60c2c-7c74-4554-8861-daa0d6fc22b4"
	fdnB1CrashThroughOracle    = "41c15160-ee8c-42e3-bc6c-593b8c8ae335"
	fdnB1CrowOracle            = "a4ffe297-5e82-43f9-91a4-7aa3d8dd3b4a"
	fdnB1MammothOracle         = "032b5fc4-5aca-41b6-9bf6-2c1ed0018968"
	fdnB1CyclopsOracle         = "16ab9451-f3da-4f61-8e2b-d9c84e0e5428"
)

func TestFdnB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnB1AdamantWillOracle:     "Adamant Will",
		fdnB1AngelicEdictOracle:    "Angelic Edict",
		fdnB1BakeIntoAPieOracle:    "Bake into a Pie",
		fdnB1BigfinBouncerOracle:   "Bigfin Bouncer",
		fdnB1BurrogBefuddlerOracle: "Burrog Befuddler",
		fdnB1CrashThroughOracle:    "Crash Through",
		fdnB1CrowOracle:            "Crow of Dark Tidings",
		fdnB1MammothOracle:         "Aggressive Mammoth",
		fdnB1CyclopsOracle:         "Crackling Cyclops",
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

// fdnB1HandSize is the size of p's hand.
func fdnB1HandSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}

// fdnB1GraveyardSize is the size of p's graveyard.
func fdnB1GraveyardSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Graveyard.Size() })
	return n
}

// --- Adamant Will -------------------------------------------------------------

func TestFdnB1AdamantWillPumpsAndMakesIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Adamant Will", "Instant", fdnB1AdamantWillOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Fatal("the target did not gain indestructible")
	}
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the indestructible creature was destroyed")
	}
}

// --- Angelic Edict ------------------------------------------------------------

func TestFdnB1AngelicEdictExilesACreatureOrAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	castCatalogSpell(t, g, "Angelic Edict", "Sorcery", fdnB1AngelicEdictOracle, cardRefs(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !g.Exile.Contains(aura) {
		t.Error("the enchantment was not exiled")
	}

	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Angelic Edict", "Sorcery", fdnB1AngelicEdictOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestFdnB1AngelicEdictRefusesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Angelic Edict", "Sorcery", fdnB1AngelicEdictOracle, cardRefs(rock)); err == nil {
		t.Error("a noncreature artifact was accepted as a target")
	}
}

// --- Bake into a Pie ----------------------------------------------------------

func TestFdnB1BakeIntoAPieDestroysAndMakesYouAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Bake into a Pie", "Instant", fdnB1BakeIntoAPieOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if got := tmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("the caster has %d Foods, want 1", got)
	}
	if got := tmtB1Count(g, opp.ID, "Food"); got != 0 {
		t.Errorf("the target's controller has %d Foods, want 0", got)
	}
}

func TestFdnB1BakeIntoAPieStillMakesTheFoodWhenTheCreatureSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"indestructible"}})
	castCatalogSpell(t, g, "Bake into a Pie", "Instant", fdnB1BakeIntoAPieOracle, cardRefs(wall))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("an indestructible creature was destroyed")
	}
	if got := tmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("the caster has %d Foods, want 1", got)
	}
}

// --- Bigfin Bouncer -----------------------------------------------------------

func TestFdnB1BigfinBouncerReturnsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Bigfin Bouncer", "Creature — Shark Pirate", fdnB1BigfinBouncerOracle, 3, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("your own creature was accepted as a target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature left the battlefield")
	}
}

// --- Burrog Befuddler ---------------------------------------------------------

func TestFdnB1BurrogBefuddlerShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	frog := castHoldCreature(t, g, "Burrog Befuddler", "Creature — Frog Wizard", fdnB1BurrogBefuddlerOracle, 2, 1)
	assertKeywords(t, g, frog, "flash")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("your own creature was accepted as a target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 1 || tough != 2 {
		t.Errorf("the target is %d/%d, want 1/2", p, tough)
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("your own creature's power is %d, want 2", got)
	}
}

// --- Crash Through ------------------------------------------------------------

func TestFdnB1CrashThroughGivesYourCreaturesTrampleAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Crash Through", "Sorcery", fdnB1CrashThroughOracle, nil)
	before := fdnB1HandSize(g, me)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, mine, "trample") {
		t.Error("your creature did not gain trample")
	}
	if effectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	if got := fdnB1HandSize(g, me); got != before+1 {
		t.Errorf("hand %d after the spell, want %d", got, before+1)
	}
	// The set is locked as the spell resolves (CR 611.2c).
	later := b12Creature(g, me.ID, "Later Bear", "Creature — Bear", 2, 2)
	if effectiveAbilitiesContain(t, g, later, "trample") {
		t.Error("a creature that entered after the spell resolved has trample")
	}
}

// --- Crow of Dark Tidings -----------------------------------------------------

func TestFdnB1CrowOfDarkTidingsMillsTwoOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := fdnB1GraveyardSize(g, me)
	crow := castHoldCreature(t, g, "Crow of Dark Tidings", "Creature — Zombie Bird", fdnB1CrowOracle, 2, 1)
	assertKeywords(t, g, crow, "flying")
	if got := fdnB1GraveyardSize(g, me); got != before+2 {
		t.Errorf("graveyard %d after the Crow entered, want %d", got, before+2)
	}
}

func TestFdnB1CrowOfDarkTidingsMillsTwoOnDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	crow := b27Push(g, me.ID, "Crow of Dark Tidings", "Creature — Zombie Bird", fdnB1CrowOracle, "{2}{B}", 2, 1, "B")
	before := fdnB1GraveyardSize(g, me)
	b27Kill(g, crow)
	passPriorityAroundTable(t, g)
	// The Crow itself, then the two milled cards.
	if got := fdnB1GraveyardSize(g, me); got != before+3 {
		t.Errorf("graveyard %d after the Crow died, want %d", got, before+3)
	}
}

// --- Aggressive Mammoth -------------------------------------------------------

func TestFdnB1AggressiveMammothGivesYourOtherCreaturesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mammoth := b12Push(g, me.ID, "Aggressive Mammoth", "Creature — Elephant", fdnB1MammothOracle, 8, 8)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, mammoth, "trample")
	if !effectiveAbilitiesContain(t, g, mine, "trample") {
		t.Error("your other creature does not have trample")
	}
	if effectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
	b27Kill(g, mammoth)
	g.RunStateChecksForTest()
	if effectiveAbilitiesContain(t, g, mine, "trample") {
		t.Error("your creature kept trample after the Mammoth left")
	}
}

// --- Crackling Cyclops --------------------------------------------------------

func TestFdnB1CracklingCyclopsGrowsOnYourNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cyclops := b12Push(g, me.ID, "Crackling Cyclops", "Creature — Cyclops Wizard", fdnB1CyclopsOracle, 0, 4)

	castFree(t, g, me, "A Creature", "Creature — Bear")
	if got := effectivePower(t, g, cyclops); got != 0 {
		t.Errorf("power after a creature spell = %d, want 0", got)
	}
	castFree(t, g, opp, "Their Instant", "Instant")
	if got := effectivePower(t, g, cyclops); got != 0 {
		t.Errorf("power after an opponent's spell = %d, want 0", got)
	}
	castFree(t, g, me, "A Sorcery", "Sorcery")
	if got := effectivePower(t, g, cyclops); got != 3 {
		t.Errorf("power after a noncreature spell = %d, want 3", got)
	}
	castFree(t, g, me, "An Instant", "Instant")
	if p, tough := effectivePower(t, g, cyclops), effectiveToughness(t, g, cyclops); p != 6 || tough != 4 {
		t.Errorf("the Cyclops is %d/%d after two noncreature spells, want 6/4", p, tough)
	}
}
