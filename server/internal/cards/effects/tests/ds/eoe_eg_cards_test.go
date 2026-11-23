package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_eg_cards_test.go — Edge of Eternities cards whose names start with E
// to G (#4461), the ones the E–L batch left. Each test plays the real
// catalog entry and asserts what its text says, and what it refuses.

const (
	eoeEGRoverOracle       = "b907d375-b715-446b-8966-79959feb9f58"
	eoeEGOblivionOracle    = "0fad5713-c09c-43d3-9cc0-10c65aead488"
	eoeEGEscortOracle      = "7d8b53f4-bdb5-4ba1-b6c0-f4846d8d3565"
	eoeEGEusocialOracle    = "2f3f5a1c-8b88-4080-bd80-14db4d6c8c50"
	eoeEGSunbornOracle     = "43febbcd-ecbe-4edb-ad90-7fc6dc381d05"
	eoeEGSaviorOracle      = "8b13ec4d-812a-48b0-8a64-6847a694a03a"
	eoeEGGravshipOracle    = "1a82be68-3b74-4dfc-9068-3abea61db709"
	eoeEGCoordinatorOracle = "57927909-11a9-44ef-989d-1dcca1f6e23f"
	eoeEGFocusFireOracle   = "6c4514be-fda7-45e8-83fe-04d49e2b932c"
	eoeEGWarRagerOracle    = "610d006a-3ceb-4136-b92b-a500c74d06b0"
	eoeEGFullBoreOracle    = "4b96d003-d3b4-4181-9f50-06a3e6eb1916"
	eoeEGColossusOracle    = "ce25dd89-dcc7-4576-b148-ef4c1cd5290c"
	eoeEGImagoOracle       = "bf3a0585-3976-4078-8beb-6afa79f97027"
	eoeEGPollinatorOracle  = "0a356d04-2e86-435b-9766-27b93b468a61"
	eoeEGTitanOracle       = "ef06e10c-553b-45b8-abf0-1d2e87c22128"
	eoeEGGodmawOracle      = "300e48df-7f7f-47ff-b450-8ba7bd1b1c67"
	eoeEGHeavyOracle       = "67821712-e58f-4fee-81af-021059ac72ea"
)

func TestEoeEGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeEGRoverOracle:       "Edge Rover",
		eoeEGOblivionOracle:    "Embrace Oblivion",
		eoeEGEscortOracle:      "Emissary Escort",
		eoeEGEusocialOracle:    "Eusocial Engineering",
		eoeEGSunbornOracle:     "Exalted Sunborn",
		eoeEGSaviorOracle:      "Exosuit Savior",
		eoeEGGravshipOracle:    "Fell Gravship",
		eoeEGCoordinatorOracle: "Flight-Deck Coordinator",
		eoeEGFocusFireOracle:   "Focus Fire",
		eoeEGWarRagerOracle:    "Frontline War-Rager",
		eoeEGFullBoreOracle:    "Full Bore",
		eoeEGColossusOracle:    "Fungal Colossus",
		eoeEGImagoOracle:       "Genemorph Imago",
		eoeEGPollinatorOracle:  "Gene Pollinator",
		eoeEGTitanOracle:       "Gigastorm Titan",
		eoeEGGodmawOracle:      "Glacier Godmaw",
		eoeEGHeavyOracle:       "Gravblade Heavy",
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
	for oracle, cost := range map[string]string{
		eoeEGEusocialOracle: "{1}{G}",
		eoeEGSunbornOracle:  "{1}{W}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- helpers ---------------------------------------------------------------

// eoeEGPushCosted seeds a noncreature permanent with a mana cost, so its
// mana value is read from the cost.
func eoeEGPushCosted(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: owner, Controller: owner,
	})
}

// eoeEGTap taps each permanent in place.
func eoeEGTap(g *game.Game, ids ...uuid.UUID) {
	for _, id := range ids {
		effectstest.EoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
}

// --- Edge Rover ------------------------------------------------------------

func TestEoeEGEdgeRoverGivesEachPlayerALanderWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rover := effectstest.B12Push(g, me.ID, "Edge Rover", "Artifact Creature — Robot Scout", eoeEGRoverOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, rover, "reach") {
		t.Error("Edge Rover has reach")
	}
	effectstest.B15Destroy(t, g, rover)
	for i, p := range g.Seats {
		if got := effectstest.EoeCDCount(g, p.ID, "Lander"); got != 1 {
			t.Errorf("seat %d has %d Landers, want 1", i, got)
		}
	}
}

// --- Embrace Oblivion --------------------------------------------------------

func TestEoeEGEmbraceOblivionSacrificesToDestroyASpacecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	ship := effectstest.B12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Embrace Oblivion", TypeLine: "Sorcery", ManaCost: "{B}", OracleID: eoeEGOblivionOracle},
		game.CastSpellParams{Targets: effectstest.CardRefs(ship), SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the artifact is sacrificed as the spell is cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ship) || !opp.Graveyard.Contains(ship) {
		t.Error("a Spacecraft that isn't a creature is destroyed")
	}
}

func TestEoeEGEmbraceOblivionNeedsTheSacrificeAndARightTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	for name, params := range map[string]game.CastSpellParams{
		"no sacrifice":         {Targets: effectstest.CardRefs(theirs)},
		"a land sacrificed":    {Targets: effectstest.CardRefs(theirs), SacrificeIDs: []uuid.UUID{effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")}},
		"an artifact targeted": {Targets: effectstest.CardRefs(rock), SacrificeIDs: []uuid.UUID{bear}},
	} {
		id := effectstest.AvB1Hand(me, game.Card{Name: "Embrace Oblivion", TypeLine: "Sorcery", ManaCost: "{B}", OracleID: eoeEGOblivionOracle})
		if err := g.CastSpell(me.ID, id, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a refused cast sacrificed the creature")
	}
}

// --- Emissary Escort ---------------------------------------------------------

func TestEoeEGEmissaryEscortGetsTheGreatestManaValueAmongOtherArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	escort := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Emissary Escort", TypeLine: "Artifact Creature — Robot Soldier", ManaCost: "{1}{U}",
		OracleID: eoeEGEscortOracle, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if p, tough := effectstest.EoeCDPT(t, g, escort); p != 0 || tough != 4 {
		t.Errorf("with no other artifact it is %d/%d, want 0/4 (it doesn't count itself)", p, tough)
	}
	eoeEGPushCosted(g, me.ID, "Small Rock", "Artifact", "{1}")
	if p, _ := effectstest.EoeCDPT(t, g, escort); p != 1 {
		t.Errorf("with a mana value 1 artifact the power is %d, want 1", p)
	}
	eoeEGPushCosted(g, me.ID, "Big Rock", "Artifact", "{4}")
	eoeEGPushCosted(g, me.ID, "Big Spell Stone", "Enchantment", "{6}")
	eoeEGPushCosted(g, opp.ID, "Their Rock", "Artifact", "{7}")
	if p, tough := effectstest.EoeCDPT(t, g, escort); p != 4 || tough != 4 {
		t.Errorf("the greatest among your other artifacts is 4: %d/%d, want 4/4", p, tough)
	}
}

// --- Eusocial Engineering ----------------------------------------------------

func TestEoeEGEusocialEngineeringMakesARobotOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Eusocial Engineering", "Enchantment", eoeEGEusocialOracle, 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	robot := effectstest.AvB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if c := effectstest.LayeredCard(t, g, robot); c.Tapped || !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("an untapped 2/2 Robot artifact creature: tapped %v, %d/%d", c.Tapped, c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.B13PlayAs(t, g, 1, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoeCDCount(g, me.ID, "Robot"); got != 1 {
		t.Errorf("an opponent's land made a Robot: %d, want 1", got)
	}
}

func TestEoeEGEusocialEngineeringWarpedIsExiledAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eng := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Eusocial Engineering", TypeLine: "Enchantment", ManaCost: "{3}{G}{G}",
		OracleID: eoeEGEusocialOracle}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eng) {
		t.Fatal("the warped enchantment enters")
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(eng) {
		t.Error("the warped enchantment is exiled at the end step")
	}
}

// --- Exalted Sunborn ---------------------------------------------------------

func TestEoeEGExaltedSunbornDoublesOnlyYourTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sunborn := effectstest.B12Push(g, me.ID, "Exalted Sunborn", "Creature — Angel Wizard", eoeEGSunbornOracle, 4, 5)
	effectstest.AssertKeywords(t, g, sunborn, "flying", "lifelink")
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if got := effectstest.EoeCDCount(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("your one Treasure became %d, want 2", got)
	}
	effectstest.MakeTokens(t, g, opp.ID, effects.TreasureToken(), 1)
	if got := effectstest.EoeCDCount(g, opp.ID, "Treasure"); got != 1 {
		t.Errorf("an opponent's one Treasure became %d, want 1", got)
	}
}

// --- Exosuit Savior ----------------------------------------------------------

func TestEoeEGExosuitSaviorReturnsAnotherPermanentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	savior := effectstest.CastHoldCreature(t, g, "Exosuit Savior", "Creature — Human Soldier", eoeEGSaviorOracle, 2, 2)
	effectstest.AssertKeywords(t, g, savior, "flying")
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{theirs, savior} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your other permanent returns to your hand")
	}
	if !g.Battlefield.Contains(savior) || !g.Battlefield.Contains(theirs) {
		t.Error("the Savior and the opponent's creature stay")
	}
}

// --- Fell Gravship -----------------------------------------------------------

func TestEoeEGFellGravshipMillsThreeAndReturnsASpacecraftCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	milled := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Milled Opt", TypeLine: "Instant"},
		game.Card{Name: "Milled Ship", TypeLine: "Artifact — Spacecraft"},
		game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
	ship := effectstest.DftABCast(t, g, me, game.Card{Name: "Fell Gravship", TypeLine: "Artifact — Spacecraft", ManaCost: "{2}{B}",
		OracleID: eoeEGGravshipOracle, Power: 3, Toughness: 2})
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c != nil {
		t.Fatalf("asked a pick with one candidate: %+v", c.ChooseCards)
	}
	if !me.Hand.Contains(milled[1]) {
		t.Error("the milled Spacecraft card returns to the hand")
	}
	if !me.Graveyard.Contains(milled[0]) || !me.Graveyard.Contains(milled[2]) {
		t.Error("the instant and the plain artifact stay in the graveyard")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 7); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("below 8 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at8 := effectstest.LayeredCard(t, g, ship)
	if !at8.IsCreature() || !game.HasKeyword(&at8, "flying") || !game.HasKeyword(&at8, "lifelink") ||
		at8.CurrentPower() != 3 || at8.CurrentToughness() != 2 {
		t.Errorf("at 8+ it is a 3/2 creature with flying and lifelink, got %d/%d", at8.CurrentPower(), at8.CurrentToughness())
	}
}

func TestEoeEGFellGravshipAsksWhichOfTwoCandidates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	bear := effectstest.DftABInGraveyard(me, "Old Bear", "Creature — Bear")
	milled := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Milled Ship", TypeLine: "Artifact — Spacecraft"},
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Opt", TypeLine: "Instant"})
	effectstest.DftABCast(t, g, me, game.Card{Name: "Fell Gravship", TypeLine: "Artifact — Spacecraft", ManaCost: "{2}{B}",
		OracleID: eoeEGGravshipOracle, Power: 3, Toughness: 2})
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no return pick with a creature card and a Spacecraft card in the graveyard")
	}
	if len(c.ChooseCards) != 2 || c.ChooseMin != 1 {
		t.Fatalf("the pick offers %d cards with floor %d, want the Bear and the Spacecraft with floor 1", len(c.ChooseCards), c.ChooseMin)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Graveyard.Contains(milled[0]) {
		t.Error("only the chosen creature card returns")
	}
}

// --- Flight-Deck Coordinator -------------------------------------------------

func TestEoeEGFlightDeckCoordinatorGainsTwoWithTwoTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	coord := effectstest.B12Push(g, me.ID, "Flight-Deck Coordinator", "Creature — Human Soldier", eoeEGCoordinatorOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeEGTap(g, coord, bear)
	life := me.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestEoeEGFlightDeckCoordinatorNeedsTwoTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coord := effectstest.B12Push(g, me.ID, "Flight-Deck Coordinator", "Creature — Human Soldier", eoeEGCoordinatorOracle, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Tapped Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	eoeEGTap(g, coord, rock, theirs)
	life := me.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("one tapped creature of yours changed life %d -> %d", life, me.Life)
	}
}

// --- Focus Fire --------------------------------------------------------------

func TestEoeEGFocusFireCountsYourCreaturesAndSpacecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Docked Ship", "Artifact — Spacecraft")
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Giant", 6, 6)
	if err := effectstest.CastCatalogSpellErr(t, g, "Focus Fire", "Instant", eoeEGFocusFireOracle, effectstest.CardRefs(attacker)); err == nil {
		t.Fatal("Focus Fire targeted a creature that is neither attacking nor blocking")
	}
	effectstest.EoeCDSet(g, attacker, func(c *game.Card) { c.AttackingTarget = me.ID })
	effectstest.CastCatalogSpell(t, g, "Focus Fire", "Instant", eoeEGFocusFireOracle, effectstest.CardRefs(attacker))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, attacker)
	if !ok {
		t.Fatal("the 6/6 died to what should be 4 damage")
	}
	if c.DamageMarked != 4 {
		t.Errorf("damage %d, want 4 (2, plus your creature and your Spacecraft)", c.DamageMarked)
	}
}

// --- Frontline War-Rager -----------------------------------------------------

func TestEoeEGFrontlineWarRagerGrowsWithTwoTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rager := effectstest.B12Push(g, me.ID, "Frontline War-Rager", "Creature — Kavu Soldier", eoeEGWarRagerOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeEGTap(g, rager, bear)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, rager, game.CounterPlusOne); n != 1 {
		t.Errorf("the War-Rager has %d counters, want 1", n)
	}
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the other creature has %d counters", n)
	}
}

func TestEoeEGFrontlineWarRagerRechecksAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rager := effectstest.B12Push(g, me.ID, "Frontline War-Rager", "Creature — Kavu Soldier", eoeEGWarRagerOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeEGTap(g, rager, bear)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.EoeCDSet(g, bear, func(c *game.Card) { c.Tapped = false })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, rager, game.CounterPlusOne); n != 0 {
		t.Errorf("the condition failed before it resolved, yet it put %d counters", n)
	}
}

// --- Full Bore ---------------------------------------------------------------

func TestEoeEGFullBoreGrantsTrampleAndHasteOnlyToAWarpedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wurm := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Germinating Wurm", TypeLine: "Creature — Plant Wurm", ManaCost: "{4}{G}",
		OracleID: eoeELWurmOracle, Power: 5, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the warped Wurm did not enter")
	}
	effectstest.CastCatalogSpell(t, g, "Full Bore", "Instant", eoeEGFullBoreOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 5 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 5/4", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") || effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a creature cast for its mana cost (here, not cast at all) gains no keywords")
	}
	effectstest.CastCatalogSpell(t, g, "Full Bore", "Instant", eoeEGFullBoreOracle, effectstest.CardRefs(wurm))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, wurm); p != 8 || tough != 7 {
		t.Errorf("the Wurm is %d/%d, want 8/7", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wurm, "trample") || !effectstest.HasEffectiveKeyword(t, g, wurm, "haste") {
		t.Error("a creature cast for its warp cost gains trample and haste")
	}
}

func TestEoeEGFullBoreTargetsOnlyYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Full Bore", "Instant", eoeEGFullBoreOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
}

// --- Fungal Colossus ---------------------------------------------------------

func TestEoeEGFungalColossusCostsLessPerDifferentlyNamedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, eoeEGColossusOracle, "Creature — Fungus Beast", "{6}{G}"); got != 7 {
		t.Errorf("with no lands: %d, want 7", got)
	}
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Mosswort Bridge", "Land")
	effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	if got := effectstest.SelfPricedMV(t, g, me, eoeEGColossusOracle, "Creature — Fungus Beast", "{6}{G}"); got != 4 {
		t.Errorf("with three differently named lands: %d, want 4", got)
	}
}

// --- Genemorph Imago ---------------------------------------------------------

func TestEoeEGGenemorphImagoSetsTheTargetToThreeThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imago := effectstest.B12Push(g, me.ID, "Genemorph Imago", "Creature — Insect Druid", eoeEGImagoOracle, 1, 3)
	effectstest.AssertKeywords(t, g, imago, "flying")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)
	effectstest.PlayLandFromHand(t, g, "Island", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want base 3/3", p, tough)
	}
}

func TestEoeEGGenemorphImagoSetsSixSixWithSixLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Genemorph Imago", "Creature — Insect Druid", eoeEGImagoOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.PlayLandFromHand(t, g, "Island", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("with six lands the target is %d/%d, want base 6/6", p, tough)
	}
}

// --- Gene Pollinator ---------------------------------------------------------

func TestEoeEGGenePollinatorTapsAnotherPermanentForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pollinator := effectstest.B12Push(g, me.ID, "Gene Pollinator", "Artifact Creature — Robot Insect", eoeEGPollinatorOracle, 1, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	for name, ids := range map[string][]uuid.UUID{
		"nothing else tapped":    nil,
		"itself":                 {pollinator},
		"an opponent's artifact": {theirs},
	} {
		if err := g.ActivateManaAbility(me.ID, pollinator, 0, game.ManaAbilityParams{TapIDs: ids}); err == nil {
			t.Fatalf("%s paid the cost", name)
		}
	}
	effectstest.ActivateManaFor(t, g, me.ID, pollinator, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{rock}})
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	effectstest.AssertTapped(t, g, true, pollinator, rock)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %v, want one mana", me.ManaPool)
	}
}

// --- Gigastorm Titan ---------------------------------------------------------

func TestEoeEGGigastormTitanCostsThreeLessAfterAnotherSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, eoeEGTitanOracle, "Creature — Elemental", "{4}{U}"); got != 5 {
		t.Errorf("with no spell cast: %d, want 5", got)
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	if got := effectstest.SelfPricedMV(t, g, me, eoeEGTitanOracle, "Creature — Elemental", "{4}{U}"); got != 2 {
		t.Errorf("after another spell: %d, want 2", got)
	}
}

// --- Glacier Godmaw ----------------------------------------------------------

func TestEoeEGGlacierGodmawMakesALanderAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	godmaw := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Glacier Godmaw", TypeLine: "Creature — Leviathan", ManaCost: "{5}{G}{G}",
		OracleID: eoeEGGodmawOracle, Power: 6, Toughness: 6}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(godmaw) {
		t.Fatal("the Godmaw did not enter")
	}
	effectstest.AssertKeywords(t, g, godmaw, "trample")
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1", got)
	}
}

func TestEoeEGGlacierGodmawPumpsYourTeamOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	godmaw := effectstest.B12Push(g, me.ID, "Glacier Godmaw", "Creature — Leviathan", eoeEGGodmawOracle, 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{godmaw: 7, bear: 3} {
		if p, tough := effectstest.EoeCDPT(t, g, id); p != want || tough != want {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tough, want, want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s gains vigilance and haste", id)
		}
	}
	if p, _ := effectstest.EoeCDPT(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature is untouched")
	}
}

// --- Gravblade Heavy ---------------------------------------------------------

func TestEoeEGGravbladeHeavyGrowsAndGainsDeathtouchWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heavy := effectstest.B12Push(g, me.ID, "Gravblade Heavy", "Creature — Human Soldier", eoeEGHeavyOracle, 3, 4)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, tough := effectstest.EoeCDPT(t, g, heavy); p != 3 || tough != 4 || effectstest.HasEffectiveKeyword(t, g, heavy, "deathtouch") {
		t.Errorf("with no artifact of yours it is a plain 3/4, got %d/%d", p, tough)
	}
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if p, tough := effectstest.EoeCDPT(t, g, heavy); p != 4 || tough != 4 || !effectstest.HasEffectiveKeyword(t, g, heavy, "deathtouch") {
		t.Errorf("with an artifact it is a 4/4 with deathtouch, got %d/%d", p, tough)
	}
	effectstest.B15Destroy(t, g, rock)
	if p, _ := effectstest.EoeCDPT(t, g, heavy); p != 3 || effectstest.HasEffectiveKeyword(t, g, heavy, "deathtouch") {
		t.Error("both end when the last artifact leaves")
	}
}
