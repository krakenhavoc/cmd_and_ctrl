package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Unidentified Hovership ----------------------------------------------------------

func TestDskUZUnidentifiedHovershipExilesAndTheOwnerManifestsDreadWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 5)
	_, oppSecond := effectstest.DreadPair(opp)
	hovership := effectstest.CastCatalogSpell(t, g, "Unidentified Hovership", "Artifact — Vehicle", dskUZHovershipOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.RooncRef(giant)); err == nil {
		t.Fatal("a creature with toughness 6 was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if !effectstest.HasEffectiveKeyword(t, g, hovership, "flying") {
		t.Error("the Hovership lacks flying")
	}
	effectstest.RooncDestroy(t, g, hovership)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the Hovership's controller manifested; the exiled card's owner does")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, oppSecond)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, opp, oppSecond)
	if !g.Exile.Contains(bear) {
		t.Error("the exiled card came back; nothing returns")
	}
}

func TestDskUZUnidentifiedHovershipWithNothingExiledManifestsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DreadPair(me)
	hovership := effectstest.CastCatalogSpell(t, g, "Unidentified Hovership", "Artifact — Vehicle", dskUZHovershipOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("pick no target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RooncDestroy(t, g, hovership)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if effectstest.ChooseCardsChoiceFor(g, s.ID) != nil {
			t.Errorf("seat %s was asked to manifest dread with nothing exiled", s.Name)
		}
	}
}

// --- Unwanted Remake -------------------------------------------------------------------

func TestDskUZUnwantedRemakeDestroysAndItsControllerManifestsDread(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oppTop, oppSecond := effectstest.DreadPair(opp)
	effectstest.CastCatalogSpell(t, g, "Unwanted Remake", "Instant", dskUZRemakeOracle, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not destroyed")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster manifested; the creature's controller does")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, oppSecond)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, opp, oppSecond)
	if !effectstest.MdInGraveyard(opp, oppTop) {
		t.Error("the other looked-at card is not in its owner's graveyard")
	}
}

// --- Valgavoth's Faithful ----------------------------------------------------------------

func TestDskUZValgavothsFaithfulSacrificesItselfToReanimateAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	faithful := effectstest.B12Push(g, me.ID, "Valgavoth's Faithful", "Creature — Human Cleric", dskUZFaithfulOracle, 1, 1)
	bear := effectstest.PushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 4, 4)
	if err := effectstest.DskRT2Activate(g, me.ID, faithful, 0, "{B}{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.DskRT2Activate(g, me.ID, faithful, 0, "{B}{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(faithful) || !me.Graveyard.Contains(faithful) {
		t.Error("the Faithful was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card did not return to the battlefield")
	}
}

// --- Vanish from Sight ---------------------------------------------------------------------

func TestDskUZVanishFromSightTucksForTheOwnerThenSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	rock := effectstest.PushArtifactForTest(g, opp.ID, "Their Mind Stone")
	if err := effectstest.CastCatalogSpellErr(t, g, "Vanish from Sight", "Instant", dskUZVanishOracle, effectstest.B16TargetCard(land)); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Vanish from Sight", "Instant", dskUZVanishOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("the surveil came before the owner answered")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{rock}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(rock) || opp.Library.Cards[0].InstanceID != rock {
		t.Fatal("the artifact is not on the bottom of its owner's library")
	}
	effectstest.EoeQZKeepOnTop(t, g, me.ID, effectstest.EoeQZSurveil(t, g, me.ID, 1))
}

// --- Vengeful Possession --------------------------------------------------------------------

func TestDskUZVengefulPossessionThreatensThenLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Vengeful Possession", "Sorcery", dskUZPossessionOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatal("control of the creature did not change")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the creature did not gain haste")
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("discard prompt for %d, want a may-discard of one", effectstest.DiscardOwed(g, me.ID))
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) || me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the discard replaced by a draw", hand, me.Hand.Size())
	}
	effectstest.EndTurn(t, g)
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestDskUZVengefulPossessionDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Keep", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Vengeful Possession", "Sorcery", dskUZPossessionOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw without a discard", hand, me.Hand.Size())
	}
}

// --- Vicious Clown ----------------------------------------------------------------------------

func TestDskUZViciousClownGrowsWhenASmallCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clown := effectstest.B12Push(g, me.ID, "Vicious Clown", "Creature — Human Clown", dskUZClownOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	if p, _ := effectstest.PtOf(t, g, clown); p != 2 {
		t.Errorf("a power-3 creature entered: power %d, want 2", p)
	}
	effectstest.DftABCast(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, clown); p != 4 || tt != 3 {
		t.Errorf("a power-2 creature entered: %d/%d, want 4/3", p, tt)
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, clown); p != 2 {
		t.Errorf("the boost outlasted the turn: power %d", p)
	}
}

// --- Vile Mutilator ------------------------------------------------------------------------------

func TestDskUZVileMutilatorCostsASacrificeAndEdictsAnEnchantmentThenACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	myLand := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	if _, err := effectstest.TryCastWithSacrifice(g, "Vile Mutilator", "Creature — Demon", dskUZMutilatorOracle, []uuid.UUID{myLand}); err == nil {
		t.Fatal("a land was accepted as the additional cost")
	}
	myAura := effectstest.B12Permanent(g, me.ID, "My Shrine", "Enchantment")
	myOther := effectstest.B12Permanent(g, me.ID, "My Other Shrine", "Enchantment")
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirShrine := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The third seat has no enchantment: it is not asked for one, and
	// still loses its creature.
	thirdBear := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.CastWithSacrifice(t, g, "Vile Mutilator", "Creature — Demon", dskUZMutilatorOracle, myAura)
	if g.Battlefield.Contains(myAura) {
		t.Error("the sacrifice is paid while casting")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
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
	mutilator := effectstest.AvB1Named(g, me.ID, "Vile Mutilator")
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, mutilator, kw) {
			t.Errorf("the Mutilator lacks %s", kw)
		}
	}
}

// --- Violent Urge ----------------------------------------------------------------------------------

func TestDskUZViolentUrgeGrantsDoubleStrikeOnlyWithDelirium(t *testing.T) {
	for _, types := range []int{3, 4} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.DskRT2Delirium(g, me, types)
		effectstest.CastCatalogSpell(t, g, "Violent Urge", "Instant", dskUZUrgeOracle, effectstest.B16TargetCard(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 2 {
			t.Errorf("%d types: %d/%d, want 3/2", types, p, tt)
		}
		if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
			t.Errorf("%d types: no first strike", types)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "double strike"); got != (types == 4) {
			t.Errorf("%d types: double strike %v", types, got)
		}
	}
}

// --- Wary Watchdog ------------------------------------------------------------------------------------

func TestDskUZWaryWatchdogSurveilsAsItEntersAndAsItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Wary Watchdog", "Creature — Dog", dskUZWatchdogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeQZKeepOnTop(t, g, me.ID, effectstest.EoeQZSurveil(t, g, me.ID, 1))
	dog := effectstest.AvB1Named(g, me.ID, "Wary Watchdog")
	if dog == uuid.Nil {
		t.Fatal("Wary Watchdog is not on the battlefield")
	}
	effectstest.B15Destroy(t, g, dog)
	effectstest.EoeQZSurveil(t, g, me.ID, 1)
}

// --- Wickerfolk Thresher ------------------------------------------------------------------------------

func TestDskUZWickerfolkThresherPutsALandOntoTheBattlefieldWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := effectstest.B12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.DskRT2Delirium(g, me, 4)
	effectstest.DeclareAttack(t, g, opp.ID, thresher)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the Thresher did not offer the land")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("the land did not enter untapped")
	}
}

func TestDskUZWickerfolkThresherHandsANonlandWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := effectstest.B12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	bolt := effectstest.PlTop(me, "Lightning Bolt", "Instant", "{R}")
	effectstest.DskRT2Delirium(g, me, 4)
	effectstest.DeclareAttack(t, g, opp.ID, thresher)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a nonland card has nothing to offer")
	}
	if !me.Hand.Contains(bolt) {
		t.Error("the nonland did not go to hand")
	}
}

func TestDskUZWickerfolkThresherDoesNothingWithoutDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thresher := effectstest.B12Push(g, me.ID, "Wickerfolk Thresher", "Artifact Creature — Scarecrow", dskUZThresherOracle, 5, 4)
	bolt := effectstest.PlTop(me, "Lightning Bolt", "Instant", "{R}")
	effectstest.DskRT2Delirium(g, me, 3)
	effectstest.DeclareAttack(t, g, opp.ID, thresher)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bolt) || !me.Library.Contains(bolt) {
		t.Error("three card types: the Thresher looked anyway")
	}
}

// --- Wildfire Wickerfolk ----------------------------------------------------------------------------

func TestDskUZWildfireWickerfolkGrowsAndTramplesWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wicker := effectstest.B12Push(g, me.ID, "Wildfire Wickerfolk", "Artifact Creature — Scarecrow", dskUZWildfireOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, wicker, "haste") {
		t.Error("no haste")
	}
	effectstest.DskRT2Delirium(g, me, 3)
	if p, tt := effectstest.PtOf(t, g, wicker); p != 3 || tt != 2 {
		t.Errorf("three card types: %d/%d, want 3/2", p, tt)
	}
	if effectstest.HasEffectiveKeyword(t, g, wicker, "trample") {
		t.Error("three card types: trample")
	}
	effectstest.DskRT2Delirium(g, me, 4)
	if p, tt := effectstest.PtOf(t, g, wicker); p != 4 || tt != 3 {
		t.Errorf("delirium: %d/%d, want 4/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wicker, "trample") {
		t.Error("delirium: no trample")
	}
}

// --- Winter's Intervention ---------------------------------------------------------------------------

func TestDskUZWintersInterventionDealsTwoAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Winter's Intervention", "Instant", dskUZWinterOracle, effectstest.B16TargetCard(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, giant); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}
