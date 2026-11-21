package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b131_test.go — GAP batch 131 (S82 #5948), The Lord of the Rings:
// Tales of Middle-earth: Council's Deliberation; Deceive the Messenger;
// Isolation at Orthanc; Press the Enemy; Saruman's Trickery; Escape from
// Orthanc; Hobbit's Sting; Lost to Legend; Second Breakfast; You Cannot Pass!;
// Great Hall of the Citadel; The Grey Havens.

const (
	gapCouncilsOracle    = "3280488d-e94b-4d53-b279-c32ed711d4bc"
	gapDeceiveOracle     = "a9720545-37f4-42f1-932f-13ef88c9ea72"
	gapIsolationOracle   = "6dc6536f-455a-4d7b-beed-a78ec63c0030"
	gapPressOracle       = "347a3622-4425-4ba4-821d-ef3d00beb058"
	gapTrickeryOracle    = "1b415806-a3e3-4d87-b4c1-ccd11a8ee589"
	gapEscapeOracle      = "f2025f2e-c931-48dd-af21-8f0a591befb4"
	gapStingOracle       = "845bbad1-02f9-432b-aca4-13a5ec71fd06"
	gapLostToLegendOracl = "a92edf30-6c7f-481c-855e-2a7dc4963ea7"
	gapBreakfastOracle   = "bf5f68b8-8ae2-4d08-b86c-4e6d3b653fec"
	gapCannotPassOracle  = "8e94ccdb-5978-440b-b7cd-11c7ca6b2c98"
	gapGreatHallOracle   = "b3e28bcf-0ed0-4406-b615-68ddc55b349a"
	gapGreyHavensOracle  = "a1a9695e-073b-4a65-b3ec-2cfddc23202a"
)

func TestGapB131CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapCouncilsOracle:    "Council's Deliberation",
		gapDeceiveOracle:     "Deceive the Messenger",
		gapIsolationOracle:   "Isolation at Orthanc",
		gapPressOracle:       "Press the Enemy",
		gapTrickeryOracle:    "Saruman's Trickery",
		gapEscapeOracle:      "Escape from Orthanc",
		gapStingOracle:       "Hobbit's Sting",
		gapLostToLegendOracl: "Lost to Legend",
		gapBreakfastOracle:   "Second Breakfast",
		gapCannotPassOracle:  "You Cannot Pass!",
		gapGreatHallOracle:   "Great Hall of the Citadel",
		gapGreyHavensOracle:  "The Grey Havens",
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
}

// gapB131Instant casts an instant from the active seat's hand in the current
// step, strictly, and returns the cast's error.
func gapB131Instant(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

func gapB131PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

func gapTop(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Test", Owner: p.ID, Controller: p.ID})
	}
}

// --- Council's Deliberation ---------------------------------------------------

func TestGapCouncilsDeliberationDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapTop(me, 3)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Council's Deliberation", gapCouncilsOracle, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func gapCouncilsInGraveyard(t *testing.T, island bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapTop(me, 4)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Council's Deliberation", TypeLine: "Instant", OracleID: gapCouncilsOracle, Owner: me.ID, Controller: me.ID})
	if island {
		effectstest.PushPermanentForTest(g, me.ID, "Island", "", "Basic Land — Island")
	}
	effectstest.AdvanceToMain(t, g)
	return g, me, id
}

func gapScryOnce(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapCouncilsDeliberationExilesItselfFromTheGraveyardToDrawWhenYouScry(t *testing.T) {
	g, me, id := gapCouncilsInGraveyard(t, true)
	hand := me.Hand.Size()
	gapScryOnce(t, g, me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
		t.Errorf("the card is in %s, want exile", effectstest.B12ZoneOf(g, id))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestGapCouncilsDeliberationMayStayInTheGraveyard(t *testing.T) {
	g, me, id := gapCouncilsInGraveyard(t, true)
	hand := me.Hand.Size()
	gapScryOnce(t, g, me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, id) != game.ZoneGraveyard || me.Hand.Size() != hand {
		t.Errorf("declined: card in %s, hand %d → %d", effectstest.B12ZoneOf(g, id), hand, me.Hand.Size())
	}
}

func TestGapCouncilsDeliberationNeedsAnIsland(t *testing.T) {
	g, me, id := gapCouncilsInGraveyard(t, false)
	gapScryOnce(t, g, me)
	if len(g.PendingChoices) != 0 {
		t.Errorf("the trigger asked something without an Island: %d prompts open", len(g.PendingChoices))
	}
	if effectstest.B12ZoneOf(g, id) != game.ZoneGraveyard {
		t.Errorf("the card is in %s, want the graveyard", effectstest.B12ZoneOf(g, id))
	}
}

// --- Deceive the Messenger / Saruman's Trickery -------------------------------

func TestGapDeceiveTheMessengerShrinksAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Deceive the Messenger", gapDeceiveOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("power %d, want 5-3=2", p)
	}
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 || !army.HasSubtype("Orc") {
		t.Errorf("Army: %d counters, Orc %v; want 1 and an Orc", got, army.HasSubtype("Orc"))
	}
}

func TestGapSarumansTrickeryCountersTheSpellAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast the bolt: %v", err)
	}
	if err := gapB131Instant(t, g, "Saruman's Trickery", gapTrickeryOracle, nil, effectstest.CardRefs(bolt)); err != nil {
		t.Fatalf("cast Saruman's Trickery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the bolt resolved: life %d → %d", life, opp.Life)
	}
	if effectstest.B12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Errorf("the countered bolt is in %s", effectstest.B12ZoneOf(g, bolt))
	}
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters %d, want 1", got)
	}
}

// --- Isolation at Orthanc / Lost to Legend ------------------------------------

func TestGapIsolationAtOrthancPutsTheCreatureSecondFromTheTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	top, _ := opp.Library.Top()
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Isolation at Orthanc", gapIsolationOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := opp.Library.Size()
	if opp.Library.Cards[n-2].InstanceID != theirs || opp.Library.Cards[n-1].InstanceID != top.InstanceID {
		t.Error("the creature is not second from the top of its owner's library")
	}
	if me.Library.Contains(theirs) {
		t.Error("the creature went to the caster's library")
	}
}

func TestGapLostToLegendTakesOnlyNonlandHistoricPermanentsFourthFromTheTop(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"Rock", "Artifact", true},
		{"Legend", "Legendary Creature — Human", true},
		{"Old Tale", "Enchantment — Saga", true},
		{"Bear", "Creature — Bear", false},
		{"Legendary Land", "Legendary Land", false},
	} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		target := effectstest.PushPermanentForTest(g, opp.ID, c.name, "", c.typeLine)
		effectstest.AdvanceToMain(t, g)
		err := gapB131Instant(t, g, "Lost to Legend", gapLostToLegendOracl, nil, effectstest.CardRefs(target))
		if (err == nil) != c.ok {
			t.Errorf("%s: cast err = %v, legal %v", c.name, err, c.ok)
			continue
		}
		if !c.ok {
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		n := opp.Library.Size()
		if opp.Library.Cards[n-4].InstanceID != target {
			t.Errorf("%s is not fourth from the top", c.name)
		}
	}
}

// --- Press the Enemy ----------------------------------------------------------

func TestGapPressTheEnemyBouncesAndOffersACheaperInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	cheap := effectstest.HandCardFull(me, "Cheap Instant", "Instant", "{1}{R}", "", []string{"R"})
	equal := effectstest.HandCardFull(me, "Equal Sorcery", "Sorcery", "{2}{R}", "", []string{"R"})
	dear := effectstest.HandCardFull(me, "Dear Sorcery", "Sorcery", "{3}{R}", "", []string{"R"})
	creature := effectstest.HandCardFull(me, "A Creature", "Creature — Bear", "{1}", "", nil)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Press the Enemy", gapPressOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneHand {
		t.Fatalf("the permanent is in %s, want its owner's hand", effectstest.B12ZoneOf(g, theirs))
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap, equal)
	for _, id := range offer.DiscardOptions {
		if id == dear || id == creature {
			t.Errorf("an instant or sorcery with a greater mana value, or a creature, is offered: %v", id)
		}
	}
}

func TestGapPressTheEnemyRefusesYourOwnPermanentsAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	for name, target := range map[string]uuid.UUID{"my own creature": mine, "an opponent's land": land} {
		if err := gapB131Instant(t, g, "Press the Enemy", gapPressOracle, nil, effectstest.CardRefs(target)); err == nil {
			t.Errorf("%s was a legal target", name)
		}
	}
}

// --- Escape from Orthanc ------------------------------------------------------

func TestGapEscapeFromOrthancPumpsGivesFlyingAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Escape from Orthanc", gapEscapeOracle, nil, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := gapB131PT(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the creature is %d/%d, want 3/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("no flying")
	}
	if c := effectstest.LayeredCard(t, g, bear); c.Tapped {
		t.Error("the creature is still tapped")
	}
}

// --- Hobbit's Sting -----------------------------------------------------------

func TestGapHobbitsStingCountsCreaturesPlusFoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 1, 1)
	effectstest.PushToken(g, me.ID, effects.FoodToken())
	effectstest.PushToken(g, me.ID, effects.FoodToken())
	// Neither the opponent's creatures nor their Foods count.
	effectstest.B12Creature(g, opp.ID, "Theirs One", "Creature — Bear", 1, 1)
	effectstest.PushToken(g, opp.ID, effects.FoodToken())
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 4)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Hobbit's Sting", gapStingOracle, nil, effectstest.CardRefs(small)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, small) != game.ZoneGraveyard {
		t.Errorf("the 1/4 took 2 creatures + 2 Foods = 4 and is in %s", effectstest.B12ZoneOf(g, small))
	}
	if c := effectstest.LayeredCard(t, g, big); c.DamageMarked != 0 {
		t.Errorf("an untargeted creature took %d", c.DamageMarked)
	}
}

// --- Second Breakfast ---------------------------------------------------------

func TestGapSecondBreakfastPumpsUpToTwoAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c3 := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Second Breakfast", gapBreakfastOracle, nil, effectstest.CardRefs(a, b, c3)); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := gapB131Instant(t, g, "Second Breakfast", gapBreakfastOracle, nil, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := gapB131PT(t, g, id); p != 4 || tough != 3 {
			t.Errorf("a target is %d/%d, want 4/3", p, tough)
		}
	}
	if p, _ := gapB131PT(t, g, c3); p != 2 {
		t.Errorf("the untargeted creature has power %d", p)
	}
	if effectstest.CountBattlefieldByName(g, "Food") != 1 {
		t.Errorf("Foods = %d, want 1", effectstest.CountBattlefieldByName(g, "Food"))
	}
}

func TestGapSecondBreakfastWithNoTargetsStillMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	if err := gapB131Instant(t, g, "Second Breakfast", gapBreakfastOracle, nil, nil); err != nil {
		t.Fatalf("cast with no targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Food") != 1 {
		t.Errorf("Foods = %d, want 1", effectstest.CountBattlefieldByName(g, "Food"))
	}
}

// --- You Cannot Pass! ---------------------------------------------------------

func TestGapYouCannotPassDestroysACreatureThatBlockedOrWasBlockedByALegend(t *testing.T) {
	for _, legendIsBlocker := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 5)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 5)
		bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
		if legendIsBlocker {
			g.WithWriteLock(func() { gapMakeLegendary(g, blocker) })
		} else {
			g.WithWriteLock(func() { gapMakeLegendary(g, attacker) })
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInCombatDeclaration(t, g)
		victim := attacker
		if !legendIsBlocker {
			victim = blocker
		}
		if err := gapB131Instant(t, g, "You Cannot Pass!", gapCannotPassOracle, nil, effectstest.CardRefs(bystander)); err == nil {
			t.Error("a creature that did not block or get blocked by a legend was a legal target")
		}
		if err := gapB131Instant(t, g, "You Cannot Pass!", gapCannotPassOracle, nil, effectstest.CardRefs(victim)); err != nil {
			t.Fatalf("legend as blocker %v: cast at %s: %v", legendIsBlocker, victim, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.B12ZoneOf(g, victim) != game.ZoneGraveyard {
			t.Errorf("legend as blocker %v: the target is in %s, want the graveyard", legendIsBlocker, effectstest.B12ZoneOf(g, victim))
		}
	}
}

// gapMakeLegendary adds the legendary supertype to a battlefield card.
func gapMakeLegendary(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].TypeLine = "Legendary " + g.Battlefield.Cards[i].TypeLine
		}
	}
}

// --- Great Hall of the Citadel ------------------------------------------------

func TestGapGreatHallMakesColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hall := effectstest.B12Push(g, me.ID, "Great Hall of the Citadel", "Land", gapGreatHallOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if effectstest.PoolCount(me, "C") != 1 {
		t.Errorf("colorless mana in the pool = %d, want 1", effectstest.PoolCount(me, "C"))
	}
}

func TestGapGreatHallsTwoManaPayOnlyForLegendarySpells(t *testing.T) {
	for _, legendary := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		hall := effectstest.B12Push(g, me.ID, "Great Hall of the Citadel", "Land", gapGreatHallOracle, 0, 0)
		typeLine := "Creature — Bear"
		if legendary {
			typeLine = "Legendary Creature — Bear"
		}
		spell := effectstest.HandCardFull(me, "Test Spell", typeLine, "{R}{R}", "", []string{"R"})
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "C")
		if err := g.ActivateManaAbility(me.ID, hall, 1, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("{1},{T}: %v", err)
		}
		effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R")
		err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true})
		if legendary && err != nil {
			t.Errorf("a legendary spell could not be paid with the restricted mana: %v", err)
		}
		if !legendary && err == nil {
			t.Error("a non-legendary spell was paid with mana restricted to legendary spells")
		}
	}
}

// --- The Grey Havens ----------------------------------------------------------

func TestGapGreyHavensScriesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapTop(me, 3)
	effectstest.PlayLandFromHand(t, g, "The Grey Havens", gapGreyHavensOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestGapGreyHavensMakesAColorAmongLegendaryCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	havens := effectstest.B12Push(g, me.ID, "The Grey Havens", "Legendary Land", gapGreyHavensOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	// A red legendary creature card, a green non-legendary one and a blue
	// legendary land card: only red counts.
	for _, c := range []game.Card{
		{Name: "Red Legend", TypeLine: "Legendary Creature — Human", Colors: []string{"R"}},
		{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}},
		{Name: "Blue Place", TypeLine: "Legendary Land", Colors: []string{"U"}},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	if err := g.ActivateManaAbility(me.ID, havens, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add one mana of any color among legendary creature cards: %v", err)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R")
	if effectstest.PoolCount(me, "R") != 1 || effectstest.PoolCount(me, "G") != 0 || effectstest.PoolCount(me, "U") != 0 {
		t.Errorf("pool R=%d G=%d U=%d, want only red", effectstest.PoolCount(me, "R"), effectstest.PoolCount(me, "G"), effectstest.PoolCount(me, "U"))
	}
}

func TestGapGreyHavensWithNoLegendaryCreatureCardAddsNothingAndStillTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	havens := effectstest.B12Push(g, me.ID, "The Grey Havens", "Legendary Land", gapGreyHavensOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, havens, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if n := len(me.ManaPool.Units()); n != 0 {
		t.Errorf("%d mana in the pool, want none", n)
	}
	if c := effectstest.LayeredCard(t, g, havens); !c.Tapped {
		t.Error("the land did not tap")
	}
}
