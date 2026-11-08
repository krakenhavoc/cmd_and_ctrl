package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_hl_cards_test.go — Edge of Eternities cards whose names start with H
// to L, and one E–G leftover (#4461). Each test plays the real catalog
// entry and asserts what its text says, and what it refuses.

const (
	eoeHLCadetOracle        = "6846448b-d5a0-4ae0-bb11-ea7fa6ed2558"
	eoeHLGuidedOracle       = "9b0af3ad-1b4d-4480-af91-fe9baee5d1c5"
	eoeHLGrovestriderOracle = "0ebec97e-3cb1-41e2-a693-0febb5623016"
	eoeHLMiteOracle         = "103e0b63-fbde-4054-91f7-dd04b40b59ef"
	eoeHLKnightCaptOracle   = "69741d3b-1259-43f5-897f-6f4c3ca0844b"
	eoeHLInfiltratorOracle  = "3084377f-73fb-4b00-9dd9-3844b1ecafc2"
	eoeHLJammerOracle       = "cbc0b136-544d-45a6-be3e-cd21a573bf19"
	eoeHLOperativeOracle    = "2e0b225a-d4b4-4f53-8fea-8c6c37102920"
	eoeHLTurbodroneOracle   = "3ebb8542-5d1d-4f0b-92a9-5a98425860fd"
	eoeHLLashwhipOracle     = "56035e66-e730-436d-8de2-4ded7fd1b51e"
	eoeHLEvangelOracle      = "b162ccf8-e0c2-46ae-84e7-dff2b6fceec1"
	eoeHLLostOracle         = "cec95b9f-9d1e-4988-befe-d0c9e20e428d"
	eoeHLFrigateOracle      = "3aa0cda4-fb2d-4fd4-84f6-482089a85547"
	eoeHLBreacherOracle     = "e0a508b5-0896-4620-a93e-60b958829703"
	eoeHLBattleshipOracle   = "cce3dcc3-57bb-4b95-8b70-337c67bb3c4e"
)

func TestEoeHLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeHLCadetOracle:        "Haliya, Ascendant Cadet",
		eoeHLGuidedOracle:       "Haliya, Guided by Light",
		eoeHLGrovestriderOracle: "Harmonious Grovestrider",
		eoeHLMiteOracle:         "Hemosymbic Mite",
		eoeHLKnightCaptOracle:   "Honored Knight-Captain",
		eoeHLInfiltratorOracle:  "Illvoi Infiltrator",
		eoeHLJammerOracle:       "Illvoi Light Jammer",
		eoeHLOperativeOracle:    "Illvoi Operative",
		eoeHLTurbodroneOracle:   "Kavaron Turbodrone",
		eoeHLLashwhipOracle:     "Lashwhip Predator",
		eoeHLEvangelOracle:      "Lightless Evangel",
		eoeHLLostOracle:         "Lost in Space",
		eoeHLFrigateOracle:      "Lumen-Class Frigate",
		eoeHLBreacherOracle:     "Luxknight Breacher",
		eoeHLBattleshipOracle:   "Extinguisher Battleship",
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
	if ac := game.AlternativeCostByKey(eoeHLGuidedOracle, "warp"); ac == nil || ac.ManaCost != "{W}" {
		t.Errorf("Haliya, Guided by Light: warp %+v, want {W}", ac)
	}
	if spec, _ := Lookup(eoeHLBattleshipOracle); !spec.Purpose.DeclaresSweep() {
		t.Error("Extinguisher Battleship declares a sweep")
	}
}

// --- helpers ---------------------------------------------------------------

// eoeHLAttackAndPick declares attackers against defender, locks them in,
// settles the attack triggers and answers a target prompt with pick.
func eoeHLAttackAndPick(t *testing.T, g *game.Game, defender, pick uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, g.Seats[g.Turn.ActiveSeat].ID, pick)
	passPriorityAroundTable(t, g)
}

// eoeHLRefusePick asserts that the chooser's open target prompt refuses id
// as the pick.
func eoeHLRefusePick(t *testing.T, g *game.Game, chooser, id uuid.UUID, what string) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
		t.Fatalf("%s was accepted as the target", what)
	}
}

// --- Haliya, Ascendant Cadet -------------------------------------------------

func TestEoeHLHaliyaAscendantCadetPutsACounterAsSheEntersAndAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	haliya := castCatalogSpell(t, g, "Haliya, Ascendant Cadet", "Legendary Creature — Human Soldier", eoeHLCadetOracle, nil)
	passPriorityAroundTable(t, g)
	eoeHLRefusePick(t, g, me.ID, theirs, "an opponent's creature")
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("on entry: %d counters on your creature, want 1", got)
	}

	findBattlefieldCardForTest(g, haliya).SummonedThisTurn = false
	eoeHLAttackAndPick(t, g, opp.ID, bear, haliya)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("on attack: %d counters, want 2", got)
	}
}

func TestEoeHLHaliyaAscendantCadetDrawsOnlyForCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	b12Push(g, me.ID, "Haliya, Ascendant Cadet", "Legendary Creature — Human Soldier", eoeHLCadetOracle, 3, 3)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	grown := b12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(grown, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, plain, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("a creature with no counter connecting drew: hand %d -> %d", hand, got)
	}
	dealCombatDamageToPlayer(g, grown, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a creature with a +1/+1 counter connecting: hand %d -> %d, want +1", hand, got)
	}
}

// --- Haliya, Guided by Light -------------------------------------------------

func TestEoeHLHaliyaGuidedByLightGainsForYourCreaturesAndArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	castCatalogSpell(t, g, "Haliya, Guided by Light", "Legendary Creature — Human Soldier", eoeHLGuidedOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("Haliya entering: life %d -> %d, want +1", life, me.Life)
	}
	avB1OpponentCreatureEnters(t, g, me.ID, "Artifact — Clue")
	if me.Life != life+2 {
		t.Fatalf("your artifact entering: life %d, want %d", me.Life, life+2)
	}
	avB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Bear")
	if me.Life != life+2 {
		t.Errorf("an opponent's creature changed your life to %d", me.Life)
	}
	avB1OpponentCreatureEnters(t, g, me.ID, "Enchantment")
	if me.Life != life+2 {
		t.Errorf("your enchantment changed your life to %d", me.Life)
	}
}

func TestEoeHLHaliyaGuidedByLightDrawsAtYourEndStepAfterThreeLife(t *testing.T) {
	for _, tc := range []struct {
		gain, draw int
	}{{2, 0}, {3, 1}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedLibrary(me, "A", "B", "C")
		haliya := b12Push(g, me.ID, "Haliya, Guided by Light", "Legendary Creature — Human Soldier", eoeHLGuidedOracle, 3, 3)
		advanceToMain(t, g)
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(haliya, me.ID, tc.gain); err != nil {
				t.Fatal(err)
			}
		})
		hand := me.Hand.Size()
		advanceThroughEndStep(t, g)
		if got := me.Hand.Size() - hand; got != tc.draw {
			t.Errorf("gained %d this turn: drew %d at the end step, want %d", tc.gain, got, tc.draw)
		}
	}
}

// --- Harmonious Grovestrider -------------------------------------------------

func TestEoeHLHarmoniousGrovestriderIsTheSizeOfYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strider := b12Push(g, me.ID, "Harmonious Grovestrider", "Creature — Beast", eoeHLGrovestriderOracle, 0, 0)
	for i := 0; i < 3; i++ {
		pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if p, tough := eoeCDPT(t, g, strider); p != 3 || tough != 3 {
		t.Fatalf("with three lands it is %d/%d, want 3/3", p, tough)
	}
	advanceToMain(t, g)
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	if p, tough := eoeCDPT(t, g, strider); p != 4 || tough != 4 {
		t.Errorf("with four lands it is %d/%d, want 4/4", p, tough)
	}
	if spec, _ := Lookup(eoeHLGrovestriderOracle); len(spec.Triggered) != 1 {
		t.Errorf("ward {2} is its one trigger, got %d", len(spec.Triggered))
	}
}

// --- Hemosymbic Mite ---------------------------------------------------------

func TestEoeHLHemosymbicMitePumpsAnotherCreatureWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mite := b12Push(g, me.ID, "Hemosymbic Mite", "Creature — Mite", eoeHLMiteOracle, 1, 1)
	if err := g.AddCounter(mite, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	eoeHLAttackAndPick(t, g, opp.ID, bear, mite)
	if p, tough := eoeCDPT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("your other creature is %d/%d, want 4/4 (+X/+X for the Mite's power of 2)", p, tough)
	}
	if p, _ := eoeCDPT(t, g, mite); p != 2 {
		t.Errorf("the Mite pumped itself: power %d", p)
	}
}

// --- Honored Knight-Captain --------------------------------------------------

func TestEoeHLHonoredKnightCaptainMakesASoldierAndFetchesAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	captain := castCatalogSpell(t, g, "Honored Knight-Captain", "Creature — Human Advisor Knight", eoeHLKnightCaptOracle, nil)
	passPriorityAroundTable(t, g)
	if got := eoeCDCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Fatalf("Human Soldier tokens %d, want 1", got)
	}
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Some Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})

	fillPoolColored(me, "W", 5)
	if err := eoeCDActivate(g, me.ID, captain, 0); err == nil {
		t.Fatal("the ability was paid with five mana")
	}
	fillPoolColored(me, "W", 1)
	if err := eoeCDActivate(g, me.ID, captain, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(captain) {
		t.Error("the Knight-Captain is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Some Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is offered")
	}
	answerSearchNamed(t, g, me.ID, "Bonesplitter")
	passPriorityAroundTable(t, g)
	if eoeCDCount(g, me.ID, "Bonesplitter") != 1 {
		t.Error("the Equipment is put onto the battlefield")
	}
}

// --- Illvoi Infiltrator ------------------------------------------------------

func TestEoeHLIllvoiInfiltratorCanBeBlockedAfterOneSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jelly := b12Push(g, me.ID, "Illvoi Infiltrator", "Creature — Jellyfish Rogue", eoeHLInfiltratorOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	pr7bAttack(t, g, opp.ID, jelly)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, jelly); err != nil {
		t.Errorf("after one spell it can be blocked: %v", err)
	}
}

func TestEoeHLIllvoiInfiltratorIsUnblockableAfterTwoSpellsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	jelly := b12Push(g, me.ID, "Illvoi Infiltrator", "Creature — Jellyfish Rogue", eoeHLInfiltratorOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	for _, name := range []string{"First", "Second"} {
		b24CastNoncreature(t, g, name, "Instant", "{U}")
		passPriorityAroundTable(t, g)
	}
	pr7bAttack(t, g, opp.ID, jelly)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, jelly); err == nil {
		t.Fatal("after two spells this turn it can't be blocked")
	}
	finishBlockDeclarations(t, g)
	passPriorityAroundTable(t, g)
	hand, life := me.Hand.Size(), opp.Life
	pr7bDamageStep(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("it deals 1: %d -> %d", life, opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on combat damage to a player, want 1", got)
	}
}

// --- Illvoi Light Jammer -----------------------------------------------------

func TestEoeHLIllvoiLightJammerAttachesWithHexproofAndEquipsForThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	jammer := avB1Cast(t, g, me, game.Card{Name: "Illvoi Light Jammer", TypeLine: "Artifact — Equipment", ManaCost: "{1}{U}",
		OracleID: eoeHLJammerOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !avB1Card(t, g, jammer).IsAttachedTo(bear) {
		t.Fatal("the Jammer is attached to your creature as it enters")
	}
	if p, tough := eoeCDPT(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("equipped creature is %d/%d, want 3/4", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "hexproof") {
		t.Error("the creature gains hexproof until end of turn")
	}

	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "C", 2)
	if err := eoeCDActivate(g, me.ID, jammer, 0, eoeCDCard(other)); err == nil {
		t.Fatal("equip was paid with two mana")
	}
	fillPoolColored(me, "C", 1)
	if err := eoeCDActivate(g, me.ID, jammer, 0, eoeCDCard(other)); err != nil {
		t.Fatalf("equip {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB1Card(t, g, jammer).IsAttachedTo(other) {
		t.Error("equip moves the Jammer")
	}
	if p, tough := eoeCDPT(t, g, other); p != 3 || tough != 4 {
		t.Errorf("the new host is %d/%d, want 3/4", p, tough)
	}
	if hasKeywordOnBattlefield(t, g, other, "hexproof") {
		t.Error("equipping later grants no hexproof")
	}
}

// --- Illvoi Operative --------------------------------------------------------

func TestEoeHLIllvoiOperativeGrowsOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	op := b12Push(g, me.ID, "Illvoi Operative", "Creature — Jellyfish Rogue", eoeHLOperativeOracle, 2, 1)
	for i, want := range []int{0, 1, 1} {
		castCatalogSpell(t, g, "Sami's Curiosity", "Sorcery", eoeQZSamisOracle, nil)
		passPriorityAroundTable(t, g)
		if got := countersOn(g, op, game.CounterPlusOne); got != want {
			t.Errorf("after spell %d: %d counters, want %d", i+1, got, want)
		}
	}
}

// --- Kavaron Turbodrone ------------------------------------------------------

func TestEoeHLKavaronTurbodronePumpsAndHastesYourCreatureAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drone := b12Push(g, me.ID, "Kavaron Turbodrone", "Artifact Creature — Robot Scout", eoeHLTurbodroneOracle, 2, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := eoeCDActivate(g, me.ID, drone, 0, eoeCDCard(theirs)); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	if err := eoeCDActivate(g, me.ID, drone, 0, eoeCDCard(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("your creature gains haste")
	}
	if !avB1Card(t, g, drone).Tapped {
		t.Error("the Turbodrone taps to pay")
	}
}

func TestEoeHLKavaronTurbodroneCannotBeActivatedInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drone := b12Push(g, me.ID, "Kavaron Turbodrone", "Artifact Creature — Robot Scout", eoeHLTurbodroneOracle, 2, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if err := eoeCDActivate(g, me.ID, drone, 0, eoeCDCard(bear)); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Lashwhip Predator -------------------------------------------------------

func TestEoeHLLashwhipPredatorCostsTwoLessAgainstThreeOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	if mv := selfPricedMV(t, g, me, eoeHLLashwhipOracle, "Creature — Plant Beast", "{4}{G}{G}"); mv != 6 {
		t.Fatalf("against two opposing creatures it costs %d, want 6", mv)
	}
	b12Creature(g, third.ID, "Third Wolf", "Creature — Wolf", 2, 2)
	if mv := selfPricedMV(t, g, me, eoeHLLashwhipOracle, "Creature — Plant Beast", "{4}{G}{G}"); mv != 4 {
		t.Errorf("against three creatures across your opponents it costs %d, want 4", mv)
	}
}

// --- Lightless Evangel -------------------------------------------------------

func TestEoeHLLightlessEvangelGrowsWhenYouSacrificeACreatureOrArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	evangel := b12Push(g, me.ID, "Lightless Evangel", "Creature — Vampire Cleric", eoeHLEvangelOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, step := range []struct {
		id   uuid.UUID
		want int
		what string
	}{
		{land, 0, "your land"},
		{theirs, 0, "an opponent's creature"},
		{bear, 1, "your creature"},
		{rock, 2, "your artifact"},
	} {
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(step.id) })
		passPriorityAroundTable(t, g)
		if got := countersOn(g, evangel, game.CounterPlusOne); got != step.want {
			t.Errorf("after sacrificing %s: %d counters, want %d", step.what, got, step.want)
		}
	}
}

// --- Lost in Space -----------------------------------------------------------

func TestEoeHLLostInSpaceTucksForTheOwnerThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, bad := range []uuid.UUID{land, ench} {
		if err := castCatalogSpellErr(t, g, "Lost in Space", "Instant", eoeHLLostOracle, cardRefs(bad)); err == nil {
			t.Fatal("a target that is neither an artifact nor a creature was accepted")
		}
	}
	castCatalogSpell(t, g, "Lost in Space", "Instant", eoeHLLostOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("the surveil came before the owner answered")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(bear) || !opp.Library.Contains(bear) {
		t.Fatal("the creature is not in its owner's library")
	}
	eoeQZKeepOnTop(t, g, me.ID, eoeQZSurveil(t, g, me.ID, 1))
}

// --- Lumen-Class Frigate -----------------------------------------------------

func TestEoeHLLumenClassFrigatePumpsOthersAtTwoAndFliesAtTwelve(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Push(g, me.ID, "Lumen-Class Frigate", "Artifact — Spacecraft", eoeHLFrigateOracle, 3, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if p, _ := eoeCDPT(t, g, bear); p != 2 {
		t.Fatalf("at one charge counter your creature has power %d, want 2", p)
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if p, tough := eoeCDPT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("at two your creature is %d/%d, want 3/3", p, tough)
	}
	if p, _ := eoeCDPT(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is pumped: power %d", p)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("below 12 it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 10); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at12 := layeredCard(t, g, ship)
	if !at12.IsCreature() || !game.HasKeyword(&at12, "flying") || !game.HasKeyword(&at12, "lifelink") {
		t.Error("at 12+ it is a creature with flying and lifelink")
	}
	if p, tough := eoeCDPT(t, g, ship); p != 3 || tough != 5 {
		t.Errorf("at 12+ it is %d/%d, want 3/5: it doesn't pump itself", p, tough)
	}
}

// --- Luxknight Breacher ------------------------------------------------------

func TestEoeHLLuxknightBreacherEntersWithACounterPerCreatureAndArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	b12Creature(g, me.ID, "My Construct", "Artifact Creature — Construct", 1, 1)
	pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	knight := castCatalogSpell(t, g, "Luxknight Breacher", "Creature — Human Knight", eoeHLBreacherOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, knight, game.CounterPlusOne); got != 3 {
		t.Errorf("enters with %d counters, want 3 (an artifact creature counts once)", got)
	}
}

// --- Extinguisher Battleship -------------------------------------------------

func TestEoeHLExtinguisherBattleshipDestroysThenDamagesEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	titan := b12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 6, 6)
	ship := avB1Cast(t, g, me, game.Card{Name: "Extinguisher Battleship", TypeLine: "Artifact — Spacecraft", ManaCost: "{8}",
		OracleID: eoeHLBattleshipOracle, Power: 10, Toughness: 10}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	eoeHLRefusePick(t, g, me.ID, titan, "a creature")
	pickCard(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(aura) {
		t.Error("the noncreature permanent is destroyed")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("4 damage kills your own 2/2 too")
	}
	if !g.Battlefield.Contains(titan) {
		t.Error("a 6/6 survives 4 damage")
	}
	if !g.Battlefield.Contains(ship) {
		t.Fatal("the Battleship is not a creature as it enters and survives")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 5); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at5 := layeredCard(t, g, ship)
	if !at5.IsCreature() || !game.HasKeyword(&at5, "flying") || !game.HasKeyword(&at5, "trample") ||
		at5.CurrentPower() != 10 || at5.CurrentToughness() != 10 {
		t.Errorf("at 5+ it is a 10/10 with flying and trample, got %d/%d", at5.CurrentPower(), at5.CurrentToughness())
	}
}
