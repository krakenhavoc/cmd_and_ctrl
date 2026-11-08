package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_uz_cards_test.go — Duskmourn: House of Horror, the U–Z batch
// (#4466). Each test plays the real catalog entry and asserts what its
// text says.

const (
	dskUZHovershipOracle  = "bde67dfb-706b-4942-a365-984ea6e97a2b"
	dskUZRemakeOracle     = "90aa9113-5148-486b-a1c6-b4230e6d3e41"
	dskUZFaithfulOracle   = "a871dfbf-1fb9-4d1c-81bb-e4f0b131a11b"
	dskUZVanishOracle     = "1056b27c-125c-4daa-a0c8-533efc9ccc32"
	dskUZPossessionOracle = "1afb9ce6-658e-40e4-a4a2-9a68182a1e73"
	dskUZClownOracle      = "e02abbc0-687b-4aef-95b8-bf18e199725a"
	dskUZMutilatorOracle  = "1aae80a5-079c-48f3-aad1-563931e1a031"
	dskUZUrgeOracle       = "0ef80a38-f464-4d1f-9e81-d4087d6ecc9b"
	dskUZWatchdogOracle   = "7e063f60-876a-48fe-8bc6-9888180176d1"
	dskUZThresherOracle   = "fa6d7d68-34e3-4ff6-ae6c-36848974555a"
	dskUZWildfireOracle   = "50c7b13d-f217-45de-837c-49f934840368"
	dskUZWinterOracle     = "91967284-6711-4ef3-bb9b-5ba873da8bb2"
)

func TestDskUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskUZHovershipOracle:  "Unidentified Hovership",
		dskUZRemakeOracle:     "Unwanted Remake",
		dskUZFaithfulOracle:   "Valgavoth's Faithful",
		dskUZVanishOracle:     "Vanish from Sight",
		dskUZPossessionOracle: "Vengeful Possession",
		dskUZClownOracle:      "Vicious Clown",
		dskUZMutilatorOracle:  "Vile Mutilator",
		dskUZUrgeOracle:       "Violent Urge",
		dskUZWatchdogOracle:   "Wary Watchdog",
		dskUZThresherOracle:   "Wickerfolk Thresher",
		dskUZWildfireOracle:   "Wildfire Wickerfolk",
		dskUZWinterOracle:     "Winter's Intervention",
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

// --- Unidentified Hovership ----------------------------------------------------------

func TestDskUZUnidentifiedHovershipExilesAndTheOwnerManifestsDreadWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 5)
	_, oppSecond := dreadPair(opp)
	hovership := castCatalogSpell(t, g, "Unidentified Hovership", "Artifact — Vehicle", dskUZHovershipOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, rooncRef(giant)); err == nil {
		t.Fatal("a creature with toughness 6 was accepted as the target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if !hasEffectiveKeyword(t, g, hovership, "flying") {
		t.Error("the Hovership lacks flying")
	}
	rooncDestroy(t, g, hovership)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the Hovership's controller manifested; the exiled card's owner does")
	}
	answerChooseCards(t, g, opp.ID, oppSecond)
	passPriorityAroundTable(t, g)
	requireManifested(t, g, opp, oppSecond)
	if !g.Exile.Contains(bear) {
		t.Error("the exiled card came back; nothing returns")
	}
}

func TestDskUZUnidentifiedHovershipWithNothingExiledManifestsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dreadPair(me)
	hovership := castCatalogSpell(t, g, "Unidentified Hovership", "Artifact — Vehicle", dskUZHovershipOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("pick no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	rooncDestroy(t, g, hovership)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if chooseCardsChoiceFor(g, s.ID) != nil {
			t.Errorf("seat %s was asked to manifest dread with nothing exiled", s.Name)
		}
	}
}

// --- Unwanted Remake -------------------------------------------------------------------

func TestDskUZUnwantedRemakeDestroysAndItsControllerManifestsDread(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oppTop, oppSecond := dreadPair(opp)
	castCatalogSpell(t, g, "Unwanted Remake", "Instant", dskUZRemakeOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not destroyed")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster manifested; the creature's controller does")
	}
	answerChooseCards(t, g, opp.ID, oppSecond)
	passPriorityAroundTable(t, g)
	requireManifested(t, g, opp, oppSecond)
	if !mdInGraveyard(opp, oppTop) {
		t.Error("the other looked-at card is not in its owner's graveyard")
	}
}

// --- Valgavoth's Faithful ----------------------------------------------------------------

func TestDskUZValgavothsFaithfulSacrificesItselfToReanimateAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	faithful := b12Push(g, me.ID, "Valgavoth's Faithful", "Creature — Human Cleric", dskUZFaithfulOracle, 1, 1)
	bear := pushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 4, 4)
	if err := dskRT2Activate(g, me.ID, faithful, 0, "{B}{C}{C}{C}", game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	if err := dskRT2Activate(g, me.ID, faithful, 0, "{B}{C}{C}{C}", game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(faithful) || !me.Graveyard.Contains(faithful) {
		t.Error("the Faithful was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card did not return to the battlefield")
	}
}

// --- Vanish from Sight ---------------------------------------------------------------------

func TestDskUZVanishFromSightTucksForTheOwnerThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	rock := pushArtifactForTest(g, opp.ID, "Their Mind Stone")
	if err := castCatalogSpellErr(t, g, "Vanish from Sight", "Instant", dskUZVanishOracle, b16TargetCard(land)); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	castCatalogSpell(t, g, "Vanish from Sight", "Instant", dskUZVanishOracle, b16TargetCard(rock))
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
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{rock}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(rock) || opp.Library.Cards[0].InstanceID != rock {
		t.Fatal("the artifact is not on the bottom of its owner's library")
	}
	eoeQZKeepOnTop(t, g, me.ID, eoeQZSurveil(t, g, me.ID, 1))
}

// --- Vengeful Possession --------------------------------------------------------------------

func TestDskUZVengefulPossessionThreatensThenLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	findBattlefieldCardForTest(g, theirs).Tapped = true
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Vengeful Possession", "Sorcery", dskUZPossessionOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatal("control of the creature did not change")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the creature did not gain haste")
	}
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard prompt for %d, want a may-discard of one", discardOwed(g, me.ID))
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) || me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the discard replaced by a draw", hand, me.Hand.Size())
	}
	endTurn(t, g)
	if c, _ := battlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestDskUZVengefulPossessionDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Keep", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Vengeful Possession", "Sorcery", dskUZPossessionOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw without a discard", hand, me.Hand.Size())
	}
}

// --- Vicious Clown ----------------------------------------------------------------------------

func TestDskUZViciousClownGrowsWhenASmallCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clown := b12Push(g, me.ID, "Vicious Clown", "Creature — Human Clown", dskUZClownOracle, 2, 3)
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	if p, _ := ptOf(t, g, clown); p != 2 {
		t.Errorf("a power-3 creature entered: power %d, want 2", p)
	}
	dftABCast(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, clown); p != 4 || tt != 3 {
		t.Errorf("a power-2 creature entered: %d/%d, want 4/3", p, tt)
	}
	endTurn(t, g)
	if p, _ := ptOf(t, g, clown); p != 2 {
		t.Errorf("the boost outlasted the turn: power %d", p)
	}
}

// --- Vile Mutilator ------------------------------------------------------------------------------

func TestDskUZVileMutilatorCostsASacrificeAndEdictsAnEnchantmentThenACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	myLand := b12Permanent(g, me.ID, "My Land", "Land")
	if _, err := tryCastWithSacrifice(g, "Vile Mutilator", "Creature — Demon", dskUZMutilatorOracle, []uuid.UUID{myLand}); err == nil {
		t.Fatal("a land was accepted as the additional cost")
	}
	myAura := b12Permanent(g, me.ID, "My Shrine", "Enchantment")
	myOther := b12Permanent(g, me.ID, "My Other Shrine", "Enchantment")
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirShrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The third seat has no enchantment: it is not asked for one, and
	// still loses its creature.
	thirdBear := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	castWithSacrifice(t, g, "Vile Mutilator", "Creature — Demon", dskUZMutilatorOracle, myAura)
	if g.Battlefield.Contains(myAura) {
		t.Error("the sacrifice is paid while casting")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the controller is asked to sacrifice")
	}
	if g.Battlefield.Contains(theirShrine) {
		t.Error("the opponent's nontoken enchantment stayed")
	}
	for _, id := range []uuid.UUID{theirBear, thirdBear} {
		if g.Battlefield.Contains(id) {
			t.Error("an opponent's nontoken creature stayed")
		}
	}
	if !g.Battlefield.Contains(myOther) || !g.Battlefield.Contains(myBear) {
		t.Error("the controller's own permanents were sacrificed")
	}
	mutilator := avB1Named(g, me.ID, "Vile Mutilator")
	for _, kw := range []string{"flying", "trample"} {
		if !hasEffectiveKeyword(t, g, mutilator, kw) {
			t.Errorf("the Mutilator lacks %s", kw)
		}
	}
}

// --- Violent Urge ----------------------------------------------------------------------------------

func TestDskUZViolentUrgeGrantsDoubleStrikeOnlyWithDelirium(t *testing.T) {
	for _, types := range []int{3, 4} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		dskRT2Delirium(g, me, types)
		castCatalogSpell(t, g, "Violent Urge", "Instant", dskUZUrgeOracle, b16TargetCard(bear))
		passPriorityAroundTable(t, g)
		if p, tt := ptOf(t, g, bear); p != 3 || tt != 2 {
			t.Errorf("%d types: %d/%d, want 3/2", types, p, tt)
		}
		if !hasEffectiveKeyword(t, g, bear, "first strike") {
			t.Errorf("%d types: no first strike", types)
		}
		if got := hasEffectiveKeyword(t, g, bear, "double strike"); got != (types == 4) {
			t.Errorf("%d types: double strike %v", types, got)
		}
	}
}

// --- Wary Watchdog ------------------------------------------------------------------------------------

func TestDskUZWaryWatchdogSurveilsAsItEntersAndAsItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Wary Watchdog", "Creature — Dog", dskUZWatchdogOracle, nil)
	passPriorityAroundTable(t, g)
	eoeQZKeepOnTop(t, g, me.ID, eoeQZSurveil(t, g, me.ID, 1))
	dog := avB1Named(g, me.ID, "Wary Watchdog")
	if dog == uuid.Nil {
		t.Fatal("Wary Watchdog is not on the battlefield")
	}
	b15Destroy(t, g, dog)
	eoeQZSurveil(t, g, me.ID, 1)
}

// --- Wickerfolk Thresher ------------------------------------------------------------------------------

func TestDskUZWickerfolkThresherPutsALandOntoTheBattlefieldWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := b12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	dskRT2Delirium(g, me, 4)
	declareAttack(t, g, opp.ID, thresher)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the Thresher did not offer the land")
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || tappedOnBattlefield(t, g, land) {
		t.Error("the land did not enter untapped")
	}
}

func TestDskUZWickerfolkThresherHandsANonlandWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := b12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	dskRT2Delirium(g, me, 4)
	declareAttack(t, g, opp.ID, thresher)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a nonland card has nothing to offer")
	}
	if !me.Hand.Contains(bolt) {
		t.Error("the nonland did not go to hand")
	}
}

func TestDskUZWickerfolkThresherDoesNothingWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := b12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	dskRT2Delirium(g, me, 3)
	declareAttack(t, g, opp.ID, thresher)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(bolt) || !me.Library.Contains(bolt) {
		t.Error("three card types: the Thresher looked anyway")
	}
}

// --- Wildfire Wickerfolk ----------------------------------------------------------------------------

func TestDskUZWildfireWickerfolkGrowsAndTramplesWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wicker := b12Push(g, me.ID, "Wildfire Wickerfolk", "Artifact Creature — Scarecrow", dskUZWildfireOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, wicker, "haste") {
		t.Error("no haste")
	}
	dskRT2Delirium(g, me, 3)
	if p, tt := ptOf(t, g, wicker); p != 3 || tt != 2 {
		t.Errorf("three card types: %d/%d, want 3/2", p, tt)
	}
	if hasEffectiveKeyword(t, g, wicker, "trample") {
		t.Error("three card types: trample")
	}
	dskRT2Delirium(g, me, 4)
	if p, tt := ptOf(t, g, wicker); p != 4 || tt != 3 {
		t.Errorf("delirium: %d/%d, want 4/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, wicker, "trample") {
		t.Error("delirium: no trample")
	}
}

// --- Winter's Intervention ---------------------------------------------------------------------------

func TestDskUZWintersInterventionDealsTwoAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	life := me.Life
	castCatalogSpell(t, g, "Winter's Intervention", "Instant", dskUZWinterOracle, b16TargetCard(giant))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}
