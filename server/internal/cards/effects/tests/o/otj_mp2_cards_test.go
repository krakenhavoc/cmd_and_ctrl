package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_mp2_cards_test.go — Outlaws of Thunder Junction (OTJ), the second
// M–P batch (#4468): cards that need no new engine seam. Each test plays
// the real catalog entry through the engine and asserts what the printed
// text says.

const (
	otjMP2MalcolmOracle             = "7932ab35-39dc-4500-aa44-2ad7228f0d2a"
	otjMP2MarchesaOracle            = "37449c93-4d32-4f1a-982f-8f3416214368"
	otjMP2MysticalTetherOracle      = "fb1a6986-39f7-4592-9039-7e30f5ccd009"
	otjMP2NeutralizeTheGuardsOracle = "b793141d-581c-4cc7-b09f-1fe8bd5536c5"
	otjMP2NurturingPixieOracle      = "8ecd3743-7c32-41ce-9ac1-c8c855a47302"
	otjMP2OutcasterGreenbladeOracle = "855a2579-8f78-46ce-a3a8-eddc808bd7b4"
	otjMP2OutcasterTrailblazer      = "3b5ff623-c44e-4f2d-8ab7-ee6b11f1deee"
	otjMP2OutlawsFuryOracle         = "5b5ef88c-f863-4f55-91fc-a63f7f3c1b3f"
	otjMP2PatientNaturalistOracle   = "ba7c1062-f777-4ba9-84df-ec340a9ec514"
	otjMP2PhantomInterference       = "0db20037-35b3-4b1f-8dd6-3ee797c421ca"
	otjMP2PillageTheBogOracle       = "91740570-1a97-4513-8af9-b32c363daf15"
	otjMP2PitilessCarnageOracle     = "186bfdac-ce9c-476f-89b6-ef467487fca5"
)

func TestOtjMP2EveryCardIsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		otjMP2MalcolmOracle:             "Malcolm, the Eyes",
		otjMP2MarchesaOracle:            "Marchesa, Dealer of Death",
		otjMP2MysticalTetherOracle:      "Mystical Tether",
		otjMP2NeutralizeTheGuardsOracle: "Neutralize the Guards",
		otjMP2NurturingPixieOracle:      "Nurturing Pixie",
		otjMP2OutcasterGreenbladeOracle: "Outcaster Greenblade",
		otjMP2OutcasterTrailblazer:      "Outcaster Trailblazer",
		otjMP2OutlawsFuryOracle:         "Outlaws' Fury",
		otjMP2PatientNaturalistOracle:   "Patient Naturalist",
		otjMP2PhantomInterference:       "Phantom Interference",
		otjMP2PillageTheBogOracle:       "Pillage the Bog",
		otjMP2PitilessCarnageOracle:     "Pitiless Carnage",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", name, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOtjMP2PlotCosts(t *testing.T) {
	for oracle, cost := range map[string]string{
		otjMP2OutcasterTrailblazer:  "{2}{G}",
		otjMP2PillageTheBogOracle:   "{1}{B}{G}",
		otjMP2PitilessCarnageOracle: "{1}{B}{B}",
	} {
		sa := effectstest.PlotOf(oracle)
		if sa == nil {
			t.Errorf("%s: no plot special action", oracle)
			continue
		}
		if sa.Cost != cost || sa.CastCost != "" {
			t.Errorf("%s: plot %q (cast %q), want %q and a free cast", oracle, sa.Cost, sa.CastCost, cost)
		}
	}
}

// --- Malcolm, the Eyes -----------------------------------------------------------

func TestOtjMP2MalcolmInvestigatesOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	malcolm := effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, the Eyes", "Legendary Creature — Siren Pirate", otjMP2MalcolmOracle, false)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, malcolm, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Clue"); got != 0 {
		t.Fatalf("the first spell made %d Clues", got)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Clue"); got != 1 {
		t.Fatalf("the second spell: %d Clues, want 1", got)
	}
	effectstest.B24CastNoncreature(t, g, "Third", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Clue"); got != 1 {
		t.Errorf("the third spell made another Clue: %d", got)
	}
}

// --- Marchesa, Dealer of Death ---------------------------------------------------

func TestOtjMP2MarchesaPaysOneOnACrimeToTakeOneOfTheTopTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Marchesa, Dealer of Death", "Legendary Creature — Human Rogue", otjMP2MarchesaOracle, false)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	lower := effectstest.TopLibraryCard(me, "Lower", "Sorcery")
	upper := effectstest.TopLibraryCard(me, "Upper", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("targeting your own creature is no crime")
	}

	otjB1BoltOpponent(t, g, opp.ID)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a crime: you are offered to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	if ask := effectstest.AnswerPayUnless(t, g, me.ID, true); ask.PayCost != "{1}" {
		t.Errorf("the cost is %s, want {1}", ask.PayCost)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("paid: no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lower) || !me.Graveyard.Contains(upper) {
		t.Errorf("one to hand (%v) and the other to the graveyard (%v)", me.Hand.Contains(lower), me.Graveyard.Contains(upper))
	}
}

func TestOtjMP2MarchesaDeclinedLooksAtNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Marchesa, Dealer of Death", "Legendary Creature — Human Rogue", otjMP2MarchesaOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	lib := me.Library.Size()
	otjB1BoltOpponent(t, g, opp.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("declined: no pick")
	}
	if me.Library.Size() != lib {
		t.Errorf("declined: library %d → %d", lib, me.Library.Size())
	}
}

// --- Mystical Tether -------------------------------------------------------------

func TestOtjMP2MysticalTetherExilesAnOpponentsCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	tether := effectstest.CastCatalogSpell(t, g, "Mystical Tether", "Enchantment", otjMP2MysticalTetherOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, rock) {
		t.Errorf("an opponent's artifact or creature: %v", p.PickTargetCards)
	}
	if effectstest.HasID(p.PickTargetCards, land) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("not a land, and not your own creature: %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("pick their Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("their Bear is exiled")
	}
	effectstest.B25Destroy(g, tether)
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("the Tether left: %d of their Bear back under their control, want 1", n)
	}
}

func TestOtjMP2MysticalTetherHasFlashForTwoMore(t *testing.T) {
	spec, ok := effects.Lookup(otjMP2MysticalTetherOracle)
	if !ok {
		t.Fatal("not registered")
	}
	// #5207: an optional additional {2} that opens the window, not an
	// alternative cost.
	if len(spec.AlternativeCosts) != 0 || len(spec.OptionalCosts) != 1 {
		t.Fatalf("%d alternative costs, %d optional costs; want the {2} optional cost only", len(spec.AlternativeCosts), len(spec.OptionalCosts))
	}
	if oc := spec.OptionalCosts[0]; !oc.AsThoughFlash || oc.ManaCost != "{2}" {
		t.Errorf("flash for {2} more: %+v", oc)
	}
}

// --- Neutralize the Guards -------------------------------------------------------

func TestOtjMP2NeutralizeTheGuardsShrinksTheOpponentsCreaturesAndSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.Msc18Library(me, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	weenie := effectstest.B12Creature(g, opp.ID, "Their Weenie", "Creature — Human", 1, 1)
	bystander := effectstest.B12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Neutralize the Guards", "Instant", otjMP2NeutralizeTheGuardsOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	s := effectstest.SurveilChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no surveil prompt")
	}
	if len(s.ScryCards) != 2 {
		t.Errorf("surveil 2 looks at %d cards", len(s.ScryCards))
	}
	if err := g.ResolveSurveil(s.ID, me.ID, nil, s.ScryCards); err != nil {
		t.Fatalf("surveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", p, tt)
	}
	if g.Battlefield.Contains(weenie) {
		t.Error("their 1/1 dies")
	}
	for _, id := range []uuid.UUID{bystander, mine} {
		if p, tt := effectstest.PtOf(t, g, id); p != 2 || tt != 2 {
			t.Errorf("only the target opponent's creatures: %s is %d/%d", id, p, tt)
		}
	}
}

func TestOtjMP2NeutralizeTheGuardsTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Neutralize the Guards", "Instant", otjMP2NeutralizeTheGuardsOracle, effectstest.B16TargetPlayer(me.ID)); err == nil {
		t.Error("you are not an opponent")
	}
}

// --- Nurturing Pixie -------------------------------------------------------------

func TestOtjMP2NurturingPixieReturnsANonFaerieAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	faerie := effectstest.B12Creature(g, me.ID, "My Faerie", "Creature — Faerie", 1, 1)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	pixie := effectstest.Msc17Play(t, g, me, game.Card{Name: "Nurturing Pixie", TypeLine: "Creature — Faerie Rogue",
		ManaCost: "{W}", OracleID: otjMP2NurturingPixieOracle, Power: 1, Toughness: 1})
	if !effectstest.HasEffectiveKeyword(t, g, pixie, "flying") {
		t.Error("flying is printed")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, bear) {
		t.Errorf("your non-Faerie, nonland permanent: %v", p.PickTargetCards)
	}
	for _, id := range []uuid.UUID{faerie, land, theirs, pixie} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("%s may not be targeted", id)
		}
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 1 {
		t.Errorf("up to one: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("pick the Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, bear) {
		t.Fatal("the Bear returns to your hand")
	}
	if got := effectstest.CountersOn(g, pixie, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Pixie, want 1", got)
	}
}

func TestOtjMP2NurturingPixieWithNoTargetGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pixie := effectstest.Msc17Play(t, g, me, game.Card{Name: "Nurturing Pixie", TypeLine: "Creature — Faerie Rogue",
		ManaCost: "{W}", OracleID: otjMP2NurturingPixieOracle, Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, pixie, game.CounterPlusOne); got != 0 {
		t.Errorf("nothing returned: %d counters, want 0", got)
	}
}

// --- Outcaster Greenblade --------------------------------------------------------

func TestOtjMP2OutcasterGreenbladeFindsADesertAndCountsDeserts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	desert := effectstest.StapleLibraryCard(me, "Sandy Desert", "Land — Desert")
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.Msc17Play(t, g, me, game.Card{Name: "Outcaster Greenblade", TypeLine: "Creature — Human Mercenary",
		ManaCost: "{2}{G}", OracleID: otjMP2OutcasterGreenbladeOracle, Power: 1, Toughness: 2})
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, desert)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, desert) {
		t.Fatal("the Desert card goes to your hand")
	}
	if p, tt := effectstest.PtOf(t, g, blade); p != 1 || tt != 2 {
		t.Fatalf("no Deserts: %d/%d, want 1/2", p, tt)
	}
	effectstest.B12Permanent(g, me.ID, "Desert One", "Land — Desert")
	effectstest.B12Permanent(g, me.ID, "Desert Two", "Land — Desert")
	effectstest.B12Permanent(g, opp.ID, "Their Desert", "Land — Desert")
	if p, tt := effectstest.PtOf(t, g, blade); p != 3 || tt != 4 {
		t.Errorf("two Deserts you control: %d/%d, want 3/4", p, tt)
	}
}

func TestOtjMP2OutcasterGreenbladeFindsABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Outcaster Greenblade", TypeLine: "Creature — Human Mercenary",
		ManaCost: "{2}{G}", OracleID: otjMP2OutcasterGreenbladeOracle, Power: 1, Toughness: 2})
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, forest) {
		t.Error("the basic land card goes to your hand")
	}
}

// --- Outcaster Trailblazer -------------------------------------------------------

func TestOtjMP2OutcasterTrailblazerAddsManaAndDrawsForBigCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Outcaster Trailblazer", TypeLine: "Creature — Human Druid",
		ManaCost: "{2}{G}", OracleID: otjMP2OutcasterTrailblazer, Power: 4, Toughness: 2})
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the enters trigger offers any colour: %+v", pick)
	}
	effectstest.RiderAnswerManaPicks(t, g, me.ID, "R")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want {R}", got)
	}
	hand := me.Hand.Size()
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear",
		ManaCost: "{0}", Power: 3, Toughness: 3})
	if me.Hand.Size() != hand {
		t.Fatalf("a 3-power creature draws nothing: hand %d → %d", hand, me.Hand.Size())
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Big Beast", TypeLine: "Creature — Beast",
		ManaCost: "{0}", Power: 4, Toughness: 4})
	if me.Hand.Size() != hand+1 {
		t.Errorf("a 4-power creature you control entered: hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

// --- Outlaws' Fury ---------------------------------------------------------------

func TestOtjMP2OutlawsFuryPumpsAndExilesWithAnOutlaw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Top", "Second")
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	rogue := effectstest.B12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Outlaws' Fury", "Instant", otjMP2OutlawsFuryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, rogue); p != 4 || tt != 2 {
		t.Errorf("your Rogue is %d/%d, want 4/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tt)
	}
	if !g.Exile.Contains(top) {
		t.Error("you control an outlaw: the top card is exiled")
	}
}

func TestOtjMP2OutlawsFuryWithoutAnOutlawExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top", "Second")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Outlaws' Fury", "Instant", otjMP2OutlawsFuryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("your Bear has power %d, want 4", p)
	}
	if me.Library.Size() != lib || g.Exile.Size() != 0 {
		t.Error("no outlaw: nothing is exiled")
	}
}

// --- Patient Naturalist ----------------------------------------------------------

func TestOtjMP2PatientNaturalistTakesAMilledLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom")
	land := effectstest.TopLibraryCard(me, "Milled Forest", "Basic Land — Forest")
	effectstest.TopLibraryCard(me, "Milled Spell", "Sorcery")
	effectstest.TopLibraryCard(me, "Milled Plains", "Basic Land — Plains")
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Patient Naturalist", TypeLine: "Creature — Human Scout",
		ManaCost: "{2}{G}", OracleID: otjMP2PatientNaturalistOracle, Power: 2, Toughness: 3})
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("two lands milled: no pick")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the two lands", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, land) {
		t.Error("the chosen land goes to your hand")
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("the other two milled cards stay in the graveyard: %d", me.Graveyard.Size())
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != 0 {
		t.Errorf("a land was taken: %d Treasures, want 0", got)
	}
}

func TestOtjMP2PatientNaturalistMakesATreasureWithNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Patient Naturalist", TypeLine: "Creature — Human Scout",
		ManaCost: "{2}{G}", OracleID: otjMP2PatientNaturalistOracle, Power: 2, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("no land milled: nothing to pick")
	}
	if me.Graveyard.Size() != 3 {
		t.Errorf("milled %d, want 3", me.Graveyard.Size())
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Phantom Interference --------------------------------------------------------

func TestOtjMP2PhantomInterferenceMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastModal(t, g, "Phantom Interference", "Instant", otjMP2PhantomInterference, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			spirits++
			if c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("a 2/2 white Spirit: %d/%d %v", c.Power, c.Toughness, c.Colors)
			}
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("the Spirit has flying")
			}
		}
	}
	if spirits != 1 {
		t.Errorf("%d Spirits, want 1", spirits)
	}
}

func TestOtjMP2PhantomInterferenceTaxesASpellTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Their Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastModal(t, g, "Phantom Interference", "Instant", otjMP2PhantomInterference,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, bolt, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the spell's controller is asked to pay")
	}
	if ask.Chooser != opp.ID || ask.PayCost != "{2}" {
		t.Errorf("asked %s to pay %s, want the Bolt's controller and {2}", ask.Chooser, ask.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("unpaid: the Bolt is countered")
	}
	if effectstest.B16CountNamed(g, "Spirit") != 0 {
		t.Error("only the counter bullet was chosen: no Spirit")
	}
}

// --- Pillage the Bog -------------------------------------------------------------

func TestOtjMP2PillageTheBogLooksAtTwiceYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B36Lands(g, me.ID, 2)
	effectstest.B36Lands(g, opp.ID, 3)
	effectstest.Msc18Library(me, 8)
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Pillage the Bog", "Sorcery", otjMP2PillageTheBogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want one of four (twice your two lands)", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	taken := pick.ChooseCards[2]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{taken}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, taken) {
		t.Error("the chosen card goes to your hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d → %d, want the rest back in it", lib, me.Library.Size())
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range pick.ChooseCards {
		if id != taken && !bottom[id] {
			t.Errorf("%s was not put on the bottom", id)
		}
	}
}

// --- Pitiless Carnage ------------------------------------------------------------

func TestOtjMP2PitilessCarnageSacrificesAnyNumberThenDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pitiless Carnage", "Sorcery", otjMP2PitilessCarnageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no sacrifice pick")
	}
	if !effectstest.HasID(p.ChooseCards, rock) || !effectstest.HasID(p.ChooseCards, land) || !effectstest.HasID(p.ChooseCards, bear) || effectstest.HasID(p.ChooseCards, theirs) {
		t.Errorf("every permanent you control, none of theirs: %v", p.ChooseCards)
	}
	if p.ChooseMin != 0 {
		t.Errorf("any number: min %d, want 0", p.ChooseMin)
	}
	hand := me.Hand.Size()
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock, land}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) || !g.Battlefield.Contains(bear) {
		t.Error("exactly the chosen permanents are sacrificed")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("then draw that many: hand %d → %d", hand, got)
	}
}
