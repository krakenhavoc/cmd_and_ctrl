package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// --- Lands --------------------------------------------------------------------

func TestGapB41BrotherhoodHeadquartersRestrictsItsColoredMana(t *testing.T) {
	g, me, _ := effectstest.C2LandTable(t)
	hq := effectstest.C2Land(g, me.ID, "Brotherhood Headquarters", "Land", gapb41HQOracle)
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
	g, me, opp := effectstest.C2LandTable(t)
	land := effectstest.C2Land(g, me.ID, "Abstergo Entertainment", "Legendary Land", gapb41AbstergoOracle)
	relic := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: relic, Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	plain := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Relic", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}}, Strict: true}); err == nil {
		t.Fatal("targeted a nonhistoric card")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, relic) {
		t.Error("the historic card did not return to hand")
	}
	if me.Graveyard.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("graveyards not exiled: %d, %d", me.Graveyard.Size(), opp.Graveyard.Size())
	}
}

// --- Chain Assassination ------------------------------------------------------

func TestGapB41ChainAssassinationDrawsOnlyIfAnotherCreatureDied(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if earlier {
			doomed := effectstest.B12Creature(g, me.ID, "Earlier Victim", "Creature — Bear", 1, 1)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(doomed) })
			g.RunStateChecksForTest()
		}
		ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
		effectstest.AdvanceToMain(t, g)
		before := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Chain Assassination", "Instant", gapb41ChainOracle, effectstest.CardRefs(ogre))
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.Acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	spell := effectstest.HandCardFull(me, "Chain Assassination", "Instant", "{2}{B}{B}", gapb41ChainOracle, []string{"B"})
	effectstest.FloatForTest(g, me, "CB")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("freerunning {1}{B} was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the target survived")
	}
}

// --- Escape Detection ---------------------------------------------------------

func TestGapB41EscapeDetectionFreerunsByReturningABlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.Acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	blue := effectstest.B12Creature(g, me.ID, "Blue Wizard", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blue {
				g.Battlefield.Cards[i].Colors = []string{"U"}
			}
		}
	})
	spell := effectstest.HandCardFull(me, "Escape Detection", "Instant", "{1}{U}{U}", gapb41EscapeOracle, []string{"U"})
	before := me.Hand.Size()
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", AltCostIDs: []uuid.UUID{blue}, Targets: effectstest.CardRefs(theirs)})
	if err != nil {
		t.Fatalf("freerunning (return a blue creature) was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target creature was not returned")
	}
	if !effectstest.InHandOf(me, blue) {
		t.Error("the blue creature was not returned as the cost")
	}
	// The spell left the hand, the wizard came in, one card was drawn.
	if got := me.Hand.Size(); got != before-1+1+1 {
		t.Errorf("hand %d -> %d, want spell gone, wizard back, one drawn", before, got)
	}
}

func TestGapB41EscapeDetectionFreerunningNeedsAnAssassinToConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	blue := effectstest.B12Creature(g, me.ID, "Blue Wizard", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blue {
				g.Battlefield.Cards[i].Colors = []string{"U"}
			}
		}
	})
	spell := effectstest.HandCardFull(me, "Escape Detection", "Instant", "{1}{U}{U}", gapb41EscapeOracle, []string{"U"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", AltCostIDs: []uuid.UUID{blue}, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("freerunning was accepted without an Assassin connecting")
	}
}

// --- Battlefield Improvisation ------------------------------------------------

func TestGapB41BattlefieldImprovisationPumpsAndAttachesOnlyToAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	sword := effectstest.B12Permanent(g, me.ID, "Test Sword", "Artifact — Equipment")
	// Not attacking: only the pump.
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{W}")
	effectstest.CastCatalogSpell(t, g, "Battlefield Improvisation", "Instant", gapb41ImprovOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked about Equipment for a creature that is not attacking")
	}
	// Attacking: the Equipment may be attached.
	effectstest.AttackWith(t, g, opp.ID, bear)
	spell := effectstest.HandCardFull(me, "Battlefield Improvisation", "Instant", "{1}{W}", gapb41ImprovOracle, []string{"W"})
	effectstest.FloatForTest(g, me, "CW")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	if c, _ := effectstest.BattlefieldCard(g, sword); c.AttachedTo.ID != bear {
		t.Errorf("the Equipment is attached to %v, want the attacking Bear", c.AttachedTo)
	}
}

// --- Become Anonymous ---------------------------------------------------------

func TestGapB41BecomeAnonymousCloaksTheTargetAndTwoCardsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	top1 := effectstest.PushLibraryTyped(me, "Top One", "Creature — Elf")
	top2 := effectstest.PushLibraryTyped(me, "Top Two", "Instant")
	effectstest.CastCatalogSpell(t, g, "Become Anonymous", "Instant", gapb41AnonymousOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	assassin := effectstest.B12Creature(g, me.ID, "Old Assassin", "Creature — Human Assassin", 1, 1)
	saga := effectstest.CastCatalogSpell(t, g, "Origin of the Hidden Ones", "Enchantment — Saga", gapb41OriginOracle, nil)
	_ = saga
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("chapter I: life %d -> %d, want -4", life, opp.Life)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(assassin, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.SncSettle(t, g, me.ID, uuid.Nil)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.PushLibraryTyped(me, "Relic", "Artifact")
	effectstest.PushLibraryTyped(me, "Bear", "Creature — Bear")
	effectstest.PushLibraryTyped(me, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Ballad of the Black Flag", "Enchantment — Saga", gapb41BalladOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, relic) {
		t.Error("the historic card was not put into hand")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Relic" {
			t.Error("the taken card stayed in the graveyard")
		}
	}
}

func TestGapB41BalladChapterFourDiscountsHistoricSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.CastCatalogSpell(t, g, "Ballad of the Black Flag", "Enchantment — Saga", gapb41BalladOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B22SetLore(g, mustBattlefieldPtr(t, g, saga), 3)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	relic := effectstest.HandSpell(me, "Test Relic", "Artifact", "{3}")
	bauble := effectstest.HandSpell(me, "Test Bauble", "Instant", "{3}")
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, relic, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a historic {3} spell with one mana after chapter IV: %v", err)
	}
	effectstest.FloatForTest(g, me, "C")
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	myRelic := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Fall of the First Civilization", "Enchantment — Saga", gapb41FallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine+2 || opp.Hand.Size() != theirs+2 {
		t.Errorf("hands %d->%d, %d->%d, want +2 each", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickTriggerTarget(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("chapter II did not exile the opponent's artifact")
	}
	if !g.Battlefield.Contains(myRelic) {
		t.Error("chapter II exiled my own artifact")
	}
}

func TestGapB41FallOfTheFirstCivilizationChapterThreeKeepsThreeEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	var mine, theirs []uuid.UUID
	for i := 0; i < 5; i++ {
		mine = append(mine, effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1))
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1))
	}
	land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	saga := effectstest.CastCatalogSpell(t, g, "Fall of the First Civilization", "Enchantment — Saga", gapb41FallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B22SetLore(g, mustBattlefieldPtr(t, g, saga), 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	// Lore 3 is chapter III directly.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, mine[0], mine[1], mine[2])
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[0], theirs[1], theirs[2])
	effectstest.PassPriorityAroundTable(t, g)
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
