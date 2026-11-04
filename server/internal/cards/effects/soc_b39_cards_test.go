package effects

import (
	"testing"

	"github.com/google/uuid"

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

// --- Vicious Rivalry ------------------------------------------------------------

func TestSocB39ViciousRivalryPaysXLifeAndDestroysArtifactsAndCreaturesUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheapRock := pushTypedCard(g, opp.ID, "Cheap Rock", "Artifact", "{2}")
	dearRock := pushTypedCard(g, opp.ID, "Dear Rock", "Artifact", "{3}")
	cheapBear := pushTypedCard(g, opp.ID, "Cheap Bear", socB39BearType, "{1}{G}")
	dearBear := pushTypedCard(g, opp.ID, "Dear Bear", socB39BearType, "{2}{G}")
	mine := pushTypedCard(g, me.ID, "My Cub", socB39BearType, "{1}")
	aura := pushTypedCard(g, opp.ID, "Cheap Charm", "Enchantment", "{1}")
	land := pushTypedCard(g, opp.ID, "Their Land", "Land", "")
	life := me.Life

	castWipe(t, g, "Vicious Rivalry", "Sorcery", socB39ViciousRivalryOracle, game.CastSpellParams{XValue: 2})
	if me.Life != life-2 {
		t.Fatalf("life %d -> %d at cast, want X = 2 paid as an additional cost", life, me.Life)
	}
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	food := pushToken(g, opp.ID, FoodToken())
	bear := pushTypedCard(g, opp.ID, "Bear", socB39BearType, "{1}{G}")
	life := me.Life

	castWipe(t, g, "Vicious Rivalry", "Sorcery", socB39ViciousRivalryOracle, game.CastSpellParams{XValue: 0})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := countBattlefieldNamed(g, me.ID, "Elemental")
	castCatalogSpell(t, g, "Visionary's Dance", "Sorcery", socB39DanceOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Elemental") - before; got != 2 {
		t.Fatalf("%d Elementals made, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Elemental" || c.Controller != me.ID {
			continue
		}
		if !IsToken(c) || c.Power != 3 || c.Toughness != 3 || !game.HasKeyword(&c, "flying") || !c.HasColor("U") || !c.HasColor("R") {
			t.Errorf("Elemental token is %d/%d %v colors %v, want a 3/3 blue and red flier", c.Power, c.Toughness, c.Keywords, c.Colors)
		}
	}
}

func TestSocB39VisionarysDanceDiscardAbilityTakesOneOfTheTopTwoAndBinsTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	id := pushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
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
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	only := topLibraryCard(me, "Only", "Sorcery")
	id := pushCatalogHandCard(me, "Visionary's Dance", "Sorcery", socB39DanceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(only) {
		t.Error("the lone library card should go to the hand")
	}
}

// --- Wander Off ----------------------------------------------------------------------------

func TestSocB39WanderOffExilesTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", socB39BearType, 2, 2)
	castCatalogSpell(t, g, "Wander Off", "Instant", socB39WanderOffOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the creature should be exiled, not destroyed")
	}
	if !inExile(g, bear) {
		t.Error("the creature is not in exile")
	}
}

func TestSocB39WanderOffRefusesANoncreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	rock := pushTypedCard(g, g.Seats[1].ID, "Rock", "Artifact", "{1}")
	if err := castCatalogSpellErr(t, g, "Wander Off", "Instant", socB39WanderOffOracle, cardRefs(rock)); err == nil {
		t.Error("an artifact was accepted as a creature target")
	}
}

func TestSocB39WanderOffDoesNothingWhenItsTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", socB39BearType, 2, 2)
	castCatalogSpell(t, g, "Wander Off", "Instant", socB39WanderOffOracle, cardRefs(bear))
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, bear); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) || inExile(g, bear) {
		t.Error("a target that left should not be exiled from its new zone")
	}
}

// --- Wild Hypothesis ---------------------------------------------------------------------

func socB39CastHypothesis(t *testing.T, g *game.Game, me *game.Player, x int) {
	t.Helper()
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Wild Hypothesis", TypeLine: "Sorcery", ManaCost: "{X}{G}", OracleID: socB39HypothesisOracle},
		"G", x+1, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast Wild Hypothesis for X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
}

func socB39Fractal(g *game.Game, owner uuid.UUID) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Fractal" && c.Controller == owner && IsToken(c) {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

func TestSocB39WildHypothesisMakesAFractalWithXCountersThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	socB39CastHypothesis(t, g, me, 3)

	fractal, ok := socB39Fractal(g, me.ID)
	if !ok {
		t.Fatal("no Fractal token")
	}
	if n := countersOn(g, fractal, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters %d, want X = 3", n)
	}
	if p := b31CurrentPowerOf(g, fractal); p != 3 {
		t.Errorf("Fractal power %d, want 0 plus three counters", p)
	}
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	socB39CastHypothesis(t, g, me, 0)
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushTypedCard(g, me.ID, "My Land", "Land", "")
	castModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, land) {
		t.Fatal("the permanent is chosen as the spell resolves, with a floor of zero for the \"may\"")
	}
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, me.Hand.Size())
	}
}

func TestSocB39CharmDecliningTheSacrificeDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushTypedCard(g, me.ID, "My Land", "Land", "")
	castModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("a declined sacrifice still sacrificed")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: \"if you do\" means no draw without a sacrifice", hand, me.Hand.Size())
	}
}

func TestSocB39CharmWithNoPermanentAsksNothingAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("a prompt was opened with nothing to sacrifice")
	}
	// The charm is on the stack while it resolves, so it cannot be sacrificed to itself.
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

func TestSocB39CharmGainsFiveLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
}

func TestSocB39CharmDestroysANonlandPermanentOfManaValueTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	cheap := pushTypedCard(g, opp.ID, "Cheap Rock", "Artifact", "{2}")
	castModal(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{2}, cardRefs(cheap))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the target survived")
	}
}

func TestSocB39CharmRefusesAnExpensiveOrLandTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dear := pushTypedCard(g, opp.ID, "Dear Rock", "Artifact", "{3}")
	land := pushTypedCard(g, opp.ID, "Their Land", "Land", "")
	for _, bad := range []uuid.UUID{dear, land} {
		if _, err := castModalErr(t, g, "Witherbloom Charm", "Instant", socB39CharmOracle, []int{2}, cardRefs(bad)); err == nil {
			t.Errorf("target %v was accepted", bad)
		}
	}
}

// --- Zealous Lorecaster ----------------------------------------------------------------------

func TestSocB39LorecasterReturnsAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := graveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	bear := graveyardCardOf(g, me, "Dead Bear", socB39BearType, "{1}{G}", "")
	theirs := graveyardCardOf(g, opp, "Their Bolt", "Sorcery", "{R}", "")

	id := castAndResolveCreature(t, g, "Zealous Lorecaster", "Creature — Giant Sorcerer", socB39LorecasterOracle)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target pick")
	}
	if !hasID(prompt.PickTargetCards, bolt) || hasID(prompt.PickTargetCards, bear) || hasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v, want only my instant or sorcery", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, bolt)
	if triggerOnStack(g, id) == nil {
		t.Fatal("trigger not on the stack after the pick")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the chosen card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("another card left a graveyard")
	}
}

func TestSocB39LorecasterWithNoInstantOrSorceryAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	graveyardCardOf(g, me, "Dead Bear", socB39BearType, "{1}{G}", "")
	hand := me.Hand.Size()
	castAndResolveCreature(t, g, "Zealous Lorecaster", "Creature — Giant Sorcerer", socB39LorecasterOracle)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for with none legal (CR 603.3d)")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want nothing returned", hand, me.Hand.Size())
	}
}

// --- Zimone's Experiment ----------------------------------------------------------------------

// socB39Five stacks the five looked-at cards and returns them top first:
// a land, a creature, a creature, a sorcery and a land.
func socB39Five(me *game.Player) []uuid.UUID {
	e := topLibraryCard(me, "Fifth Land", "Basic Land — Forest")
	d := topLibraryCard(me, "Fourth Sorcery", "Sorcery")
	c := topLibraryCard(me, "Third Creature", socB39BearType)
	b := topLibraryCard(me, "Second Creature", socB39BearType)
	a := topLibraryCard(me, "First Land", "Basic Land — Forest")
	return []uuid.UUID{a, b, c, d, e}
}

func socB39CastExperiment(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Zimone's Experiment", "Sorcery", socB39ExperimentOracle, nil)
	passPriorityAroundTable(t, g)
}

func TestSocB39ExperimentOffersOnlyCreatureAndLandCardsOfTheFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := latestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if len(prompt.ChooseCards) != 4 || hasID(prompt.ChooseCards, five[3]) {
		t.Errorf("offered %v, want the two lands and two creatures but not the sorcery", prompt.ChooseCards)
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 2 {
		t.Errorf("bounds [%d,%d], want \"up to two\"", prompt.ChooseMin, prompt.ChooseMax)
	}
	if latestChooseCardsFor(g, opp.ID) != nil {
		t.Error("the look is private: no other seat is asked")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[1], five[2]}); err == nil {
		t.Error("three cards were accepted for \"up to two\"")
	}
}

func TestSocB39ExperimentPutsTheLandOntoTheBattlefieldTappedAndTheCreatureIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := latestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	lib := me.Library.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	land := findBattlefieldCardByID(g, five[0])
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
	if bottom := libraryBottom(me, 3); !sameSet(bottom, []uuid.UUID{five[2], five[3], five[4]}) {
		t.Errorf("bottom three %v, want the unrevealed three", bottom)
	}
	for _, id := range five[:2] {
		if !hornRevealed(g, id) {
			t.Errorf("revealed card %v has no reveal event", id)
		}
	}
	for _, id := range five[2:] {
		if hornRevealed(g, id) {
			t.Errorf("card %v was only looked at but was revealed", id)
		}
	}
}

func TestSocB39ExperimentTwoLandsBothEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := latestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[0], five[4]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range []uuid.UUID{five[0], five[4]} {
		if c := findBattlefieldCardByID(g, id); c == nil || !c.Tapped {
			t.Errorf("land %v is not on the battlefield tapped", id)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no creature taken", hand, me.Hand.Size())
	}
	if bottom := libraryBottom(me, 3); !sameSet(bottom, []uuid.UUID{five[1], five[2], five[3]}) {
		t.Errorf("bottom three %v, want the three unrevealed cards", bottom)
	}
}

func TestSocB39ExperimentTwoCreaturesBothGoToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := latestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{five[1], five[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(five[1]) || !me.Hand.Contains(five[2]) || me.Hand.Size() != hand+2 {
		t.Error("both revealed creature cards belong in hand")
	}
	if bottom := libraryBottom(me, 3); !sameSet(bottom, []uuid.UUID{five[0], five[3], five[4]}) {
		t.Errorf("bottom three %v, want the three unrevealed cards", bottom)
	}
}

func TestSocB39ExperimentRevealingNothingPutsAllFiveOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	five := socB39Five(me)
	socB39CastExperiment(t, g)
	prompt := latestChooseCardsFor(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the reveal: %v", err)
	}
	if me.Hand.Size() != hand {
		t.Error("a card was taken though none was revealed")
	}
	if bottom := libraryBottom(me, 5); !sameSet(bottom, five) {
		t.Errorf("bottom five %v, want all five looked-at cards", bottom)
	}
}

func TestSocB39ExperimentWithNoCreatureOrLandAsksNothingAndBottomsAllFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		five = append(five, topLibraryCard(me, "Spell", "Sorcery"))
	}
	socB39CastExperiment(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("a prompt was opened with nothing to reveal")
	}
	if bottom := libraryBottom(me, 5); !sameSet(bottom, five) {
		t.Errorf("bottom five %v, want all five looked-at cards", bottom)
	}
}
