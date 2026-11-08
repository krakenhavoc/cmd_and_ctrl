package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_dl2_cards_test.go — Murders at Karlov Manor cards D–L, second batch
// (#4469). Each test plays the real catalog entry and asserts what its
// text says. The face-down helpers are the M–P batch's (mkmMPCastFaceDown,
// mkmMPTurnFaceUp, mkmMPPT).

const (
	mkmDL2DragOracle      = "0c0ffb66-94db-4ba9-afde-c5b4fbf09578"
	mkmDL2EssenceOracle   = "26a7f92d-6914-466d-b7c7-9ac121dd5118"
	mkmDL2EzrimOracle     = "ddcd8a0a-7ac3-4400-a16f-8bb5aa50994e"
	mkmDL2FaeFlightOracle = "d3912c82-37f8-456e-ba49-65c7f5b39d13"
	mkmDL2SnoopOracle     = "b759e8a6-8dd0-48ce-8816-e45b35a79895"
	mkmDL2LeechOracle     = "d14fdc1d-62bc-4e9c-b79b-8c9c3232f0bc"
	mkmDL2GalvanizeOracle = "9b44c005-6799-47fd-8b0d-7179b5281d78"
	mkmDL2LegUpOracle     = "a80519c4-f778-4b8e-9920-64013e98c95b"
	mkmDL2GeardrakeOracle = "af294df7-f8a0-4d52-88cb-086678463970"
	mkmDL2BonebruteOracle = "c1d9c41a-0941-4b7e-b8a6-890d30386458"
	mkmDL2AnalystOracle   = "e43400fe-3df5-4942-abc3-b9bba34fc7cd"
	mkmDL2KnifeOracle     = "7627163a-1689-446f-b5aa-a6c56aefb532"
	mkmDL2PipeOracle      = "ed2ef95e-8851-4226-9c63-4c19ea323b82"
	mkmDL2OnlookerOracle  = "06a37f16-c929-4b89-bf18-1b3f8f472eef"
	mkmDL2LoxodonOracle   = "141a2757-8f20-4e10-9866-bb5d612ed62d"
)

func TestMkmDL2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmDL2DragOracle:      "Drag the Canal",
		mkmDL2EssenceOracle:   "Essence of Antiquity",
		mkmDL2EzrimOracle:     "Ezrim, Agency Chief",
		mkmDL2FaeFlightOracle: "Fae Flight",
		mkmDL2SnoopOracle:     "Faerie Snoop",
		mkmDL2LeechOracle:     "Festerleech",
		mkmDL2GalvanizeOracle: "Galvanize",
		mkmDL2LegUpOracle:     "Get a Leg Up",
		mkmDL2GeardrakeOracle: "Gleaming Geardrake",
		mkmDL2BonebruteOracle: "Hunted Bonebrute",
		mkmDL2AnalystOracle:   "Jaded Analyst",
		mkmDL2KnifeOracle:     "Knife",
		mkmDL2PipeOracle:      "Lead Pipe",
		mkmDL2OnlookerOracle:  "Leering Onlooker",
		mkmDL2LoxodonOracle:   "Loxodon Eavesdropper",
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
}

// TestMkmDL2DisguiseDeclarationsAreWired pins each disguise card's
// keyword: the {3} face-down cast, the disguise kind (ward {2}) and the
// card's own face-up cost.
func TestMkmDL2DisguiseDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Essence of Antiquity", mkmDL2EssenceOracle, "{2}{W}"},
		{"Faerie Snoop", mkmDL2SnoopOracle, "{1}{U/B}{U/B}"},
		{"Hunted Bonebrute", mkmDL2BonebruteOracle, "{1}{B}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "disguise")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no disguise cast", tc.name)
			continue
		}
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, MorphCastCost, tc.faceUp)
		}
	}
}

// mkmDL2Draw has p draw one card and lets the stack settle.
func mkmDL2Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	passPriorityAroundTable(t, g)
}

// mkmDL2DrawUpTo has p draw until they have drawn n cards this turn (the
// active player may already have drawn for the turn), failing if they
// are already past n.
func mkmDL2DrawUpTo(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	if got := g.TurnTallyFor(p.ID).CardsDrawn; got > n {
		t.Fatalf("%s has drawn %d cards this turn, past %d", p.Name, got, n)
	}
	for g.TurnTallyFor(p.ID).CardsDrawn < n {
		mkmDL2Draw(t, g, p)
	}
}

// --- Drag the Canal ------------------------------------------------------------

func TestMkmDL2DragTheCanalWithNoDeathMakesOnlyADetective(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	life := me.Life
	castCatalogSpell(t, g, "Drag the Canal", "Instant", mkmDL2DragOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
	if surveilChoiceFor(g, me.ID) != nil || me.Life != life || tmtB1Count(g, me.ID, "Clue") != 0 {
		t.Errorf("with no creature dead: surveil %v, life %d (was %d), %d Clues; want none of it",
			surveilChoiceFor(g, me.ID) != nil, me.Life, life, tmtB1Count(g, me.ID, "Clue"))
	}
}

func TestMkmDL2DragTheCanalAfterAnyDeathGainsSurveilsThenInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(me, "One", "Two", "Three")
	// An opponent's creature dying is enough: "a creature died this turn".
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	life := me.Life

	castCatalogSpell(t, g, "Drag the Canal", "Instant", mkmDL2DragOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil after a creature died")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	if tmtB1Count(g, me.ID, "Clue") != 0 {
		t.Error("the Clue came before the surveil was answered (CR 608.2c)")
	}
	looked := append([]uuid.UUID(nil), c.ScryCards...)
	if err := g.ResolveSurveil(c.ID, me.ID, looked, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after the surveil, want 1", got)
	}
}

// --- Essence of Antiquity ------------------------------------------------------

func TestMkmDL2EssenceOfAntiquityHexproofsAndUntapsYourTeamWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.TapTargetForEffect(mine)
		_ = g.TapTargetForEffect(theirs)
	})
	id := mkmMPCastFaceDown(t, g, "Essence of Antiquity", "Artifact Creature — Golem", mkmDL2EssenceOracle, 1, 10)
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Fatal("entering face down gave hexproof (CR 708.3)")
	}

	mkmMPTurnFaceUp(t, g, me, id, "W", "C", "C")
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
	}{{"my bear", mine}, {"the Essence", id}} {
		if !hasEffectiveKeyword(t, g, c.id, "hexproof") {
			t.Errorf("%s has no hexproof", c.name)
		}
		if b16Tapped(t, g, c.id) {
			t.Errorf("%s is still tapped", c.name)
		}
	}
	if hasEffectiveKeyword(t, g, theirs, "hexproof") || !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature was given hexproof or untapped")
	}
	if p, tough := mkmMPPT(t, g, id); p != 1 || tough != 10 {
		t.Errorf("face up the Essence is %d/%d, want 1/10", p, tough)
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("hexproof outlasted the turn")
	}
}

// --- Ezrim, Agency Chief -------------------------------------------------------

func TestMkmDL2EzrimInvestigatesTwiceAndSacrificesAnArtifactForAKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ezrim := castHoldCreature(t, g, "Ezrim, Agency Chief", "Legendary Creature — Archon Detective", mkmDL2EzrimOracle, 5, 5)
	passPriorityAroundTable(t, g)
	clues := tmtB1Named(g, "Clue")
	if len(clues) != 2 {
		t.Fatalf("%d Clues, want 2", len(clues))
	}
	assertKeywords(t, g, ezrim, "flying")

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ezrim, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is not an artifact was accepted as the sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, ezrim, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{clues[0]}}); err != nil {
		t.Fatalf("activate with a Clue: %v", err)
	}
	if g.Battlefield.Contains(clues[0]) {
		t.Error("the Clue was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 2)
	if !hasEffectiveKeyword(t, g, ezrim, "hexproof") {
		t.Error("choosing hexproof gave Ezrim no hexproof")
	}
	if hasEffectiveKeyword(t, g, ezrim, "lifelink") || hasEffectiveKeyword(t, g, ezrim, "vigilance") {
		t.Error("Ezrim gained a keyword that was not chosen")
	}
}

// --- Fae Flight ----------------------------------------------------------------

func TestMkmDL2FaeFlightFlashesOnAndGrantsFlyingAndHexproofForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	// "Enchant creature": any creature, an opponent's included.
	aura, err := ecl2CastAt(t, g, "Fae Flight", "Enchantment — Aura", mkmDL2FaeFlightOracle, b16TargetCard(theirs))
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	if p, tt := ptOf(t, g, theirs); p != 3 || tt != 2 {
		t.Errorf("enchanted creature is %d/%d, want 3/2", p, tt)
	}
	for _, kw := range []string{"flying", "hexproof"} {
		if !hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("enchanted creature has no %s", kw)
		}
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("hexproof lasts only until end of turn")
	}
	if !hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("flying is the Aura's static and stays")
	}
}

// --- Faerie Snoop --------------------------------------------------------------

func TestMkmDL2FaerieSnoopLooksAtTwoWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Faerie Snoop", "Creature — Faerie Detective", mkmDL2SnoopOracle, 1, 4)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Snoop flies (CR 708.2a)")
	}
	top := seedLibrary(me, "One", "Two", "Three")
	mkmMPTurnFaceUp(t, g, me, id, "U", "B", "C")
	passPriorityAroundTable(t, g)
	c := latestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("turning the Snoop face up asked for no card")
	}
	if len(c.ChooseCards) != 2 {
		t.Fatalf("the prompt offers %d cards, want the top two", len(c.ChooseCards))
	}
	keep, other := c.ChooseCards[0], c.ChooseCards[1]
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err == nil {
		t.Error("putting no card into hand was accepted: \"put one\" is mandatory")
	}
	answerChooseCards(t, g, me.ID, keep)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) || !me.Graveyard.Contains(other) {
		t.Error("the picked card is not in hand, or the other is not in the graveyard")
	}
	for _, untouched := range top {
		if untouched != keep && untouched != other && !me.Library.Contains(untouched) {
			t.Error("a card below the top two left the library")
		}
	}
	assertKeywords(t, g, id, "flying")
}

// --- Festerleech ---------------------------------------------------------------

func TestMkmDL2FesterleechMillsTwoOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leech := b12Push(g, me.ID, "Festerleech", "Creature — Zombie Leech", mkmDL2LeechOracle, 1, 1)
	seedLibrary(me, "One", "Two", "Three")
	library, graveyard := me.Library.Size(), me.Graveyard.Size()

	declareAttack(t, g, opp.ID, leech)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - graveyard; got != 2 {
		t.Errorf("milled %d cards, want 2", got)
	}
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("the library shrank by %d, want 2", got)
	}
}

func TestMkmDL2FesterleechPumpsOnlyOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	leech := b12Push(g, me.ID, "Festerleech", "Creature — Zombie Leech", mkmDL2LeechOracle, 1, 1)
	floatMana(t, g, me, "{B}{B}{C}{C}")
	p7Activate(t, g, me, leech, 0, game.ActivateAbilityParams{Strict: true})
	if p, tough := mkmMPPT(t, g, leech); p != 3 || tough != 3 {
		t.Errorf("the Leech is %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, leech, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in one turn was accepted")
	}
}

// --- Galvanize -----------------------------------------------------------------

func TestMkmDL2GalvanizeDealsThreeOrFiveAfterTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	four := b12Creature(g, opp.ID, "Their Four", "Creature — Elk", 4, 4)
	castCatalogSpell(t, g, "Galvanize", "Instant", mkmDL2GalvanizeOracle, []game.TargetRef{cardRef(four)})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(four) || damageMarkedOn(g, four) != 3 {
		t.Fatalf("with no draws: on the battlefield %v, %d damage; want it alive with 3",
			g.Battlefield.Contains(four), damageMarkedOn(g, four))
	}

	seedLibrary(me, "One", "Two", "Three")
	mkmDL2DrawUpTo(t, g, me, 2)
	five := b12Creature(g, opp.ID, "Their Five", "Creature — Elk", 5, 5)
	castCatalogSpell(t, g, "Galvanize", "Instant", mkmDL2GalvanizeOracle, []game.TargetRef{cardRef(five)})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(five) {
		t.Error("after two draws a 5/5 survived: Galvanize should deal 5")
	}
}

// --- Get a Leg Up --------------------------------------------------------------

func TestMkmDL2GetALegUpPumpsPerCreatureYouControlAndGrantsReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	castCatalogSpell(t, g, "Get a Leg Up", "Instant", mkmDL2LegUpOracle, []game.TargetRef{cardRef(bear)})
	passPriorityAroundTable(t, g)
	// Two creatures I control, the target among them; the opponent's Elk
	// does not count.
	if p, tough := mkmMPPT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the target has no reach")
	}
}

// --- Gleaming Geardrake --------------------------------------------------------

func TestMkmDL2GleamingGeardrakeInvestigatesAndGrowsOnASacrificedArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	drake := castHoldCreature(t, g, "Gleaming Geardrake", "Artifact Creature — Drake", mkmDL2GeardrakeOracle, 1, 1)
	passPriorityAroundTable(t, g)
	clues := tmtB1Named(g, "Clue")
	if len(clues) != 1 {
		t.Fatalf("%d Clues, want 1", len(clues))
	}
	assertKeywords(t, g, drake, "flying")

	seedLibrary(me, "One", "Two")
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, clues[0], 0, game.ActivateAbilityParams{Strict: true})
	if got := counterCount(g, drake, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after sacrificing a Clue, want 1", got)
	}

	// An opponent's sacrifice is not "you sacrifice".
	theirs := pushToken(g, opp.ID, ClueToken())
	floatMana(t, g, opp, "{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the opponent cracks a Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drake, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after an opponent's sacrifice, want it to stay 1", got)
	}
}

// --- Hunted Bonebrute ----------------------------------------------------------

func TestMkmDL2HuntedBonebruteGivesAnOpponentDogsAndDrainsOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	brute := castHoldCreature(t, g, "Hunted Bonebrute", "Creature — Skeleton Beast", mkmDL2BonebruteOracle, 6, 2)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, opp.ID, "Dog"); got != 2 {
		t.Errorf("the opponent controls %d Dogs, want 2", got)
	}
	if got := tmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Errorf("I control %d Dogs, want 0", got)
	}
	assertKeywords(t, g, brute, "menace")

	mine := me.Life
	before := lifeOfOpponents(g)
	killCreature(t, g, me.ID, brute)
	passPriorityAroundTable(t, g)
	for i, life := range lifeOfOpponents(g) {
		if life != before[i]-3 {
			t.Errorf("opponent %d is at %d, want %d", i, life, before[i]-3)
		}
	}
	if me.Life != mine {
		t.Errorf("my life moved: %d, was %d", me.Life, mine)
	}
}

func TestMkmDL2HuntedBonebruteFaceDownHandsOutNoDogs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Hunted Bonebrute", "Creature — Skeleton Beast", mkmDL2BonebruteOracle, 6, 2)
	if latestPickTarget(g, me.ID) != nil || len(tmtB1Named(g, "Dog")) != 0 {
		t.Fatal("a face-down Bonebrute triggered on entering (CR 708.3)")
	}
	mkmMPTurnFaceUp(t, g, me, id, "B", "C")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || len(tmtB1Named(g, "Dog")) != 0 {
		t.Error("turning the Bonebrute face up counted as entering")
	}
	if p, tough := mkmMPPT(t, g, id); p != 6 || tough != 2 {
		t.Errorf("face up the Bonebrute is %d/%d, want 6/2", p, tough)
	}
}

// --- Jaded Analyst -------------------------------------------------------------

func TestMkmDL2JadedAnalystLosesDefenderOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	analyst := b12Push(g, me.ID, "Jaded Analyst", "Creature — Human Detective", mkmDL2AnalystOracle, 3, 2)
	assertKeywords(t, g, analyst, "defender")
	seedLibrary(me, "One", "Two", "Three")

	mkmDL2DrawUpTo(t, g, me, 1)
	if !hasEffectiveKeyword(t, g, analyst, "defender") {
		t.Fatal("the first draw removed defender")
	}
	mkmDL2DrawUpTo(t, g, me, 2)
	if hasEffectiveKeyword(t, g, analyst, "defender") {
		t.Error("the second draw left defender")
	}
	if !hasEffectiveKeyword(t, g, analyst, "vigilance") {
		t.Error("the second draw gave no vigilance")
	}
	declareAttack(t, g, opp.ID, analyst)
	if b16Tapped(t, g, analyst) {
		t.Error("the Analyst tapped to attack despite vigilance")
	}
}

func TestMkmDL2JadedAnalystCantAttackBeforeTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	analyst := b12Push(g, me.ID, "Jaded Analyst", "Creature — Human Detective", mkmDL2AnalystOracle, 3, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(analyst, opp.ID); err == nil {
		t.Error("a creature with defender was declared as an attacker")
	}
}

// --- Knife ---------------------------------------------------------------------

func TestMkmDL2KnifeGivesFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	knife := b12Push(g, me.ID, "Knife", "Artifact — Clue Equipment", mkmDL2KnifeOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, knife, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(bear)}})
	if !attachedTo(t, g, knife, bear) {
		t.Fatal("Knife is not attached")
	}
	if p, tough := mkmMPPT(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("on my turn the equipped creature is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("on my turn the equipped creature has no first strike")
	}

	endTurn(t, g)
	if p, tough := mkmMPPT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("on an opponent's turn the equipped creature is %d/%d, want 2/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("on an opponent's turn the equipped creature has first strike")
	}
}

func TestMkmDL2KnifeCracksForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	knife := b12Push(g, me.ID, "Knife", "Artifact — Clue Equipment", mkmDL2KnifeOracle, 0, 0)
	seedLibrary(me, "One")
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, knife, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(knife) {
		t.Error("Knife was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Lead Pipe -----------------------------------------------------------------

func TestMkmDL2LeadPipePumpsAndDrainsWhenTheEquippedCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pipe := b12Push(g, me.ID, "Lead Pipe", "Artifact — Clue Equipment", mkmDL2PipeOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, pipe, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(bear)}})
	if !attachedTo(t, g, pipe, bear) {
		t.Fatal("Lead Pipe is not attached")
	}
	if p, tough := mkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped creature is %d/%d, want 4/2", p, tough)
	}

	mine := me.Life
	before := lifeOfOpponents(g)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	for i, life := range lifeOfOpponents(g) {
		if life != before[i]-1 {
			t.Errorf("opponent %d is at %d, want %d", i, life, before[i]-1)
		}
	}
	if me.Life != mine {
		t.Errorf("my life moved: %d, was %d", me.Life, mine)
	}

	seedLibrary(me, "One")
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, pipe, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(pipe) {
		t.Error("Lead Pipe was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Leering Onlooker ----------------------------------------------------------

func TestMkmDL2LeeringOnlookerMakesTwoTappedBatsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	onBattlefield := b12Push(g, me.ID, "Leering Onlooker", "Creature — Vampire", mkmDL2OnlookerOracle, 1, 3)
	assertKeywords(t, g, onBattlefield, "flying")
	floatMana(t, g, me, "{B}{B}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, onBattlefield, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the graveyard ability was activated from the battlefield")
	}

	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Leering Onlooker", TypeLine: "Creature — Vampire",
		OracleID: mkmDL2OnlookerOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("the card was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	bats := tmtB1Named(g, "Bat")
	if len(bats) != 2 {
		t.Fatalf("%d Bats, want 2", len(bats))
	}
	for _, bat := range bats {
		c := tmtB1Card(t, g, bat)
		if !c.Tapped || c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "flying") {
			t.Errorf("a Bat is %d/%d tapped=%v under %s, want my tapped 1/1 flyer", c.Power, c.Toughness, c.Tapped, c.Controller)
		}
	}
}

// --- Loxodon Eavesdropper ------------------------------------------------------

func TestMkmDL2LoxodonEavesdropperInvestigatesAndGrowsOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elephant := castHoldCreature(t, g, "Loxodon Eavesdropper", "Creature — Elephant Detective", mkmDL2LoxodonOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}

	seedLibrary(me, "One", "Two", "Three")
	mkmDL2DrawUpTo(t, g, me, 1)
	if p, tough := mkmMPPT(t, g, elephant); p != 3 || tough != 3 || hasEffectiveKeyword(t, g, elephant, "vigilance") {
		t.Fatalf("after one draw the Eavesdropper is %d/%d (vigilance %v), want a plain 3/3",
			p, tough, hasEffectiveKeyword(t, g, elephant, "vigilance"))
	}
	mkmDL2DrawUpTo(t, g, me, 2)
	if p, tough := mkmMPPT(t, g, elephant); p != 4 || tough != 4 {
		t.Errorf("after the second draw the Eavesdropper is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, elephant, "vigilance") {
		t.Error("after the second draw the Eavesdropper has no vigilance")
	}
	mkmDL2Draw(t, g, me)
	if p, _ := mkmMPPT(t, g, elephant); p != 4 {
		t.Errorf("the third draw changed the Eavesdropper's power to %d, want it to stay 4", p)
	}
}
