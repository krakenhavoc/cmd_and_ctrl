package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b2_cards_test.go — Teenage Mutant Ninja Turtles, batch 2 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB2NewsCrewOracle   = "a9d68cf5-3055-4ec9-9893-244bfd7046be"
	tmtB2PizzaOracle      = "5f4834c7-cd63-420b-8492-ef127ffd4f45"
	tmtB2AprilONeilOracle = "3f166376-2739-411c-8f42-96f2ef28ef86"
	tmtB2ReporterOracle   = "1c371003-e4f0-4d7c-b021-273176772f97"
	tmtB2BaxterOracle     = "e610bfd4-d064-410b-929c-1d9dd82828b1"
	tmtB2BebopOracle      = "39c813e1-0e26-440a-b9bd-e781f7576b8b"
	tmtB2BoOracle         = "679480aa-f9ad-4a6c-bd79-db9f9f75433d"
	tmtB2BuzzBotsOracle   = "434e720f-2bfa-49b6-a5ac-fe0c0b24764d"
	tmtB2CaseyOracle      = "b4c4fdb9-e034-49c0-9f7f-07e8cae19eb2"
	tmtB2CowabungaOracle  = "a1ddf985-1646-468c-b1a4-ab6bb8e9dd7c"
	tmtB2DeathOracle      = "f599471a-31a6-4fb0-bf27-c5b707e1db6a"
	tmtB2DimensionXOracle = "8683341a-e836-41e5-9034-4da09eb0ab42"
	tmtB2DimExileOracle   = "8f0777bc-4fca-4601-beed-1fd060f28075"
	tmtB2DonatelloOracle  = "f84850bc-6348-449e-bd82-bb39e2119bec"
)

func TestTmtB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB2NewsCrewOracle:   "Action News Crew",
		tmtB2PizzaOracle:      "Anchovy & Banana Pizza",
		tmtB2AprilONeilOracle: "April O'Neil, Kunoichi Trainee",
		tmtB2ReporterOracle:   "April, Reporter of the Weird",
		tmtB2BaxterOracle:     "Baxter Stockman",
		tmtB2BebopOracle:      "Bebop, Warthog Warrior",
		tmtB2BoOracle:         "Bespoke Bō",
		tmtB2BuzzBotsOracle:   "Buzz Bots",
		tmtB2CaseyOracle:      "Casey Jones, Jury-Rig Justiciar",
		tmtB2CowabungaOracle:  "Cowabunga!",
		tmtB2DeathOracle:      "Death in the Family",
		tmtB2DimensionXOracle: "Dimension X",
		tmtB2DimExileOracle:   "Dimensional Exile",
		tmtB2DonatelloOracle:  "Donatello, Turtle Techie",
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

// tmtB2Resolve passes priority until the stack is empty or a prompt is open.
func tmtB2Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
}

// --- Action News Crew ----------------------------------------------------------

func TestTmtB2ActionNewsCrewChannelCountersYourCreaturesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B")
	crew := pushCatalogHandCard(me, "Action News Crew", "Creature — Human Citizen", tmtB2NewsCrewOracle)
	if err := hobB4Activate(t, g, me.ID, crew, 0, "{C}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if got := zoneKindOf(g, crew); got != game.ZoneGraveyard {
		t.Fatalf("the channeled card is in %v, want the graveyard", got)
	}
	hand := me.Hand.Size()
	tmtB2Resolve(t, g)
	if plusOnes(g, mine) != 1 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: mine %d, theirs %d; want 1, 0", plusOnes(g, mine), plusOnes(g, theirs))
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Anchovy & Banana Pizza ----------------------------------------------------

func TestTmtB2AnchovyAndBananaPizzaDestroysACreatureAndIsAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pizza := castCatalogSpell(t, g, "Anchovy & Banana Pizza", "Artifact — Food", tmtB2PizzaOracle, nil)
	tmtB2Resolve(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(victim)...)
	tmtB2Resolve(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the target creature was not destroyed")
	}
	life := me.Life
	if err := hobB4Activate(t, g, me.ID, pizza, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("food: %v", err)
	}
	tmtB2Resolve(t, g)
	if me.Life != life+3 || g.Battlefield.Contains(pizza) {
		t.Errorf("life %d -> %d, Pizza on battlefield %v; want +3 and sacrificed", life, me.Life, g.Battlefield.Contains(pizza))
	}
}

// --- April O'Neil, Kunoichi Trainee --------------------------------------------

func TestTmtB2AprilONeilScriesTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three")
	castCatalogSpell(t, g, "April O'Neil, Kunoichi Trainee", "Legendary Creature — Human Ninja", tmtB2AprilONeilOracle, nil)
	tmtB2Resolve(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt on entry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
}

func TestTmtB2AprilONeilCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	april := b27Push(g, me.ID, "April O'Neil, Kunoichi Trainee", "Legendary Creature — Human Ninja", tmtB2AprilONeilOracle, "{1}{W}", 2, 2, "W")
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(april, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, april); err == nil {
		t.Error("a creature with power 3 blocked April O'Neil")
	}
	if err := g.DeclareBlocker(small, april); err != nil {
		t.Errorf("a creature with power 2 could not block: %v", err)
	}
}

// --- April, Reporter of the Weird ----------------------------------------------

func TestTmtB2AprilReporterDrawsThatManyThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D")
	april := b27Push(g, me.ID, "April, Reporter of the Weird", "Legendary Creature — Human Detective", tmtB2ReporterOracle, "{2}{U}", 2, 2, "U")
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, april, opp.ID, 3)
	tmtB2Resolve(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d before the discard, want 3", got)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed %d, want 1", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("net cards %d, want 2", got)
	}
}

// --- Baxter Stockman -----------------------------------------------------------

func TestTmtB2BaxterStockmanMakesARobotAndArmsAnArtifactCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := b12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Baxter Stockman", "Legendary Creature — Human Scientist", tmtB2BaxterOracle, nil)
	tmtB2Resolve(t, g)
	tmtB2Resolve(t, g)
	if got := tmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("made %d Robots, want 1", got)
	}

	advanceTo(t, g, game.StepBeginCombat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if hasID(pick.PickTargetCards, bear) {
		t.Error("a nonartifact creature is offered as a target")
	}
	pickCard(t, g, me.ID, golem)
	tmtB2Resolve(t, g)
	if got := effectivePower(t, g, golem); got != 5 {
		t.Errorf("the Golem has power %d, want 5", got)
	}
	assertKeywords(t, g, golem, "first strike", "vigilance")
	if effectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("the untargeted creature gained first strike")
	}
}

// --- Bebop, Warthog Warrior ----------------------------------------------------

func TestTmtB2BebopGivesYourRhinosMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bebop := b27Push(g, me.ID, "Bebop, Warthog Warrior", "Legendary Creature — Boar Mutant Warrior", tmtB2BebopOracle, "{4}{B}", 5, 4, "B")
	rhino := b12Creature(g, me.ID, "My Rhino", "Creature — Rhino", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Rhino", "Creature — Rhino", 4, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, bebop, "menace")
	assertKeywords(t, g, rhino, "menace")
	for _, id := range []uuid.UUID{theirs, bear} {
		if effectiveAbilitiesContain(t, g, id, "menace") {
			t.Errorf("%s gained menace", id)
		}
	}
}

func TestTmtB2BebopSwampcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	cid, p := cycleFromHand(t, g, "Bebop, Warthog Warrior", "Legendary Creature — Boar Mutant Warrior", tmtB2BebopOracle, "{C}{C}")
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	tmtB2Resolve(t, g)
	c := searchChoiceFor(g, p.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != swamp {
		t.Fatalf("want a search offering only the Swamp, got %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, p.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !p.Hand.Contains(swamp) {
		t.Error("swampcycling did not fetch the Swamp")
	}
}

// --- Bespoke Bō ----------------------------------------------------------------

func TestTmtB2BespokeBoBouncesOnEntryAndEquips(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	bo := castCatalogSpell(t, g, "Bespoke Bō", "Artifact — Equipment", tmtB2BoOracle, nil)
	tmtB2Resolve(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, bo) || hasID(pick.PickTargetCards, land) {
		t.Error("the Bō itself and a land are not legal targets")
	}
	pickCard(t, g, me.ID, rock)
	tmtB2Resolve(t, g)
	if !opp.Hand.Contains(rock) {
		t.Fatal("the artifact did not return to its owner's hand")
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, bo, 0, "{C}{C}{C}", cardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	tmtB2Resolve(t, g)
	if !attachedTo(t, g, bo, bear) {
		t.Fatal("the Bō is not attached")
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("equipped Bear is %d/%d, want 4/3", p, tough)
	}
	assertKeywords(t, g, bear, "vigilance")
}

// --- Buzz Bots -----------------------------------------------------------------

func TestTmtB2BuzzBotsDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	bots := b27Push(g, me.ID, "Buzz Bots", "Artifact Creature — Robot Insect", tmtB2BuzzBotsOracle, "{1}{U}", 1, 1, "U")
	assertKeywords(t, g, bots, "flying", "vigilance")
	hand := me.Hand.Size()
	b27Kill(g, bots)
	tmtB2Resolve(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Casey Jones, Jury-Rig Justiciar -------------------------------------------

func TestTmtB2CaseyJonesTakesAnArtifactFromTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Fourth", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Artifact"},
	)
	lib := me.Library.Size()
	casey := castCatalogSpell(t, g, "Casey Jones, Jury-Rig Justiciar", "Legendary Creature — Human Berserker", tmtB2CaseyOracle, nil)
	tmtB2Resolve(t, g)
	answerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the artifact did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went to the bottom")
	}
	assertKeywords(t, g, casey, "haste")
}

// --- Cowabunga! ----------------------------------------------------------------

func TestTmtB2CowabungaTakesATurtleFromTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Leo", TypeLine: "Creature — Mutant Ninja Turtle"},
		game.Card{Name: "Fourth", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Creature — Turtle"},
	)
	castCatalogSpell(t, g, "Cowabunga!", "Sorcery", tmtB2CowabungaOracle, nil)
	tmtB2Resolve(t, g)
	if err := func() error {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceChooseCards && pc.Chooser == me.ID {
				c = pc
			}
		}
		if c == nil {
			t.Fatal("no choose-cards prompt")
		}
		return g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]})
	}(); err == nil {
		t.Fatal("a Bear (neither Mutant, Ninja, Turtle nor land) was taken")
	}
	answerChooseCards(t, g, me.ID, ids[2])
	if !me.Hand.Contains(ids[2]) {
		t.Error("the Turtle did not reach the hand")
	}
}

func TestTmtB2CowabungaCanTakeALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "Cowabunga!", "Sorcery", tmtB2CowabungaOracle, nil)
	tmtB2Resolve(t, g)
	answerChooseCards(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Error("the land did not reach the hand")
	}
}

// --- Death in the Family -------------------------------------------------------

func TestTmtB2DeathInTheFamilyExilesOnlyManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b27Push(g, opp.ID, "Small Threat", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	big := b27Push(g, opp.ID, "Big Threat", "Creature — Beast", "", "{3}{G}", 4, 4, "G")
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogHandCard(me, "Death in the Family", "Instant", tmtB2DeathOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(big)}); err == nil {
		t.Fatal("a mana value 4 creature was a legal target")
	}
	castCatalogSpell(t, g, "Death in the Family", "Instant", tmtB2DeathOracle, cardRefs(small))
	tmtB2Resolve(t, g)
	if g.Battlefield.Contains(small) || zoneKindOf(g, small) != game.ZoneExile {
		t.Errorf("the mana value 3 creature is in %v, want exile", zoneKindOf(g, small))
	}
}

// --- Dimension X ---------------------------------------------------------------

func TestTmtB2DimensionXEntersTappedGainsOneAndTapsForRedOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := playLandFromHand(t, g, "Dimension X", tmtB2DimensionXOracle)
	top100AssertEnteredTapped(t, g, id, "Dimension X")
	tmtB2Resolve(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d -> %d, want +1", before, me.Life)
	}
	spec, _ := Lookup(tmtB2DimensionXOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|W}" {
		t.Errorf("mana abilities %+v, want one producing {R|W}", spec.ManaAbilities)
	}
}

// --- Dimensional Exile ---------------------------------------------------------

func TestTmtB2DimensionalExileExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	aura := castCatalogSpell(t, g, "Dimensional Exile", "Enchantment — Aura", tmtB2DimExileOracle, cardRefs(plains))
	tmtB2Resolve(t, g)
	b04WaitForPick(t, g, me.ID)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, mine) {
		t.Error("a creature you control is offered as a target")
	}
	pickCard(t, g, me.ID, victim)
	tmtB2Resolve(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if got := tmtB1Count(g, opp.ID, "Their Threat"); got != 1 {
		t.Errorf("after the Aura left, %d copies of the creature are back under its owner, want 1", got)
	}
}

// --- Donatello, Turtle Techie --------------------------------------------------

func TestTmtB2DonatelloDrawsOnlyWithAnArtifact(t *testing.T) {
	for _, withArtifact := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedLibrary(me, "A", "B")
		if withArtifact {
			b12Permanent(g, me.ID, "My Rock", "Artifact")
		}
		castCatalogSpell(t, g, "Donatello, Turtle Techie", "Legendary Creature — Mutant Ninja Turtle", tmtB2DonatelloOracle, nil)
		hand := me.Hand.Size()
		tmtB2Resolve(t, g)
		tmtB2Resolve(t, g)
		want := 0
		if withArtifact {
			want = 1
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("with an artifact %v: drew %d, want %d", withArtifact, got, want)
		}
	}
}
