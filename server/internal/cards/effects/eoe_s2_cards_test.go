package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_s2_cards_test.go — Edge of Eternities cards whose names start with
// S, the second batch (#4461). Each test plays the real catalog entry and
// asserts what its text says, and what it refuses.

const (
	eoeS2SamiOracle        = "470c16c5-b078-47f5-89b0-42991e1f7157"
	eoeS2ScoutOracle       = "5a367c9c-9a7f-4587-aa05-c0fba46ed838"
	eoeS2ScroungeOracle    = "71982b17-1d9a-4f98-b8b5-d232e3a8cf1d"
	eoeS2SeamRipOracle     = "dfd5939e-a71a-4e16-9b93-752162be4ccd"
	eoeS2BroodtenderOracle = "15c994b9-901a-46d6-804b-24560a14a481"
	eoeS2SlagdrillOracle   = "7f1edf4c-46a1-426d-8ce9-757a4c27367d"
	eoeS2SledgeOracle      = "bcd086ce-de5c-4b34-92a1-ece22055e479"
	eoeS2FreighterOracle   = "fb84b94c-dc8c-4182-9b96-abd092b20981"
	eoeS2LightbladeOracle  = "0ea2a099-7c81-4012-8148-4d377f2d64e1"
	eoeS2ShepherdOracle    = "42fa99ab-d35f-4439-ab00-fe99bad626ac"
	eoeS2SecurityOracle    = "5454dddf-c6e5-4b0d-b96d-4904c13606dc"
	eoeS2ChaplainOracle    = "de6b26e4-8700-490f-9df2-dac430463274"
	eoeS2LabshipOracle     = "a37314cf-bbbd-4003-b94d-0e6e60edfc15"
)

func TestEoeS2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeS2SamiOracle:        "Sami, Ship's Engineer",
		eoeS2ScoutOracle:       "Scout for Survivors",
		eoeS2ScroungeOracle:    "Scrounge for Eternity",
		eoeS2SeamRipOracle:     "Seam Rip",
		eoeS2BroodtenderOracle: "Seedship Broodtender",
		eoeS2SlagdrillOracle:   "Slagdrill Scrapper",
		eoeS2SledgeOracle:      "Sledge-Class Seedship",
		eoeS2FreighterOracle:   "Specimen Freighter",
		eoeS2LightbladeOracle:  "Squire's Lightblade",
		eoeS2ShepherdOracle:    "Starfield Shepherd",
		eoeS2SecurityOracle:    "Starport Security",
		eoeS2ChaplainOracle:    "Sunstar Chaplain",
		eoeS2LabshipOracle:     "Synthesizer Labship",
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
	if ac := game.AlternativeCostByKey(eoeS2ShepherdOracle, "warp"); ac == nil || ac.ManaCost != "{1}{W}" {
		t.Errorf("Starfield Shepherd: warp %+v, want {1}{W}", ac)
	}
}

// --- Sami, Ship's Engineer ----------------------------------------------------

func TestEoeS2SamiMakesATappedRobotWithTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sami := b12Push(g, me.ID, "Sami, Ship's Engineer", "Legendary Creature — Human Artificer", eoeS2SamiOracle, 2, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{sami, bear} {
		eoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	robot := avB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token with two tapped creatures (Sami counts)")
	}
	c := layeredCard(t, g, robot)
	if !c.Tapped || !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 || len(c.Colors) != 0 {
		t.Errorf("want a tapped 2/2 colorless artifact creature, got tapped=%v %d/%d %v", c.Tapped, c.CurrentPower(), c.CurrentToughness(), c.Colors)
	}
}

func TestEoeS2SamiNeedsTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sami := b12Push(g, me.ID, "Sami, Ship's Engineer", "Legendary Creature — Human Artificer", eoeS2SamiOracle, 2, 4)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, sami, func(c *game.Card) { c.Tapped = true })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := eoeCDCount(g, me.ID, "Robot"); got != 0 {
		t.Errorf("one tapped creature made %d Robots", got)
	}
}

// --- Scout for Survivors ------------------------------------------------------

func TestEoeS2ScoutForSurvivorsReturnsCreaturesWithinTheTotalWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := b38Graveyard(g, me.ID, "One", "Creature — Human", "{W}", 1, 1)
	two := b38Graveyard(g, me.ID, "Two", "Creature — Human", "{1}{W}", 2, 2)
	three := b38Graveyard(g, me.ID, "Three", "Creature — Human", "{2}{W}", 3, 3)
	castCatalogSpell(t, g, "Scout for Survivors", "Sorcery", eoeS2ScoutOracle, cardRefs(one, two))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return", id)
			continue
		}
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", id, n)
		}
	}
	if !me.Graveyard.Contains(three) {
		t.Error("the untargeted card stays in the graveyard")
	}
}

func TestEoeS2ScoutForSurvivorsRefusesATotalAboveThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := b38Graveyard(g, me.ID, "Two", "Creature — Human", "{1}{W}", 2, 2)
	three := b38Graveyard(g, me.ID, "Three", "Creature — Human", "{2}{W}", 3, 3)
	theirs := b38Graveyard(g, opp.ID, "Theirs", "Creature — Human", "{W}", 1, 1)
	if err := castCatalogSpellErr(t, g, "Scout for Survivors", "Sorcery", eoeS2ScoutOracle, cardRefs(two, three)); err == nil {
		t.Error("two cards with total mana value 5 were accepted")
	}
	if err := castCatalogSpellErr(t, g, "Scout for Survivors", "Sorcery", eoeS2ScoutOracle, cardRefs(theirs)); err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
}

// --- Scrounge for Eternity ----------------------------------------------------

func TestEoeS2ScroungeForEternityReanimatesASpacecraftAndMakesALander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	ship := b38Graveyard(g, me.ID, "Dead Ship", "Artifact — Spacecraft", "{3}{U}{U}", 4, 4)
	avB1Cast(t, g, me, game.Card{Name: "Scrounge for Eternity", TypeLine: "Sorcery", ManaCost: "{2}{B}", OracleID: eoeS2ScroungeOracle},
		game.CastSpellParams{Targets: cardRefs(ship), SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is paid as the spell is cast")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ship) {
		t.Error("the mana value 5 Spacecraft card returned")
	}
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1", got)
	}
}

func TestEoeS2ScroungeForEternityRefusesAnExpensiveCardAndNoSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	wurm := b38Graveyard(g, me.ID, "Dead Wurm", "Creature — Wurm", "{4}{G}{G}", 6, 6)
	bear := b38Graveyard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	for name, params := range map[string]game.CastSpellParams{
		"a mana value 6 target": {Targets: cardRefs(wurm), SacrificeIDs: []uuid.UUID{fodder}},
		"a land as the cost":    {Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{land}},
		"no sacrifice":          {Targets: cardRefs(bear)},
	} {
		id := avB1Hand(me, game.Card{Name: "Scrounge for Eternity", TypeLine: "Sorcery", ManaCost: "{2}{B}", OracleID: eoeS2ScroungeOracle})
		if err := g.CastSpell(me.ID, id, params); err == nil {
			t.Errorf("cast with %s was accepted", name)
		}
	}
}

// --- Seam Rip -----------------------------------------------------------------

func TestEoeS2SeamRipExilesACheapOpposingPermanentUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := paPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	ogre := paPermanent(g, opp.ID, "Their Ogre", "Creature — Ogre", "{2}{R}", 3, 3)
	mine := paPermanent(g, me.ID, "My Rock", "Artifact", "{1}", 0, 0)
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	rip := castCatalogSpell(t, g, "Seam Rip", "Enchantment", eoeS2SeamRipOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{"a mana value 3 creature": ogre, "your own artifact": mine, "a land": land} {
		if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(id)); err == nil {
			t.Fatalf("Seam Rip targeted %s", name)
		}
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Fatal("the mana value 2 artifact was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rip) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Rock", opp.ID); n != 1 {
		t.Errorf("after Seam Rip left, %d of the exiled artifact under its owner's control, want 1", n)
	}
}

// --- Seedship Broodtender -----------------------------------------------------

func TestEoeS2SeedshipBroodtenderMillsThreeAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E")
	library, yard := me.Library.Size(), me.Graveyard.Size()
	avB1Cast(t, g, me, game.Card{Name: "Seedship Broodtender", TypeLine: "Creature — Insect Citizen", ManaCost: "{B}{G}",
		OracleID: eoeS2BroodtenderOracle, Power: 2, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-3 || me.Graveyard.Size() != yard+3 {
		t.Errorf("library %d -> %d, graveyard %d -> %d, want three milled", library, me.Library.Size(), yard, me.Graveyard.Size())
	}
}

func TestEoeS2SeedshipBroodtenderSacrificesToReanimateAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brood := b12Push(g, me.ID, "Seedship Broodtender", "Creature — Insect Citizen", eoeS2BroodtenderOracle, 2, 3)
	ship := b38Graveyard(g, me.ID, "Dead Ship", "Artifact — Spacecraft", "{6}", 6, 6)
	relic := b38Graveyard(g, me.ID, "Dead Relic", "Artifact", "{1}", 0, 0)
	theirs := b38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	advanceToMain(t, g)
	for name, id := range map[string]uuid.UUID{"an opponent's creature card": theirs, "a non-Spacecraft artifact card": relic} {
		if err := g.ActivateCatalogAbility(me.ID, brood, 0, game.ActivateAbilityParams{Targets: cardRefs(id)}); err == nil {
			t.Fatalf("the Broodtender targeted %s", name)
		}
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, brood, 0, game.ActivateAbilityParams{Targets: cardRefs(ship)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	b16Activate(t, g, me.ID, brood, 0, game.ActivateAbilityParams{Targets: cardRefs(ship)})
	if g.Battlefield.Contains(brood) {
		t.Error("sacrificing itself is part of the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ship) {
		t.Error("the Spacecraft card returned")
	}
}

// --- Slagdrill Scrapper -------------------------------------------------------

func TestEoeS2SlagdrillScrapperSacrificesAnotherArtifactOrLandToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	scrapper := b12Push(g, me.ID, "Slagdrill Scrapper", "Artifact Creature — Robot Scout", eoeS2SlagdrillOracle, 1, 2)
	land := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	for name, id := range map[string]uuid.UUID{"a creature": bear, "itself": scrapper} {
		if err := g.ActivateCatalogAbility(me.ID, scrapper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("%s paid the another-artifact-or-land sacrifice", name)
		}
	}
	hand := handSize(me)
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, scrapper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if !b16Tapped(t, g, scrapper) {
		t.Error("the {T} was not paid")
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d after the draw, want %d", got, hand+1)
	}
}

// --- Sledge-Class Seedship ----------------------------------------------------

func TestEoeS2SledgeClassSeedshipPutsACreatureFromHandWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Push(g, me.ID, "Sledge-Class Seedship", "Artifact — Spacecraft", eoeS2SledgeOracle, 4, 5)
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("below 7 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 7); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at7 := layeredCard(t, g, ship)
	if !at7.IsCreature() || !game.HasKeyword(&at7, "flying") || at7.CurrentPower() != 4 || at7.CurrentToughness() != 5 {
		t.Errorf("at 7+ it is a 4/5 flier, got %d/%d", at7.CurrentPower(), at7.CurrentToughness())
	}
	fatty := avB1Hand(me, game.Card{Name: "Big Fatty", TypeLine: "Creature — Wurm", ManaCost: "{8}", Power: 8, Toughness: 8})
	shock := avB1Hand(me, game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ship, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt opens")
	}
	if !hasID(c.ChooseCards, fatty) || hasID(c.ChooseCards, shock) || c.ChooseMin != 0 {
		t.Errorf("want an optional pick of the creature card only: %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fatty}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(fatty) || me.Hand.Contains(fatty) {
		t.Error("the creature arrives on the battlefield")
	}
}

// --- Specimen Freighter -------------------------------------------------------

func TestEoeS2SpecimenFreighterReturnsTwoNonSpacecraftCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	craft := b12Creature(g, opp.ID, "Their Craft", "Artifact Creature — Spacecraft", 5, 5)
	avB1Cast(t, g, me, game.Card{Name: "Specimen Freighter", TypeLine: "Artifact — Spacecraft", ManaCost: "{5}{U}",
		OracleID: eoeS2FreighterOracle, Power: 4, Toughness: 7}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(craft)); err == nil {
		t.Fatal("a Spacecraft creature is not a legal target")
	}
	if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(theirs, mine)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) || !me.Hand.Contains(mine) {
		t.Error("both creatures return to their owners' hands")
	}
	if !g.Battlefield.Contains(craft) {
		t.Error("the Spacecraft stays")
	}
}

func TestEoeS2SpecimenFreighterMillsTheDefenderFourAndFliesAtNine(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Push(g, me.ID, "Specimen Freighter", "Artifact — Spacecraft", eoeS2FreighterOracle, 4, 7)
	if err := g.AddCounter(ship, game.CounterCharge, 8); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("below 9 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at9 := layeredCard(t, g, ship)
	if !at9.IsCreature() || !game.HasKeyword(&at9, "flying") || at9.CurrentPower() != 4 || at9.CurrentToughness() != 7 {
		t.Errorf("at 9+ it is a 4/7 flier, got %d/%d", at9.CurrentPower(), at9.CurrentToughness())
	}
	seedLibrary(opp, "A", "B", "C", "D", "E", "F")
	library, mine := opp.Library.Size(), me.Library.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ship, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != library-4 {
		t.Errorf("defending player's library %d -> %d, want four milled", library, got)
	}
	if got := me.Library.Size(); got != mine {
		t.Errorf("the attacker milled: library %d -> %d", mine, got)
	}
}

// --- Squire's Lightblade ------------------------------------------------------

func TestEoeS2SquiresLightbladeAttachesAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blade := avB1Cast(t, g, me, game.Card{Name: "Squire's Lightblade", TypeLine: "Artifact — Equipment", ManaCost: "{W}",
		OracleID: eoeS2LightbladeOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature is offered")
	}
	pickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, blade).AttachedTo; got.Kind != game.TargetCard || got.ID != bear {
		t.Fatalf("the Lightblade is attached to %+v, want the Bear", got)
	}
	if p, tough := eoeCDPT(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear gains first strike")
	}
	aangAdvanceToMain(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike ends with the turn")
	}
	if p, _ := eoeCDPT(t, g, bear); p != 3 {
		t.Errorf("the +1/+0 stays while equipped: power %d", p)
	}
}

// --- Starfield Shepherd -------------------------------------------------------

func TestEoeS2StarfieldShepherdFindsAPlainsOrAOneDropAndWarpsOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Temple", TypeLine: "Land — Plains"},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	shepherd := avB1Cast(t, g, me, game.Card{Name: "Starfield Shepherd", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}",
		OracleID: eoeS2ShepherdOracle, Power: 3, Toughness: 2}, game.CastSpellParams{AlternativeCost: "warp"})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, shepherd, "flying") {
		t.Error("Starfield Shepherd has flying")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, name := range []string{"Mountain", "Temple", "Bear"} {
		if searchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is offered", name)
		}
	}
	if searchOptionNamed(g, c, "Plains") == uuid.Nil {
		t.Error("the basic Plains is not offered")
	}
	elf := searchOptionNamed(g, c, "Elf")
	answerSearchNamed(t, g, me.ID, "Elf")
	if !me.Hand.Contains(elf) {
		t.Error("the one-drop creature card went to the hand")
	}
	advanceThroughEndStep(t, g)
	if !g.Exile.Contains(shepherd) {
		t.Error("the warped Shepherd is exiled at the end step")
	}
}

// --- Starport Security --------------------------------------------------------

func TestEoeS2StarportSecurityTapsAnotherCreatureAtFullPrice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	security := b12Push(g, me.ID, "Starport Security", "Artifact Creature — Robot Soldier", eoeS2SecurityOracle, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, security, 0, game.ActivateAbilityParams{Targets: cardRefs(security)}); err == nil {
		t.Fatal("it targeted itself")
	}
	floatForTest(g, me, "WC")
	if err := g.ActivateCatalogAbility(me.ID, security, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("{1}{W} paid the full {3}{W} with no countered creature")
	}
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, security, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the target is tapped")
	}
}

func TestEoeS2StarportSecurityCostsTwoLessWithACounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	security := b12Push(g, me.ID, "Starport Security", "Artifact Creature — Robot Soldier", eoeS2SecurityOracle, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(security, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	advanceToMain(t, g)
	floatForTest(g, me, "WC")
	b16Activate(t, g, me.ID, security, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("with its own +1/+1 counter it costs {1}{W} and taps the target")
	}
}

// --- Sunstar Chaplain ---------------------------------------------------------

func TestEoeS2SunstarChaplainCountersATargetWithTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chaplain := b12Push(g, me.ID, "Sunstar Chaplain", "Creature — Human Cleric", eoeS2ChaplainOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{chaplain, bear} {
		eoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt with two tapped creatures")
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature is offered")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
}

func TestEoeS2SunstarChaplainNeedsTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chaplain := b12Push(g, me.ID, "Sunstar Chaplain", "Creature — Human Cleric", eoeS2ChaplainOracle, 3, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, chaplain, func(c *game.Card) { c.Tapped = true })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, chaplain) != nil {
		t.Error("one tapped creature triggered it")
	}
}

func TestEoeS2SunstarChaplainRemovesACounterToTapAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chaplain := b12Push(g, me.ID, "Sunstar Chaplain", "Creature — Human Cleric", eoeS2ChaplainOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirBear, game.CounterPlusOne, 1)
	})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chaplain, 0, game.ActivateAbilityParams{
		Targets: cardRefs(relic), CounterSourceIDs: []uuid.UUID{theirBear},
	}); err == nil {
		t.Fatal("a counter on an opponent's creature paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, chaplain, 0, game.ActivateAbilityParams{
		Targets: cardRefs(land), CounterSourceIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Fatal("a land is not an artifact or creature")
	}
	b16Activate(t, g, me.ID, chaplain, 0, game.ActivateAbilityParams{Targets: cardRefs(relic), CounterSourceIDs: []uuid.UUID{bear}})
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the cost left %d counters on the Bear", n)
	}
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, relic) {
		t.Error("the artifact is tapped")
	}
}

// --- Synthesizer Labship ------------------------------------------------------

func TestEoeS2SynthesizerLabshipAnimatesAnotherArtifactAtTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := b12Push(g, me.ID, "Synthesizer Labship", "Artifact — Spacecraft", eoeS2LabshipOracle, 4, 4)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	if err := g.AddCounter(ship, game.CounterCharge, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at 2+")
	}
	if hasID(p.PickTargetCards, ship) {
		t.Error("the Labship is offered for its own \"other\" trigger")
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, relic)
	if !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !game.HasKeyword(&c, "flying") {
		t.Errorf("want a 2/2 artifact creature with flying, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	aangAdvanceToMain(t, g, 1)
	if layeredCard(t, g, relic).IsCreature() {
		t.Error("the animation ends with the turn")
	}
}

func TestEoeS2SynthesizerLabshipDoesNothingBelowTwoAndFliesAtNine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := b12Push(g, me.ID, "Synthesizer Labship", "Artifact — Spacecraft", eoeS2LabshipOracle, 4, 4)
	b12Permanent(g, me.ID, "Relic", "Artifact")
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the 2+ trigger fired at one charge counter")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 8); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at9 := layeredCard(t, g, ship)
	if !at9.IsCreature() || at9.CurrentPower() != 4 || at9.CurrentToughness() != 4 {
		t.Errorf("at 9+ it is a 4/4 creature, got %d/%d", at9.CurrentPower(), at9.CurrentToughness())
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !game.HasKeyword(&at9, kw) {
			t.Errorf("at 9+ it has %s", kw)
		}
	}
}
