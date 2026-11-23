package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b39_cards_test.go — Secrets of Strixhaven (sos), batch 39 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB39ViciousRivalryOracle = "90898c76-21f9-4597-88fe-dc9179911cde"
	socB39DanceOracle          = "c9f35490-462a-4c8f-9d8b-d8abca6596f7"
	socB39WanderOffOracle      = "b14d3226-adc2-4bdd-9e69-fefdc946ac04"
	socB39HypothesisOracle     = "2e56507d-af77-47c8-a29a-410cf14cfe20"
	socB39CharmOracle          = "ee1db746-f5c1-4dce-82eb-24b273cc9507"
	socB39LorecasterOracle     = "f82265fd-e772-4347-a27f-37b00ccd416b"
	socB39ExperimentOracle     = "5faa9e54-1cc4-48c0-bbfa-a495087c36d6"
	socB39BearType             = "Creature — Bear"
)

func TestSocB39EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB39ViciousRivalryOracle: "Vicious Rivalry",
		socB39DanceOracle:          "Visionary's Dance",
		socB39WanderOffOracle:      "Wander Off",
		socB39HypothesisOracle:     "Wild Hypothesis",
		socB39CharmOracle:          "Witherbloom Charm",
		socB39LorecasterOracle:     "Zealous Lorecaster",
		socB39ExperimentOracle:     "Zimone's Experiment",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Vicious Rivalry ------------------------------------------------------------

func TestSocB39ViciousRivalryPaysXLifeAndDestroysArtifactsAndCreaturesUpToX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheapRock := effectstest.PushTypedCard(g, opp.ID, "Cheap Rock", "Artifact", "{2}")
	dearRock := effectstest.PushTypedCard(g, opp.ID, "Dear Rock", "Artifact", "{3}")
	cheapBear := effectstest.PushTypedCard(g, opp.ID, "Cheap Bear", socB39BearType, "{1}{G}")
	dearBear := effectstest.PushTypedCard(g, opp.ID, "Dear Bear", socB39BearType, "{2}{G}")
	mine := effectstest.PushTypedCard(g, me.ID, "My Cub", socB39BearType, "{1}")
	aura := effectstest.PushTypedCard(g, opp.ID, "Cheap Charm", "Enchantment", "{1}")
	land := effectstest.PushTypedCard(g, opp.ID, "Their Land", "Land", "")
	life := me.Life

	effectstest.CastWipe(t, g, "Vicious Rivalry", "Sorcery", socB39ViciousRivalryOracle, game.CastSpellParams{XValue: 2})
	if me.Life != life-2 {
		t.Fatalf("life %d -> %d at cast, want X = 2 paid as an additional cost", life, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Cheap Rock": cheapRock, "Cheap Bear": cheapBear, "My Cub": mine} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s (mana value 2 or less) survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"Dear Rock": dearRock, "Dear Bear": dearBear, "Cheap Charm": aura, "Their Land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed but is not an artifact or creature of mana value 2 or less", name)
		}
	}
}

func TestSocB39ViciousRivalryAtXZeroDestroysOnlyManaValueZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	food := effectstest.PushToken(g, opp.ID, effects.FoodToken())
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", socB39BearType, "{1}{G}")
	life := me.Life

	effectstest.CastWipe(t, g, "Vicious Rivalry", "Sorcery", socB39ViciousRivalryOracle, game.CastSpellParams{XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != life {
		t.Errorf("life %d -> %d, want no payment at X = 0", life, me.Life)
	}
	if g.Battlefield.Contains(food) {
		t.Error("a token has mana value 0, so X = 0 still destroys it")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a mana value 2 creature died to X = 0")
	}
}

func TestSocB39ViciousRivalryRefusesMoreLifeThanYouHave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vicious Rivalry", TypeLine: "Sorcery", OracleID: socB39ViciousRivalryOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: me.Life + 1, Strict: true}); err == nil {
		t.Fatal("paying more life than you have was allowed")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast still moved the card off the hand")
	}
}

// --- Visionary's Dance ----------------------------------------------------------------

func TestSocB39VisionarysDanceMakesTwoFlyingElementals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental")
	effectstest.CastCatalogSpell(t, g, "Visionary's Dance", "Sorcery", socB39DanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental") - before; got != 2 {
		t.Fatalf("%d Elementals made, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Elemental" || c.Controller != me.ID {
			continue
		}
		if !effects.IsToken(c) || c.Power != 3 || c.Toughness != 3 || !game.HasKeyword(&c, "flying") || !c.HasColor("U") || !c.HasColor("R") {
			t.Errorf("Elemental token is %d/%d %v colors %v, want a 3/3 blue and red flier", c.Power, c.Toughness, c.Keywords, c.Colors)
		}
	}
}

func TestSocB39VisionarysDanceDiscardAbilityTakesOneOfTheTopTwoAndBinsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lower := effectstest.TopLibraryCard(me, "Lower", "Sorcery")
	upper := effectstest.TopLibraryCard(me, "Upper", "Sorcery")
	id := effectstest.PushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("discarding the card is a cost, paid on activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(lower) || me.Hand.Size() != hand+1 {
		t.Error("the chosen card is not the one new card in hand")
	}
	if !me.Graveyard.Contains(upper) {
		t.Error("the other of the top two should be put into the graveyard")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d -> %d, want exactly the top two gone", lib, me.Library.Size())
	}
}

func TestSocB39VisionarysDanceAbilityNeedsTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused activation still discarded the card")
	}
}

func TestSocB39VisionarysDanceAbilityWithASingleCardLibraryTakesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	only := effectstest.TopLibraryCard(me, "Only", "Sorcery")
	id := effectstest.PushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(only) {
		t.Error("the lone library card should go to the hand")
	}
}

// --- Wander Off ----------------------------------------------------------------------------

func TestSocB39WanderOffExilesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB39BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Wander Off", "Instant", socB39WanderOffOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the creature should be exiled, not destroyed")
	}
	if !effectstest.InExile(g, bear) {
		t.Error("the creature is not in exile")
	}
}

func TestSocB39WanderOffRefusesANoncreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.PushTypedCard(g, g.Seats[1].ID, "Rock", "Artifact", "{1}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Wander Off", "Instant", socB39WanderOffOracle, effectstest.CardRefs(rock)); err == nil {
		t.Error("an artifact was accepted as a creature target")
	}
}

func TestSocB39WanderOffDoesNothingWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB39BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Wander Off", "Instant", socB39WanderOffOracle, effectstest.CardRefs(bear))
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, bear); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) || effectstest.InExile(g, bear) {
		t.Error("a target that left should not be exiled from its new zone")
	}
}

// --- Wild Hypothesis ---------------------------------------------------------------------

func socB39CastHypothesis(t *testing.T, g *game.Game, me *game.Player, x int) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Wild Hypothesis", TypeLine: "Sorcery", ManaCost: "{X}{G}", OracleID: socB39HypothesisOracle},
		"G", x+1, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast Wild Hypothesis for X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func socB39Fractal(g *game.Game, owner uuid.UUID) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Fractal" && c.Controller == owner && effects.IsToken(c) {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

func TestSocB39WildHypothesisMakesAFractalWithXCountersThenSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lower := effectstest.TopLibraryCard(me, "Lower", "Sorcery")
	upper := effectstest.TopLibraryCard(me, "Upper", "Sorcery")
	socB39CastHypothesis(t, g, me, 3)

	fractal, ok := socB39Fractal(g, me.ID)
	if !ok {
		t.Fatal("no Fractal token")
	}
	if n := effectstest.CountersOn(g, fractal, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters %d, want X = 3", n)
	}
	if p := effects.B31CurrentPowerOf(g, fractal); p != 3 {
		t.Errorf("Fractal power %d, want 0 plus three counters", p)
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	// Bin the top card and keep the second on top.
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{upper}, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(upper) {
		t.Error("the binned card is not in the graveyard")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != lower {
		t.Error("the kept card should be back on top of the library")
	}
}

func TestSocB39WildHypothesisAtXZeroFractalDiesButStillSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lower := effectstest.TopLibraryCard(me, "Lower", "Sorcery")
	upper := effectstest.TopLibraryCard(me, "Upper", "Sorcery")
	socB39CastHypothesis(t, g, me, 0)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the surveil is a separate sentence and still happens at X = 0")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, []uuid.UUID{upper, lower}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	g.RunStateChecksForTest()
	if _, ok := socB39Fractal(g, me.ID); ok {
		t.Error("a 0/0 Fractal stayed on the battlefield")
	}
}

// --- Witherbloom Charm ---------------------------------------------------------------------

func TestSocB39CharmSacrificesAPermanentAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushTypedCard(g, me.ID, "My Land", "Land", "")
	effectstest.CastModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, land) {
		t.Fatal("the permanent is chosen as the spell resolves, with a floor of zero for the \"may\"")
	}
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, me.Hand.Size())
	}
}

func TestSocB39CharmDecliningTheSacrificeDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushTypedCard(g, me.ID, "My Land", "Land", "")
	effectstest.CastModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("a declined sacrifice still sacrificed")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: \"if you do\" means no draw without a sacrifice", hand, me.Hand.Size())
	}
}

func TestSocB39CharmWithNoPermanentAsksNothingAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("a prompt was opened with nothing to sacrifice")
	}
	// The charm is on the stack while it resolves, so it cannot be sacrificed to itself.
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

func TestSocB39CharmGainsFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
}

func TestSocB39CharmDestroysANonlandPermanentOfManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	cheap := effectstest.PushTypedCard(g, opp.ID, "Cheap Rock", "Artifact", "{2}")
	effectstest.CastModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{2}, effectstest.CardRefs(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the target survived")
	}
}

func TestSocB39CharmRefusesAnExpensiveOrLandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	dear := effectstest.PushTypedCard(g, opp.ID, "Dear Rock", "Artifact", "{3}")
	land := effectstest.PushTypedCard(g, opp.ID, "Their Land", "Land", "")
	for _, bad := range []uuid.UUID{dear, land} {
		if _, err := effectstest.CastModalErr(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{2}, effectstest.CardRefs(bad)); err == nil {
			t.Errorf("target %v was accepted", bad)
		}
	}
}

// --- Zealous Lorecaster ----------------------------------------------------------------------

func TestSocB39LorecasterReturnsAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", socB39BearType, "{1}{G}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Bolt", "Sorcery", "{R}", "")

	id := effectstest.CastAndResolveCreature(t, g, "Zealous Lorecaster", "Creature — Giant Sorcerer", socB39LorecasterOracle)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target pick")
	}
	if !effectstest.HasID(prompt.PickTargetCards, bolt) || effectstest.HasID(prompt.PickTargetCards, bear) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v, want only my instant or sorcery", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("trigger not on the stack after the pick")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the chosen card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("another card left a graveyard")
	}
}

func TestSocB39LorecasterWithNoInstantOrSorceryAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.GraveyardCardOf(g, me, "Dead Bear", socB39BearType, "{1}{G}", "")
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Zealous Lorecaster", "Creature — Giant Sorcerer", socB39LorecasterOracle)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for with none legal (CR 603.3d)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want nothing returned", hand, me.Hand.Size())
	}
}

// --- Zimone's Experiment ----------------------------------------------------------------------

// socB39Five stacks the five looked-at cards and returns them top first:
// a land, a creature, a creature, a sorcery and a land.
func socB39Five(me *game.Player) []uuid.UUID {
	e := effectstest.TopLibraryCard(me, "Fifth Land", "Basic Land — Forest")
	d := effectstest.TopLibraryCard(me, "Fourth Sorcery", "Sorcery")
	c := effectstest.TopLibraryCard(me, "Third Creature", socB39BearType)
	b := effectstest.TopLibraryCard(me, "Second Creature", socB39BearType)
	a := effectstest.TopLibraryCard(me, "First Land", "Basic Land — Forest")
	return []uuid.UUID{a, b, c, d, e}
}

func socB39CastExperiment(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Zimone's Experiment", "Sorcery", socB39ExperimentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB39ExperimentOffersOnlyCreatureAndLandCardsOfTheFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if len(prompt.ChooseCards) != 4 || effectstest.HasID(prompt.ChooseCards, five[3]) {
		t.Errorf("offered %v, want the two lands and two creatures but not the sorcery", prompt.ChooseCards)
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 2 {
		t.Errorf("bounds [%d,%d], want \"up to two\"", prompt.ChooseMin, prompt.ChooseMax)
	}
	if effectstest.LatestChooseCardsFor(g, opp.ID) != nil {
		t.Error("the look is private: no other seat is asked")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[1], five[2]}); err == nil {
		t.Error("three cards were accepted for \"up to two\"")
	}
}

func TestSocB39ExperimentPutsTheLandOntoTheBattlefieldTappedAndTheCreatureIntoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	lib := me.Library.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	land := effectstest.FindBattlefieldCardByID(g, five[0])
	if land == nil || !land.Tapped {
		t.Error("the revealed land should be on the battlefield tapped")
	}
	if !me.Hand.Contains(five[1]) || me.Hand.Size() != hand+1 {
		t.Error("the revealed creature card should be the one new card in hand")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d -> %d, want only the two revealed cards gone", lib, me.Library.Size())
	}
	// Everything else is at the bottom, in some order, and none stays on top.
	if bottom := effectstest.LibraryBottom(me, 3); !effectstest.SameSet(bottom, []uuid.UUID{five[2], five[3], five[4]}) {
		t.Errorf("bottom three %v, want the unrevealed three", bottom)
	}
	for _, id := range five[:2] {
		if !effectstest.HornRevealed(g, id) {
			t.Errorf("revealed card %v has no reveal event", id)
		}
	}
	for _, id := range five[2:] {
		if effectstest.HornRevealed(g, id) {
			t.Errorf("card %v was only looked at but was revealed", id)
		}
	}
}

func TestSocB39ExperimentTwoLandsBothEnterTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[4]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range []uuid.UUID{five[0], five[4]} {
		if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || !c.Tapped {
			t.Errorf("land %v is not on the battlefield tapped", id)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no creature taken", hand, me.Hand.Size())
	}
	if bottom := effectstest.LibraryBottom(me, 3); !effectstest.SameSet(bottom, []uuid.UUID{five[1], five[2], five[3]}) {
		t.Errorf("bottom three %v, want the three unrevealed cards", bottom)
	}
}

func TestSocB39ExperimentTwoCreaturesBothGoToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[1], five[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(five[1]) || !me.Hand.Contains(five[2]) || me.Hand.Size() != hand+2 {
		t.Error("both revealed creature cards belong in hand")
	}
	if bottom := effectstest.LibraryBottom(me, 3); !effectstest.SameSet(bottom, []uuid.UUID{five[0], five[3], five[4]}) {
		t.Errorf("bottom three %v, want the three unrevealed cards", bottom)
	}
}

func TestSocB39ExperimentRevealingNothingPutsAllFiveOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the reveal: %v", err)
	}
	if me.Hand.Size() != hand {
		t.Error("a card was taken though none was revealed")
	}
	if bottom := effectstest.LibraryBottom(me, 5); !effectstest.SameSet(bottom, five) {
		t.Errorf("bottom five %v, want all five looked-at cards", bottom)
	}
}

func TestSocB39ExperimentWithNoCreatureOrLandAsksNothingAndBottomsAllFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		five = append(five, effectstest.TopLibraryCard(me, "Spell", "Sorcery"))
	}
	socB39CastExperiment(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("a prompt was opened with nothing to reveal")
	}
	if bottom := effectstest.LibraryBottom(me, 5); !effectstest.SameSet(bottom, five) {
		t.Errorf("bottom five %v, want all five looked-at cards", bottom)
	}
}
