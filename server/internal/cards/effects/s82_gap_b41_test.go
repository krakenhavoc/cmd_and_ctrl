package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b41_test.go — GAP batch 41, Assassin's Creed enchantments,
// instants and lands (#5948). Each test plays the real catalog entry.

const (
	gapb41OriginOracle      = "9c8f1abb-bd08-41e0-91ed-56ff25d7f9f2"
	gapb41BalladOracle      = "e8e15ab4-e4c7-4faf-b469-288451ddee02"
	gapb41FallOracle        = "dd0494f5-58bf-46ae-ad69-90b063f6f660"
	gapb41ChainOracle       = "37addb87-234a-402f-b7fb-f88ef3f2a589"
	gapb41AnonymousOracle   = "bac0fcee-9c1a-46b7-86c8-ffbfcc1e96de"
	gapb41EscapeOracle      = "e16c5516-f73c-40c4-9ffa-39c6b2e92ab7"
	gapb41ImprovOracle      = "c5b85aa7-6a76-4fe2-a63e-4f376b6c9af7"
	gapb41AbstergoOracle    = "d06a8026-1657-4404-8dff-64e44f1a14f8"
	gapb41HQOracle          = "0d3a06d5-5bb9-4733-a55b-9e2c75de6b6e"
	gapb41BoneyardOracle    = "f27d52e6-aab9-4f95-ae46-33d1173bf4fe"
	gapb41QuarryOracle      = "90bccf66-58ec-445d-ac96-c6013054a1b4"
	gapb41EagleVisionOracle = "f96f1b4c-f06a-4374-8d8f-aaa06c6108c3"
)

func TestGapB41CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb41OriginOracle:    "Origin of the Hidden Ones",
		gapb41BalladOracle:    "Ballad of the Black Flag",
		gapb41FallOracle:      "Fall of the First Civilization",
		gapb41ChainOracle:     "Chain Assassination",
		gapb41AnonymousOracle: "Become Anonymous",
		gapb41EscapeOracle:    "Escape Detection",
		gapb41ImprovOracle:    "Battlefield Improvisation",
		gapb41AbstergoOracle:  "Abstergo Entertainment",
		gapb41HQOracle:        "Brotherhood Headquarters",
		gapb41BoneyardOracle:  "Submerged Boneyard",
		gapb41QuarryOracle:    "Stone Quarry",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// --- Lands --------------------------------------------------------------------

func TestGapB41TappedDualsEnterTappedAndMakeEitherColor(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		colors       [2]string
	}{
		{"Stone Quarry", gapb41QuarryOracle, [2]string{"R", "W"}},
		{"Submerged Boneyard", gapb41BoneyardOracle, [2]string{"U", "B"}},
	} {
		g, me, _ := c2LandTable(t)
		id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		untapForTest(g, id)
		for _, col := range tc.colors {
			untapForTest(g, id)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{col}}); err != nil {
				t.Errorf("%s: tap for %s: %v", tc.name, col, err)
			}
		}
		untapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
			t.Errorf("%s made {G}", tc.name)
		}
	}
}

func TestGapB41BrotherhoodHeadquartersRestrictsItsColoredMana(t *testing.T) {
	g, me, _ := c2LandTable(t)
	hq := c2Land(g, me.ID, "Brotherhood Headquarters", "Land", gapb41HQOracle)
	if err := g.ActivateManaAbility(me.ID, hq, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap for coloured mana: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	spend := func(c game.Card) bool {
		return me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(c))
	}
	assassin := game.Card{InstanceID: uuid.New(), Name: "Test Assassin", TypeLine: "Creature — Human Assassin", Owner: me.ID, Controller: me.ID}
	bear := game.Card{InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	eagle := game.Card{InstanceID: uuid.New(), Name: "Eagle Vision", TypeLine: "Sorcery", OracleID: gapb41EagleVisionOracle, ManaCost: "{4}{U}", Owner: me.ID, Controller: me.ID}
	if !spend(assassin) {
		t.Error("the mana does not pay for an Assassin spell")
	}
	if !spend(eagle) {
		t.Error("the mana does not pay for a spell that has freerunning")
	}
	if spend(bear) {
		t.Error("the mana pays for a Bear spell: stronger than printed")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(assassin)) {
		t.Error("the mana does not pay for an Assassin's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana pays for a Bear's ability: stronger than printed")
	}
}

func TestGapB41AbstergoEntertainmentReturnsHistoricAndExilesGraveyards(t *testing.T) {
	g, me, opp := c2LandTable(t)
	land := c2Land(g, me.ID, "Abstergo Entertainment", "Legendary Land", gapb41AbstergoOracle)
	relic := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: relic, Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	plain := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Relic", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	acrb1AddMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}}, Strict: true}); err == nil {
		t.Fatal("targeted a nonhistoric card")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if !inHandOf(me, relic) {
		t.Error("the historic card did not return to hand")
	}
	if me.Graveyard.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("graveyards not exiled: %d, %d", me.Graveyard.Size(), opp.Graveyard.Size())
	}
}

// --- Chain Assassination ------------------------------------------------------

func TestGapB41ChainAssassinationDrawsOnlyIfAnotherCreatureDied(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if earlier {
			doomed := b12Creature(g, me.ID, "Earlier Victim", "Creature — Bear", 1, 1)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(doomed) })
			g.RunStateChecksForTest()
		}
		ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
		advanceToMain(t, g)
		before := me.Hand.Size()
		castCatalogSpell(t, g, "Chain Assassination", "Instant", gapb41ChainOracle, cardRefs(ogre))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ogre) {
			t.Fatalf("earlier=%v: the target survived", earlier)
		}
		want := before
		if earlier {
			want = before + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("earlier=%v: hand %d -> %d, want %d", earlier, before, got, want)
		}
	}
}

func TestGapB41ChainAssassinationFreerunsAfterAnAssassinConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	spell := handCardFull(me, "Chain Assassination", "Instant", "{2}{B}{B}", gapb41ChainOracle, []string{"B"})
	floatForTest(g, me, "CB")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", Targets: cardRefs(ogre)}); err != nil {
		t.Fatalf("freerunning {1}{B} was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the target survived")
	}
}

// --- Escape Detection ---------------------------------------------------------

func TestGapB41EscapeDetectionFreerunsByReturningABlueCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	blue := b12Creature(g, me.ID, "Blue Wizard", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blue {
				g.Battlefield.Cards[i].Colors = []string{"U"}
			}
		}
	})
	spell := handCardFull(me, "Escape Detection", "Instant", "{1}{U}{U}", gapb41EscapeOracle, []string{"U"})
	before := me.Hand.Size()
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", AltCostIDs: []uuid.UUID{blue}, Targets: cardRefs(theirs)})
	if err != nil {
		t.Fatalf("freerunning (return a blue creature) was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target creature was not returned")
	}
	if !inHandOf(me, blue) {
		t.Error("the blue creature was not returned as the cost")
	}
	// The spell left the hand, the wizard came in, one card was drawn.
	if got := me.Hand.Size(); got != before-1+1+1 {
		t.Errorf("hand %d -> %d, want spell gone, wizard back, one drawn", before, got)
	}
}

func TestGapB41EscapeDetectionFreerunningNeedsAnAssassinToConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	advanceToMain(t, g)
	blue := b12Creature(g, me.ID, "Blue Wizard", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blue {
				g.Battlefield.Cards[i].Colors = []string{"U"}
			}
		}
	})
	spell := handCardFull(me, "Escape Detection", "Instant", "{1}{U}{U}", gapb41EscapeOracle, []string{"U"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", AltCostIDs: []uuid.UUID{blue}, Targets: cardRefs(theirs)}); err == nil {
		t.Error("freerunning was accepted without an Assassin connecting")
	}
}

// --- Battlefield Improvisation ------------------------------------------------

func TestGapB41BattlefieldImprovisationPumpsAndAttachesOnlyToAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	sword := b12Permanent(g, me.ID, "Test Sword", "Artifact — Equipment")
	// Not attacking: only the pump.
	advanceToMain(t, g)
	acrb1AddMana(t, g, me.ID, "{C}{W}")
	castCatalogSpell(t, g, "Battlefield Improvisation", "Instant", gapb41ImprovOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked about Equipment for a creature that is not attacking")
	}
	// Attacking: the Equipment may be attached.
	attackWith(t, g, opp.ID, bear)
	spell := handCardFull(me, "Battlefield Improvisation", "Instant", "{1}{W}", gapb41ImprovOracle, []string{"W"})
	floatForTest(g, me, "CW")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, sword)
	if c, _ := battlefieldCard(g, sword); c.AttachedTo.ID != bear {
		t.Errorf("the Equipment is attached to %v, want the attacking Bear", c.AttachedTo)
	}
}

// --- Become Anonymous ---------------------------------------------------------

func TestGapB41BecomeAnonymousCloaksTheTargetAndTwoCardsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	top1 := pushLibraryTyped(me, "Top One", "Creature — Elf")
	top2 := pushLibraryTyped(me, "Top Two", "Instant")
	castCatalogSpell(t, g, "Become Anonymous", "Instant", gapb41AnonymousOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	faceDown := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.FaceDownIsPermanent() {
			faceDown++
			if !c.Tapped {
				t.Errorf("a cloaked permanent entered untapped")
			}
		}
	}
	if faceDown != 3 {
		t.Fatalf("%d cloaked permanents, want 3 (the target and the top two cards)", faceDown)
	}
	_ = top1
	_ = top2
}

// --- Sagas --------------------------------------------------------------------

func TestGapB41OriginOfTheHiddenOnesPlaysItsChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	assassin := b12Creature(g, me.ID, "Old Assassin", "Creature — Human Assassin", 1, 1)
	saga := castCatalogSpell(t, g, "Origin of the Hidden Ones", "Enchantment — Saga", gapb41OriginOracle, nil)
	_ = saga
	life := opp.Life
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("chapter I: life %d -> %d, want -4", life, opp.Life)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Assassin" {
			tokens++
			if !game.HasKeyword(&c, "menace") || c.Power != 1 || c.Toughness != 1 {
				t.Errorf("Assassin token = %+v", c)
			}
		}
	}
	if tokens != 2 {
		t.Fatalf("chapter II made %d Assassins, want 2", tokens)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(assassin, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	sncSettle(t, g, me.ID, uuid.Nil)
	attacking := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Assassin" && c.AttackingTarget != uuid.Nil && c.Tapped {
			attacking++
		}
	}
	if attacking != 1 {
		t.Errorf("%d tapped and attacking Assassin tokens, want 1", attacking)
	}
}

func TestGapB41BalladOfTheBlackFlagTakesAHistoricCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := pushLibraryTyped(me, "Relic", "Artifact")
	pushLibraryTyped(me, "Bear", "Creature — Bear")
	pushLibraryTyped(me, "Island", "Basic Land — Island")
	castCatalogSpell(t, g, "Ballad of the Black Flag", "Enchantment — Saga", gapb41BalladOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !inHandOf(me, relic) {
		t.Error("the historic card was not put into hand")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Relic" {
			t.Error("the taken card stayed in the graveyard")
		}
	}
}

func TestGapB41BalladChapterFourDiscountsHistoricSpells(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := castCatalogSpell(t, g, "Ballad of the Black Flag", "Enchantment — Saga", gapb41BalladOracle, nil)
	passPriorityAroundTable(t, g)
	b22SetLore(g, mustBattlefieldPtr(t, g, saga), 3)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	relic := handSpell(me, "Test Relic", "Artifact", "{3}")
	bauble := handSpell(me, "Test Bauble", "Instant", "{3}")
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, relic, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a historic {3} spell with one mana after chapter IV: %v", err)
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, bauble, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a nonhistoric spell was discounted")
	}
}

func mustBattlefieldPtr(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

func TestGapB41FallOfTheFirstCivilizationChapterOneAndTwo(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	myRelic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	castCatalogSpell(t, g, "Fall of the First Civilization", "Enchantment — Saga", gapb41FallOracle, nil)
	passPriorityAroundTable(t, g)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != mine+2 || opp.Hand.Size() != theirs+2 {
		t.Errorf("hands %d->%d, %d->%d, want +2 each", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
	advanceToPrecombatMainOf(t, g, seat)
	pickTriggerTarget(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("chapter II did not exile the opponent's artifact")
	}
	if !g.Battlefield.Contains(myRelic) {
		t.Error("chapter II exiled my own artifact")
	}
}

func TestGapB41FallOfTheFirstCivilizationChapterThreeKeepsThreeEach(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	var mine, theirs []uuid.UUID
	for i := 0; i < 5; i++ {
		mine = append(mine, b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1))
		theirs = append(theirs, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1))
	}
	land := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	saga := castCatalogSpell(t, g, "Fall of the First Civilization", "Enchantment — Saga", gapb41FallOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	b22SetLore(g, mustBattlefieldPtr(t, g, saga), 2)
	advanceToPrecombatMainOf(t, g, seat)
	// Lore 3 is chapter III directly.
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, mine[0], mine[1], mine[2])
	answerChooseCards(t, g, opp.ID, theirs[0], theirs[1], theirs[2])
	passPriorityAroundTable(t, g)
	survivors := func(ids []uuid.UUID) int {
		n := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if survivors(mine) != 3 || survivors(theirs) != 3 {
		t.Errorf("survivors %d and %d, want 3 each", survivors(mine), survivors(theirs))
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}
