package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lgn_batch1_cards_test.go — Legions, batch 1 (#5609, S82): thirteen
// A–C cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	lgn1AkromaOracle         = "107f204c-9fd3-490d-8887-be63d14fc6a4"
	lgn1DevotedOracle        = "d264903b-d23e-48c8-95f1-cb1606a705a2"
	lgn1ExterminatorOracle   = "a3bf5d18-77d5-4881-916c-cc125e621b6e"
	lgn1AvenEnvoyOracle      = "73cd8461-7930-4e2d-9a14-4d10db91e93e"
	lgn1AvenRedeemerOracle   = "faac29f5-10ad-4800-a9d3-2259c4800352"
	lgn1BaneOracle           = "c99e4768-192c-49bf-8f51-c204a2248326"
	lgn1MurlodontOracle      = "a762e8cb-9d33-41fc-b2e8-0c5ad4d7268e"
	lgn1BladeSliverOracle    = "110e5695-a9df-4201-9230-99f6abe9978f"
	lgn1BloodCelebrantOracle = "695ce90b-a5a2-4d8c-ac05-07f04010603e"
	lgn1BloodstokeOracle     = "41fa3954-119b-4ff2-ad56-9105482f6b7c"
	lgn1BranchsnapOracle     = "e09b24a0-675b-4c64-ac44-74ddf109598a"
	lgn1BroodSliverOracle    = "6b1d178f-9713-4dcf-920b-ddf2c94cc427"
	lgn1CallerOracle         = "ec79255f-4b4d-4a0a-89d8-c59ba640275b"
)

func TestLgn1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lgn1AkromaOracle:         "Akroma, Angel of Wrath",
		lgn1DevotedOracle:        "Akroma's Devoted",
		lgn1ExterminatorOracle:   "Aphetto Exterminator",
		lgn1AvenEnvoyOracle:      "Aven Envoy",
		lgn1AvenRedeemerOracle:   "Aven Redeemer",
		lgn1BaneOracle:           "Bane of the Living",
		lgn1MurlodontOracle:      "Berserk Murlodont",
		lgn1BladeSliverOracle:    "Blade Sliver",
		lgn1BloodCelebrantOracle: "Blood Celebrant",
		lgn1BloodstokeOracle:     "Bloodstoke Howler",
		lgn1BranchsnapOracle:     "Branchsnap Lorian",
		lgn1BroodSliverOracle:    "Brood Sliver",
		lgn1CallerOracle:         "Caller of the Claw",
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

// Every morph in the batch casts face down for the keyword's {3} and
// turns face up for the card's own printed cost (CR 702.37a).
func TestLgn1MorphCostsAreThePrintedOnes(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Aphetto Exterminator", lgn1ExterminatorOracle, "{3}{B}"},
		{"Bane of the Living", lgn1BaneOracle, "{X}{B}{B}"},
		{"Bloodstoke Howler", lgn1BloodstokeOracle, "{6}{R}"},
		{"Branchsnap Lorian", lgn1BranchsnapOracle, "{G}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "morph")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no morph cast", tc.name)
			continue
		}
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownMorphed || ac.FaceDown.FaceUpCounter {
			t.Errorf("%s morph = %q %q counter=%v", tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCounter)
		}
		if ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s turns face up for %q, want %q", tc.name, ac.FaceDown.FaceUpCost, tc.faceUp)
		}
	}
}

func TestLgn1PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	akroma := b12Push(g, me.ID, "Akroma, Angel of Wrath", "Legendary Creature — Angel", lgn1AkromaOracle, 6, 6)
	envoy := b12Push(g, me.ID, "Aven Envoy", "Creature — Bird Soldier", lgn1AvenEnvoyOracle, 0, 2)
	lorian := b12Push(g, me.ID, "Branchsnap Lorian", "Creature — Beast", lgn1BranchsnapOracle, 4, 1)
	caller := b12Push(g, me.ID, "Caller of the Claw", "Creature — Elf", lgn1CallerOracle, 2, 2)
	for _, tc := range []struct {
		id  uuid.UUID
		kws []string
	}{
		{akroma, []string{"flying", "first strike", "vigilance", "trample", "haste", game.ProtectionFromColor("B"), game.ProtectionFromColor("R")}},
		{envoy, []string{"flying"}},
		{lorian, []string{"trample"}},
		{caller, []string{"flash"}},
	} {
		for _, kw := range tc.kws {
			if !hasEffectiveKeyword(t, g, tc.id, kw) {
				t.Errorf("%s is missing %q", tmtB1Card(t, g, tc.id).Name, kw)
			}
		}
	}
	if hasEffectiveKeyword(t, g, akroma, game.ProtectionFromColor("W")) {
		t.Error("Akroma has protection from white")
	}
}

// --- Akroma's Devoted, Blade Sliver -----------------------------------------

func TestLgn1AkromasDevotedGivesEveryClericVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	devoted := b12Push(g, me.ID, "Akroma's Devoted", "Creature — Human Cleric", lgn1DevotedOracle, 2, 4)
	theirs := b12Creature(g, opp.ID, "Their Cleric", "Creature — Human Cleric", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, devoted, "vigilance") || !hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("a Cleric creature has no vigilance")
	}
	if hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("a non-Cleric has vigilance")
	}
}

func TestLgn1BladeSliverPumpsEverySliver(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := b12Push(g, me.ID, "Blade Sliver", "Creature — Sliver", lgn1BladeSliverOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID][2]int{blade: {3, 2}, theirs: {2, 1}, bear: {2, 2}} {
		if p, tt := mkmMPPT(t, g, id); p != want[0] || tt != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", tmtB1Card(t, g, id).Name, p, tt, want[0], want[1])
		}
	}
}

// --- Aphetto Exterminator, Bane of the Living, Bloodstoke Howler -----------

// lgn1FaceDown casts a morph card face down for the active seat and empties
// that seat's pool, so the turn face up pays exactly its cost.
func lgn1FaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := s36CastFaceDown(t, g, name, typeLine, oracle, "morph", power, toughness)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	return me, id
}

func TestLgn1AphettoExterminatorShrinksATargetWhenTurnedUp(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	me, id := lgn1FaceDown(t, g, "Aphetto Exterminator", "Creature — Human Wizard", lgn1ExterminatorOracle, 3, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Exterminator")
	}
	mkmMPTurnFaceUp(t, g, me, id, "B", "C", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	if p, tt := mkmMPPT(t, g, ogre); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
}

func TestLgn1BaneOfTheLivingShrinksEveryCreatureByX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	me, id := lgn1FaceDown(t, g, "Bane of the Living", "Creature — Insect", lgn1BaneOracle, 4, 3)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("B", "B", "C", "C")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("turn face up for X=2: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0: X=2 costs {2}{B}{B}", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived -2/-2")
	}
	if p, tt := mkmMPPT(t, g, ogre); p != 2 || tt != 2 {
		t.Errorf("the 4/4 is %d/%d, want 2/2", p, tt)
	}
	if p, tt := mkmMPPT(t, g, id); p != 2 || tt != 1 {
		t.Errorf("the Bane is %d/%d, want 2/1: all creatures includes itself", p, tt)
	}
}

func TestLgn1BaneOfTheLivingAtXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	me, id := lgn1FaceDown(t, g, "Bane of the Living", "Creature — Insect", lgn1BaneOracle, 4, 3)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("B", "B")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up for X=0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := mkmMPPT(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("the 2/2 is %d/%d at X=0", p, tt)
	}
}

func TestLgn1BloodstokeHowlerPumpsYourBeasts(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Beast", "Creature — Beast", 2, 2)
	_, id := lgn1FaceDown(t, g, "Bloodstoke Howler", "Creature — Beast", lgn1BloodstokeOracle, 3, 4)
	mkmMPTurnFaceUp(t, g, me, id, "R", "C", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{id: 6, mine: 5, bear: 2, theirs: 2} {
		if p, _ := mkmMPPT(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", tmtB1Card(t, g, id).Name, p, want)
		}
	}
}

// --- Aven Redeemer -----------------------------------------------------------

func TestLgn1AvenRedeemerPreventsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	redeemer := b12Push(g, me.ID, "Aven Redeemer", "Creature — Bird Cleric", lgn1AvenRedeemerOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, redeemer, "flying") {
		t.Error("the Redeemer has no flying")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	ravB2MustActivate(t, g, me.ID, redeemer, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("3 damage into a 2-shield: marked %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("the shield is spent: marked %d, want 2", got)
	}
}

// --- Berserk Murlodont -------------------------------------------------------

func TestLgn1BerserkMurlodontPumpsAnyBeastPerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	b12Push(g, g.Seats[2].ID, "Berserk Murlodont", "Creature — Beast", lgn1MurlodontOracle, 3, 3)
	beast := b12Creature(g, g.Seats[0].ID, "Their Beast", "Creature — Beast", 2, 2)
	rampageBlock(t, g, beast, rampageBlockers(g, 2)...)
	if p, tt := mkmMPPT(t, g, beast); p != 4 || tt != 4 {
		t.Errorf("a Beast blocked by two is %d/%d, want 4/4", p, tt)
	}
}

func TestLgn1BerserkMurlodontIgnoresNonBeasts(t *testing.T) {
	g := newCatalogGame(t)
	b12Push(g, g.Seats[0].ID, "Berserk Murlodont", "Creature — Beast", lgn1MurlodontOracle, 3, 3)
	bear := b12Creature(g, g.Seats[0].ID, "My Bear", "Creature — Bear", 2, 2)
	rampageBlock(t, g, bear, rampageBlockers(g, 2)...)
	if p, tt := mkmMPPT(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("a non-Beast blocked by two is %d/%d, want 2/2", p, tt)
	}
}

// --- Blood Celebrant ---------------------------------------------------------

func TestLgn1BloodCelebrantFiltersBlackIntoAnyColorForALife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	celebrant := b12Push(g, me.ID, "Blood Celebrant", "Creature — Human Cleric", lgn1BloodCelebrantOracle, 1, 1)
	if err := g.ActivateManaAbility(me.ID, celebrant, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the Celebrant activated with no {B} to pay")
	}
	life := me.Life
	b06AddMana(me, "B")
	if err := g.ActivateManaAbility(me.ID, celebrant, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("G"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d green of %d, want 1 of 1", n, me.ManaPool.Total())
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 paid", life, me.Life)
	}
}

// --- Brood Sliver --------------------------------------------------------------

func TestLgn1BroodSliverOffersTheSliversControllerAToken(t *testing.T) {
	g := newCatalogGame(t)
	attacker, victim, brooder := g.Seats[0], g.Seats[1], g.Seats[2]
	brood := b12Push(g, brooder.ID, "Brood Sliver", "Creature — Sliver", lgn1BroodSliverOracle, 3, 3)
	sliver := b12Creature(g, attacker.ID, "Attacking Sliver", "Creature — Sliver", 2, 2)
	attackWith(t, g, victim.ID, sliver)
	g.RunStateChecksForTest()
	trig := triggerOnStack(g, brood)
	if trig == nil {
		t.Fatal("no Brood Sliver trigger on the stack")
	}
	if trig.Controller != brooder.ID {
		t.Errorf("trigger controller %s, want the Brood Sliver's controller", trig.Controller)
	}
	answerLatestTriggerPrompt(t, g, attacker.ID, true)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Sliver", attacker.ID); n != 1 {
		t.Errorf("the attacking Sliver's controller has %d Sliver tokens, want 1", n)
	}
	if n := tokensNamedControlledBy(g, "Sliver", brooder.ID); n != 0 {
		t.Errorf("the Brood Sliver's controller has %d Sliver tokens, want 0", n)
	}
}

func TestLgn1BroodSliverIgnoresNonSlivers(t *testing.T) {
	g := newCatalogGame(t)
	attacker, victim := g.Seats[0], g.Seats[1]
	brood := b12Push(g, attacker.ID, "Brood Sliver", "Creature — Sliver", lgn1BroodSliverOracle, 3, 3)
	bear := b12Creature(g, attacker.ID, "My Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, victim.ID, bear)
	g.RunStateChecksForTest()
	if triggerOnStack(g, brood) != nil {
		t.Error("a non-Sliver's combat damage triggered the Brood Sliver")
	}
}

// --- Caller of the Claw ------------------------------------------------------

func TestLgn1CallerOfTheClawCountsYourNontokenCreaturesThatDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, theirs} {
		b27Kill(g, id)
	}
	// A card that has since left the graveyard still went there.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(b) })
	if !me.Graveyard.Contains(a) || me.Graveyard.Contains(b) {
		t.Fatal("setup: one card in the graveyard and one exiled from it")
	}
	castCatalogSpell(t, g, "Caller of the Claw", "Creature — Elf", lgn1CallerOracle, nil)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Bear", me.ID); n != 2 {
		t.Errorf("%d Bear tokens, want 2 for my two creatures", n)
	}
	if n := tokensNamedControlledBy(g, "Bear", opp.ID); n != 0 {
		t.Errorf("the opponent got %d Bears", n)
	}
}
