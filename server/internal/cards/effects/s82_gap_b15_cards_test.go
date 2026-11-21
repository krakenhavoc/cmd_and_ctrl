package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b15_cards_test.go — Innistrad Remastered instants, planeswalkers
// and sorceries, GAP batch 15 (#5948). Each test plays the real catalog
// entry through the engine.

const (
	gapAimHighOracle          = "c169929a-850b-4d46-9f14-a922f4465d85"
	gapThroughTheBreachOracle = "c207aaee-0b67-4520-ad02-2c289228be2a"
	gapCompellingDetOracle    = "fb0c9048-73ee-434b-a018-dbfbe32e56f5"
	gapSummaryDismissalOracle = "f69d65e0-4d8a-4e27-b4dd-dd8c50acb8eb"
	gapStrengthOfArmsOracle   = "40ff84c7-cd72-49a6-86ba-a8d23c819ee7"
	gapArlinnOracle           = "6bb7d0df-7a9f-4e17-8c31-7628c4b12356"
	gapJaceUnravelerOracle    = "ac29873e-28c3-4933-b427-61d34ff0afcd"
	gapKillingWaveOracle      = "69d6b906-5461-4eba-8667-dbb8c0ce3fcb"
)

func TestGapB15EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapAimHighOracle:          "Aim High",
		gapThroughTheBreachOracle: "Through the Breach",
		gapCompellingDetOracle:    "Compelling Deterrence",
		gapSummaryDismissalOracle: "Summary Dismissal",
		gapStrengthOfArmsOracle:   "Strength of Arms",
		gapArlinnOracle:           "Arlinn Kord",
		gapArlinnOracle + "#1":    "Arlinn, Embraced by the Moon",
		gapJaceUnravelerOracle:    "Jace, Unraveler of Secrets",
		gapKillingWaveOracle:      "Killing Wave",
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
	if spec, _ := Lookup(gapThroughTheBreachOracle); spec.Splice == nil || spec.Splice.ManaCost != "{2}{R}{R}" {
		t.Errorf("Through the Breach splice %+v, want onto Arcane {2}{R}{R}", spec.Splice)
	}
}

func TestGapB15AimHighUntapsAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if !tappedForTest(t, g, bear) {
		t.Fatal("setup: the bear should be tapped")
	}
	castCatalogSpell(t, g, "Aim High", "Instant", gapAimHighOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, bear) {
		t.Error("the creature was not untapped")
	}
	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 {
		t.Errorf("bear is %d/%d, want 4/4", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}
	assertKeywords(t, g, bear, "reach")
}

func TestGapB15ThroughTheBreachPutsACreatureInWithHasteAndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fatty := putInHand(me, game.Card{Name: "Big Fatty", TypeLine: "Creature — Beast", Power: 6, Toughness: 6})
	castCatalogSpell(t, g, "Through the Breach", "Instant — Arcane", gapThroughTheBreachOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, fatty)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fatty) {
		t.Fatal("the creature did not enter")
	}
	assertKeywords(t, g, fatty, "haste")
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fatty) {
		t.Error("the creature was not sacrificed at the beginning of the next end step")
	}
}

func TestGapB15ThroughTheBreachMayPutNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fatty := putInHand(me, game.Card{Name: "Big Fatty", TypeLine: "Creature — Beast", Power: 6, Toughness: 6})
	castCatalogSpell(t, g, "Through the Breach", "Instant — Arcane", gapThroughTheBreachOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fatty) || !me.Hand.Contains(fatty) {
		t.Error("a creature entered after the player declined")
	}
}

func TestGapB15CompellingDeterrenceBouncesAndDiscardsOnlyWithAZombie(t *testing.T) {
	for _, withZombie := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withZombie {
			b12Creature(g, me.ID, "My Zombie", "Creature — Zombie", 2, 2)
		}
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		hand := opp.Hand.Size()
		castCatalogSpell(t, g, "Compelling Deterrence", "Instant", gapCompellingDetOracle, cardRefs(theirs))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
			t.Fatalf("zombie %v: the permanent was not returned to its owner's hand", withZombie)
		}
		if withZombie {
			answerChooseCardsOrDiscard(t, g, opp)
			if got := opp.Hand.Size(); got != hand {
				t.Errorf("owner's hand %d, want %d (one in, one discarded)", got, hand)
			}
		} else if got := opp.Hand.Size(); got != hand+1 {
			t.Errorf("owner's hand %d, want %d: no Zombie, no discard", got, hand+1)
		}
	}
}

func TestGapB15CompellingDeterrenceCannotTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := castCatalogSpellErrStrict(t, g, "Compelling Deterrence", "Instant", gapCompellingDetOracle, cardRefs(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

func TestGapB15SummaryDismissalExilesOtherSpellsAndCountersAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	claws := pushCatalogPermanent(g, me.ID, "Claws of Gix", "Artifact", usgB2ClawsOfGixOracle, false)
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	advanceToMain(t, g)
	life := me.Life
	// An opponent's Bolt aimed at me, then my own ability above it.
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 1 })
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, claws, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate Claws: %v", err)
	}
	b06AddMana(me, "U", "U", "U", "U")
	dismissal := castFromSeat(t, g, me, "Summary Dismissal", "Instant", gapSummaryDismissalOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Error("the other spell was not exiled")
	}
	if g.Stack.Contains(dismissal) {
		t.Error("Summary Dismissal did not resolve")
	}
	if me.Life != life {
		t.Errorf("life %d → %d: the Bolt or the Claws ability still resolved", life, me.Life)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d items remain on the stack", len(g.StackMeta))
	}
}

func TestGapB15StrengthOfArmsMakesAHumanSoldierOnlyWithAnEquipment(t *testing.T) {
	for _, withEquipment := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		if withEquipment {
			b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
		}
		castCatalogSpell(t, g, "Strength of Arms", "Instant", gapStrengthOfArmsOracle, cardRefs(bear))
		passPriorityAroundTable(t, g)
		if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 {
			t.Errorf("bear is %d/%d, want 4/4", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
		}
		want := 0
		if withEquipment {
			want = 1
		}
		if got := b27CountNamed(g, me.ID, "Human Soldier"); got != want {
			t.Errorf("equipment %v: %d Human Soldier tokens, want %d", withEquipment, got, want)
		}
	}
}

// --- Killing Wave ---

func gapKillingWave(t *testing.T, g *game.Game, x int) {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Killing Wave", TypeLine: "Sorcery", OracleID: gapKillingWaveOracle,
		ManaCost: "{X}{B}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "B", "B", "B", "B", "B", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast Killing Wave: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestGapB15KillingWaveSacrificesWhatIsNotPaidFor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine2 := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lifeBefore := me.Life
	gapKillingWave(t, g, 3)
	// I keep one creature (3 life); the opponent pays for nothing.
	answerChooseCards(t, g, me.ID, mine)
	answerChooseCards(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature I paid for was sacrificed")
	}
	if g.Battlefield.Contains(mine2) || g.Battlefield.Contains(theirs) {
		t.Error("an unpaid creature survived")
	}
	if want := lifeBefore - 3; me.Life != want {
		t.Errorf("life %d, want %d: one payment of X = 3", me.Life, want)
	}
}

func TestGapB15KillingWaveOnlyOffersEachPlayerTheirOwnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gapKillingWave(t, g, 1)
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceChooseCards {
			continue
		}
		for _, id := range c.ChooseCards {
			owner := controllerOf(t, g, id)
			if owner != c.Chooser {
				t.Errorf("%v was offered %v, a creature %v controls", c.Chooser, id, owner)
			}
		}
	}
	answerChooseCards(t, g, me.ID, mine)
	answerChooseCards(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a creature that was paid for was sacrificed")
	}
}

func TestGapB15KillingWaveWithXZeroKeepsEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	gapKillingWave(t, g, 0)
	if !g.Battlefield.Contains(mine) {
		t.Error("a creature was sacrificed with X = 0: paying 0 life is free")
	}
}

// --- Arlinn Kord ---

func gapArlinnCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   gapArlinnOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: 3},
		Faces: []game.Face{
			{Name: "Arlinn Kord", TypeLine: "Legendary Planeswalker — Arlinn", ManaCost: "{2}{R}{G}",
				Colors: []string{"R", "G"}, StartingLoyalty: 3},
			{Name: "Arlinn, Embraced by the Moon", TypeLine: "Legendary Planeswalker — Arlinn", Colors: []string{"R", "G"}},
		},
	}
	c.SetFace(0)
	return c
}

func TestGapB15ArlinnKordPlusOnePumpsAndGrantsVigilanceAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	arlinn := pushBattlefieldCardWithTimestamp(g, gapArlinnCard(me.ID))
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	apaActivate(t, g, me, arlinn, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 {
		t.Errorf("bear is %d/%d, want 4/4", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}
	assertKeywords(t, g, bear, "vigilance", "haste")
	if c := apaLive(g, arlinn); c.Counters[game.CounterLoyalty] != 4 || c.ActiveFace != 0 {
		t.Errorf("loyalty %d face %d, want 4 and the front", c.Counters[game.CounterLoyalty], c.ActiveFace)
	}
}

func TestGapB15ArlinnKordPlusOneWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	arlinn := pushBattlefieldCardWithTimestamp(g, gapArlinnCard(me.ID))
	apaActivate(t, g, me, arlinn, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if c := apaLive(g, arlinn); c.Counters[game.CounterLoyalty] != 4 {
		t.Errorf("loyalty %d, want 4: up to one target allows none", c.Counters[game.CounterLoyalty])
	}
}

func TestGapB15ArlinnKordZeroMakesAWolfAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	arlinn := pushBattlefieldCardWithTimestamp(g, gapArlinnCard(me.ID))
	apaActivate(t, g, me, arlinn, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Wolf"); n != 1 {
		t.Fatalf("%d Wolves, want 1", n)
	}
	c := apaLive(g, arlinn)
	if c == nil || c.ActiveFace != 1 || c.Name != "Arlinn, Embraced by the Moon" {
		t.Fatalf("Arlinn did not transform: %+v", c)
	}
}

func gapArlinnBack(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	c := gapArlinnCard(owner)
	c.Counters[game.CounterLoyalty] = loyalty
	c.SetFace(1)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestGapB15ArlinnEmbracedByTheMoonPlusOnePumpsTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	arlinn := gapArlinnBack(g, me.ID, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	apaActivate(t, g, me, arlinn, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, mine) != 3 || effectiveToughness(t, g, mine) != 3 {
		t.Errorf("mine is %d/%d, want 3/3", effectivePower(t, g, mine), effectiveToughness(t, g, mine))
	}
	assertKeywords(t, g, mine, "trample")
	if effectivePower(t, g, theirs) != 2 {
		t.Error("their creature was pumped")
	}
}

func TestGapB15ArlinnEmbracedByTheMoonMinusOneDamagesThenTransformsBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	arlinn := gapArlinnBack(g, me.ID, 3)
	life := opp.Life
	apaActivate(t, g, me, arlinn, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-3)
	}
	if c := apaLive(g, arlinn); c == nil || c.ActiveFace != 0 || c.Counters[game.CounterLoyalty] != 2 {
		t.Errorf("Arlinn should be back on the front face at 2 loyalty: %+v", c)
	}
}

func TestGapB15ArlinnEmbracedByTheMoonEmblemGivesHasteAndAPingAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	arlinn := gapArlinnBack(g, me.ID, 6)
	apaActivate(t, g, me, arlinn, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	assertKeywords(t, g, bear, "haste")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d: damage equal to its power", opp.Life, life-3)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, theirs)
}

// --- Jace, Unraveler of Secrets ---

func TestGapB15JaceUnravelerPlusOneScriesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	jace := pushCatalogWalkerWithLine(g, me.ID, "Jace, Unraveler of Secrets", "Legendary Planeswalker — Jace", gapJaceUnravelerOracle, 5)
	hand := me.Hand.Size()
	top := putOnTopForTest(me, "Scried Top", "Land")
	apaActivate(t, g, me, jace, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry 1 was asked")
	}
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("the card was drawn before the scry finished: hand %d, want %d", got, hand)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the card left on top was not the card drawn")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if c := apaLive(g, jace); c.Counters[game.CounterLoyalty] != 6 {
		t.Errorf("loyalty %d, want 6", c.Counters[game.CounterLoyalty])
	}
}

func TestGapB15JaceUnravelerMinusTwoBouncesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jace := pushCatalogWalkerWithLine(g, me.ID, "Jace, Unraveler of Secrets", "Legendary Planeswalker — Jace", gapJaceUnravelerOracle, 5)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	apaActivate(t, g, me, jace, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

func TestGapB15JaceUnravelerEmblemCountersAnOpponentsFirstSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jace := pushCatalogWalkerWithLine(g, me.ID, "Jace, Unraveler of Secrets", "Legendary Planeswalker — Jace", gapJaceUnravelerOracle, 8)
	apaActivate(t, g, me, jace, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	life := me.Life
	cast := func() uuid.UUID {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
			Owner: opp.ID, Controller: opp.ID})
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
			t.Fatalf("cast Bolt: %v", err)
		}
		passPriorityAroundTable(t, g)
		return id
	}
	first := cast()
	if me.Life != life || !opp.Graveyard.Contains(first) {
		t.Errorf("the first spell was not countered: life %d → %d", life, me.Life)
	}
	cast()
	if me.Life != life-3 {
		t.Errorf("life %d, want %d: the second spell of the turn is not countered", me.Life, life-3)
	}
}

func TestGapB15JaceUnravelerEmblemIgnoresYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jace := pushCatalogWalkerWithLine(g, me.ID, "Jace, Unraveler of Secrets", "Legendary Planeswalker — Jace", gapJaceUnravelerOracle, 8)
	apaActivate(t, g, me, jace, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d: my own first spell is not countered", opp.Life, life-3)
	}
}
