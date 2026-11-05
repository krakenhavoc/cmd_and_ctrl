package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b44_cards_test.go — Star Trek Commander, batch 44 (#4974). Each
// test plays the real catalog entry through the engine and asserts what
// its text says.

const (
	trekB44RikerOracle   = "a479b8b6-79de-4810-9165-992a6a10dc1d"
	trekB44WorfOracle    = "26e86be4-52e1-47cb-b548-c434482bd3ad"
	trekB44StasisOracle  = "846ba66c-a4eb-4d00-a0f1-d9f9524b9695"
	trekB44ExocompOracle = "fb208157-67e5-45e2-bf08-7671f65cf520"
	trekB44PriestOracle  = "b03c6720-b899-4eba-b28d-54ded03e086d"
	trekB44LaForgeOracle = "338619f2-bb33-4988-bafe-7460f86acef8"
	trekB44PeliaOracle   = "4c0be0d0-766b-4c50-b1ca-cce527196666"
	trekB44TPolOracle    = "e8a548e8-057e-4023-ac20-bac973dbce09"
	trekB44BadgeyOracle  = "227fb13b-0a07-4937-9048-7cf02efbc8e3"
	trekB44EelOracle     = "b1914241-6b2b-4b3e-8626-aeb0b8cbc12d"
)

func TestTrekB44EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB44RikerOracle:   "Will Riker, Assuming Command",
		trekB44WorfOracle:    "Worf, Chief Tactical Officer",
		trekB44StasisOracle:  "Cryogenic Stasis",
		trekB44ExocompOracle: "Exocomp",
		trekB44PriestOracle:  "Kolinahr Priest",
		trekB44LaForgeOracle: "La Forge, Perceptive Engineer",
		trekB44PeliaOracle:   "Pelia, Immortal Innovator",
		trekB44TPolOracle:    "T'Pol, Vulcan Representative",
		trekB44BadgeyOracle:  "Badgey, Malicious Glitch",
		trekB44EelOracle:     "Ceti Eel",
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

// --- Will Riker, Assuming Command ---------------------------------------------

func TestTrekB44RikerPutsACounterOnAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	riker := rfdCast(t, g, "Will Riker, Assuming Command", "Legendary Creature — Human Officer", trekB44RikerOracle, 4, 5)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters on the Bear %d, want 1", n)
	}
	if n := counterCount(g, riker, "+1/+1"); n != 0 {
		t.Errorf("Riker has %d +1/+1 counters; the counter goes on another creature", n)
	}
}

func TestTrekB44RikerCannotTargetItselfOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	riker := rfdCast(t, g, "Will Riker, Assuming Command", "Legendary Creature — Human Officer", trekB44RikerOracle, 4, 5)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the trigger asked for a target with no other creature you control")
	}
	if counterCount(g, theirs, "+1/+1") != 0 || counterCount(g, riker, "+1/+1") != 0 {
		t.Error("a counter landed on Riker or on an opponent's creature")
	}
}

// --- Worf, Exocomp, Kolinahr Priest: the printed keywords --------------------

func TestTrekB44KeywordCreaturesHaveTheirKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worf := b12Push(g, me.ID, "Worf, Chief Tactical Officer", "Legendary Creature — Klingon Warrior Officer", trekB44WorfOracle, 2, 3)
	exocomp := b12Push(g, me.ID, "Exocomp", "Artifact Creature — Construct", trekB44ExocompOracle, 3, 1)
	priest := b12Push(g, me.ID, "Kolinahr Priest", "Creature — Vulcan Monk", trekB44PriestOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, worf, "vigilance") {
		t.Error("Worf lacks vigilance")
	}
	if !hasEffectiveKeyword(t, g, exocomp, "flying") {
		t.Error("Exocomp lacks flying")
	}
	if !hasEffectiveKeyword(t, g, priest, "flash") {
		t.Error("Kolinahr Priest lacks flash")
	}
}

func TestTrekB44WorfAttacksWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worf := b12Push(g, me.ID, "Worf, Chief Tactical Officer", "Legendary Creature — Klingon Warrior Officer", trekB44WorfOracle, 2, 3)
	declareAttack(t, g, opp.ID, worf)
	lockInAttacks(t, g)
	if isTapped(g, worf) {
		t.Error("Worf tapped to attack despite vigilance")
	}
}

func TestTrekB44KolinahrPriestCastsAtInstantSpeedAndExocompDoesNot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{U}")
	exo := rfdHandCard(me, "Exocomp", "Artifact Creature — Construct", "{2}{U}", trekB44ExocompOracle, 3, 1)
	if err := g.CastSpell(me.ID, exo, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Exocomp, which has no flash, was cast in the beginning of combat step")
	}
	priest := rfdHandCard(me, "Kolinahr Priest", "Creature — Vulcan Monk", "{2}{U}", trekB44PriestOracle, 3, 2)
	if err := g.CastSpell(me.ID, priest, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("flash cast in the beginning of combat step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(priest) {
		t.Error("Kolinahr Priest did not resolve onto the battlefield")
	}
}

// --- Cryogenic Stasis ---------------------------------------------------------

func TestTrekB44CryogenicStasisTapsStunsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{U}")
	castFromHandForTest(t, g, me, "Cryogenic Stasis", "Instant", "{1}{U}", trekB44StasisOracle, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the target was not tapped")
	}
	if n := counterCount(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d, want 1", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the spell, want %d (cast it, drew one)", got, hand+1)
	}
}

func TestTrekB44CryogenicStasisRefusesANoncreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushArtifactForTest(g, opp.ID, "Their Rock")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{U}")
	id := rfdHandCard(me, "Cryogenic Stasis", "Instant", "{1}{U}", trekB44StasisOracle, 0, 0)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	if err == nil {
		t.Error("Cryogenic Stasis was cast targeting a noncreature artifact")
	}
}

// --- La Forge, Perceptive Engineer ---------------------------------------------

func TestTrekB44LaForgeManaPaysOnlyForArtifactSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	laForge := pushCatalogPermanent(g, me.ID, "La Forge, Perceptive Engineer", "Legendary Creature — Human Artificer", trekB44LaForgeOracle, false)
	if !hasEffectiveKeyword(t, g, laForge, "vigilance") {
		t.Error("La Forge lacks vigilance")
	}
	if err := g.ActivateManaAbility(me.ID, laForge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	cost, _ := game.ParseCost("{U}")
	rock := game.Card{Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{U}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("the mana was refused for an artifact spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a nonartifact spell: stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the mana paid for an ability: it is for artifact spells only")
	}
}

func TestTrekB44LaForgeCannotTapForManaWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	laForge := pushCatalogPermanent(g, me.ID, "La Forge, Perceptive Engineer", "Legendary Creature — Human Artificer", trekB44LaForgeOracle, true)
	if err := g.ActivateManaAbility(me.ID, laForge, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick La Forge tapped for mana (CR 302.6)")
	}
}

// --- Pelia, Immortal Innovator ---------------------------------------------------

func TestTrekB44PeliaBouncesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rfdCast(t, g, "Pelia, Immortal Innovator", "Legendary Creature — Lanthanite Artificer", trekB44PeliaOracle, 3, 3)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
}

func TestTrekB44PeliaCannotTargetYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rfdCast(t, g, "Pelia, Immortal Innovator", "Legendary Creature — Lanthanite Artificer", trekB44PeliaOracle, 3, 3)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the trigger asked for a target with no opponent's creature")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was bounced")
	}
}

// --- T'Pol, Vulcan Representative --------------------------------------------------

func TestTrekB44TPolTapsAndPutsTwoStunCounters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rfdCast(t, g, "T'Pol, Vulcan Representative", "Legendary Creature — Vulcan Scientist", trekB44TPolOracle, 4, 3)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the target was not tapped")
	}
	if n := counterCount(g, theirs, game.CounterStun); n != 2 {
		t.Errorf("stun counters %d, want 2", n)
	}
}

func TestTrekB44TPolCannotTargetYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rfdCast(t, g, "T'Pol, Vulcan Representative", "Legendary Creature — Vulcan Scientist", trekB44TPolOracle, 4, 3)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the trigger asked for a target with no opponent's creature")
	}
	if isTapped(g, mine) || counterCount(g, mine, game.CounterStun) != 0 {
		t.Error("your own creature was tapped or stunned")
	}
}

// --- Badgey, Malicious Glitch ------------------------------------------------------

func trekB44Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

func TestTrekB44BadgeyDrainsWhenTappedByAnEffect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	badgey := b12Push(g, me.ID, "Badgey, Malicious Glitch", "Legendary Artifact Creature — Illusion Horror", trekB44BadgeyOracle, 4, 3)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := trekB44Lives(g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(other) })
	passPriorityAroundTable(t, g)
	if after := trekB44Lives(g); after[0] != before[0] || after[1] != before[1] {
		t.Errorf("another creature becoming tapped changed life: %v -> %v", before, after)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(badgey) })
	passPriorityAroundTable(t, g)
	after := trekB44Lives(g)
	if after[0] != before[0]+1 {
		t.Errorf("your life %d, want %d", after[0], before[0]+1)
	}
	for i := 1; i < len(after); i++ {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d, want %d", i, after[i], before[i]-1)
		}
	}
}

func TestTrekB44BadgeyDrainsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	badgey := b12Push(g, me.ID, "Badgey, Malicious Glitch", "Legendary Artifact Creature — Illusion Horror", trekB44BadgeyOracle, 4, 3)
	before := trekB44Lives(g)
	declareAttack(t, g, opp.ID, badgey)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	after := trekB44Lives(g)
	if after[0] != before[0]+1 {
		t.Errorf("your life %d, want %d", after[0], before[0]+1)
	}
	for i := 1; i < len(after); i++ {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d, want %d", i, after[i], before[i]-1)
		}
	}
}

// --- Ceti Eel ----------------------------------------------------------------------

func TestTrekB44CetiEelMillsThenCountsArtifactAndCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			cardNamed("Old Rock", "Artifact"),
			cardNamed("Old Myr", "Artifact Creature — Myr"),
			cardNamed("Old Spell", "Instant"),
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Graveyard.PushTop(c)
		}
		libraryTop(me, cardNamed("Milled Land", "Land"), cardNamed("Milled Bear", "Creature — Bear"))
	})
	library := me.Library.Size()
	eel := rfdCast(t, g, "Ceti Eel", "Creature — Worm", trekB44EelOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-2 {
		t.Errorf("library %d, want %d after milling two", got, library-2)
	}
	// Old Rock, Old Myr (once) and the milled Bear: three.
	if n := counterCount(g, eel, "+1/+1"); n != 3 {
		t.Errorf("+1/+1 counters %d, want 3", n)
	}
}

func TestTrekB44CetiEelWithNothingToCountGetsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		libraryTop(me, cardNamed("Milled Land", "Land"), cardNamed("Milled Spell", "Sorcery"))
	})
	eel := rfdCast(t, g, "Ceti Eel", "Creature — Worm", trekB44EelOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, eel, "+1/+1"); n != 0 {
		t.Errorf("+1/+1 counters %d, want none", n)
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("graveyard %d, want the two milled cards", me.Graveyard.Size())
	}
}
