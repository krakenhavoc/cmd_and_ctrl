package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_s3_cards_test.go — Tarkir: Dragonstorm, the third batch of cards
// from S (#4463). Each test plays the real catalog entry and asserts what
// its text says.

const (
	tdmS3SageOfTheSkiesOracle    = "cfc7cef7-0a95-49f8-970f-abff2cacd82d"
	tdmS3SaltRoadSkirmishOracle  = "f6e72f4e-d6e0-4537-93c1-7045385c8401"
	tdmS3SeizeOpportunityOracle  = "d0458ffe-9a61-443f-8e80-3d2dde224a4d"
	tdmS3ShikoOracle             = "d3bdd3f2-9e38-436a-93a3-f87d9f4d4f86"
	tdmS3SnowmeltStagOracle      = "b25ccb64-e114-4d54-a07a-c8cd4560b511"
	tdmS3SonicShriekerOracle     = "d317e3f1-553f-44f6-8543-d6d4c42534ad"
	tdmS3SpectralDenialOracle    = "f8536726-993b-4100-918c-e3c7177d4c65"
	tdmS3StaticSnareOracle       = "b5b742d2-f19b-47c9-84bd-e85d975900e6"
	tdmS3StormscaleScionOracle   = "cf37426d-b0ad-42c3-81be-79701aa4e0cf"
	tdmS3StrategicBetrayalOracle = "0127c13f-7b0f-44c3-8b35-97f306c5f239"
	tdmS3SunpearlKirinOracle     = "f17b9356-a329-4f89-b690-afc2eaff4977"
	tdmS3SurrakOracle            = "7a0fe73f-6e08-4c9a-80e1-c2ead6cebe06"
)

func TestTdmS3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmS3SageOfTheSkiesOracle:    "Sage of the Skies",
		tdmS3SaltRoadSkirmishOracle:  "Salt Road Skirmish",
		tdmS3SeizeOpportunityOracle:  "Seize Opportunity",
		tdmS3ShikoOracle:             "Shiko, Paragon of the Way",
		tdmS3SnowmeltStagOracle:      "Snowmelt Stag",
		tdmS3SonicShriekerOracle:     "Sonic Shrieker",
		tdmS3SpectralDenialOracle:    "Spectral Denial",
		tdmS3StaticSnareOracle:       "Static Snare",
		tdmS3StormscaleScionOracle:   "Stormscale Scion",
		tdmS3StrategicBetrayalOracle: "Strategic Betrayal",
		tdmS3SunpearlKirinOracle:     "Sunpearl Kirin",
		tdmS3SurrakOracle:            "Surrak, Elusive Hunter",
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

// --- fixtures ------------------------------------------------------------------

// tdmS3Opp is the seat after the active one.
func tdmS3Opp(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// tdmS3Library puts n plain cards on top of p's library and returns them,
// the last one on top.
func tdmS3Library(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, topLibraryCard(p, "Library Card", "Sorcery"))
	}
	return ids
}

// --- Sage of the Skies ---------------------------------------------------------

func TestTdmS3SageOfTheSkiesCopiesItselfAfterAnotherSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "First Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	tdmSTCast(t, g, "Sage of the Skies", "Creature — Human Monk", tdmS3SageOfTheSkiesOracle, 2, 3)
	passPriorityAroundTable(t, g)
	sages := socB8Named(g, me.ID, "Sage of the Skies")
	if len(sages) != 2 {
		t.Fatalf("%d Sages on the battlefield, want the spell and its copy", len(sages))
	}
	tokens := 0
	for _, c := range sages {
		if c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d of the Sages are tokens, want the copy only", tokens)
	}
	if !hasEffectiveKeyword(t, g, sages[0].InstanceID, "lifelink") || !hasEffectiveKeyword(t, g, sages[0].InstanceID, "flying") {
		t.Error("the Sage has no flying and lifelink")
	}
}

func TestTdmS3SageOfTheSkiesAsTheFirstSpellIsNotCopied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tdmSTCast(t, g, "Sage of the Skies", "Creature — Human Monk", tdmS3SageOfTheSkiesOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Sage of the Skies")); n != 1 {
		t.Errorf("%d Sages on the battlefield, want 1 (no other spell was cast)", n)
	}
}

// --- Salt Road Skirmish --------------------------------------------------------

func TestTdmS3SaltRoadSkirmishDestroysAndMakesTwoHastyWarriorsForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	bear := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Salt Road Skirmish", "Sorcery", tdmS3SaltRoadSkirmishOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the target creature was not destroyed")
	}
	warriors := socB8Named(g, me.ID, "Warrior")
	if len(warriors) != 2 {
		t.Fatalf("%d Warriors, want 2", len(warriors))
	}
	for _, w := range warriors {
		if w.Power != 1 || w.Toughness != 1 || !w.IsToken() {
			t.Errorf("Warrior %d/%d token=%v, want a 1/1 token", w.Power, w.Toughness, w.IsToken())
		}
		if !hasEffectiveKeyword(t, g, w.InstanceID, "haste") {
			t.Error("a Warrior has no haste this turn")
		}
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 0 {
		t.Errorf("%d Warriors survived the end step", n)
	}
}

// --- Seize Opportunity ---------------------------------------------------------

func TestTdmS3SeizeOpportunityExilesTheTopTwoToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := tdmS3Library(me, 3)
	castModal(t, g, "Seize Opportunity", "Instant", tdmS3SeizeOpportunityOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	for _, id := range top[1:] {
		if !g.Exile.Contains(id) {
			t.Fatal("the top two cards were not exiled")
		}
		var perm *game.CastPermission
		g.ReadSnapshot(func() {
			for _, c := range g.Exile.Cards {
				if c.InstanceID == id {
					perm = g.CastPermissionOnCardForEffect(c, game.ZoneExile)
				}
			}
		})
		if perm == nil || perm.Player != me.ID {
			t.Errorf("no play permission for the exiled card: %+v", perm)
		}
	}
	if g.Exile.Contains(top[0]) {
		t.Error("a third card was exiled")
	}
}

func TestTdmS3SeizeOpportunityPumpsUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b, c := seedBear(g, me.ID), seedBear(g, me.ID), seedBear(g, me.ID)
	castModal(t, g, "Seize Opportunity", "Instant", tdmS3SeizeOpportunityOracle, []int{1}, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 4 || tough != 3 {
			t.Errorf("a target is %d/%d, want 4/3", p, tough)
		}
	}
	if p := effectivePower(t, g, c); p != 2 {
		t.Errorf("the creature not targeted has power %d, want 2", p)
	}
}

// --- Shiko, Paragon of the Way -------------------------------------------------

func TestTdmS3ShikoCastsAFreeCopyOfACheapCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	big := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	tdmSTCast(t, g, "Shiko, Paragon of the Way", "Legendary Creature — Spirit Dragon", tdmS3ShikoOracle, 4, 5)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(big)); err == nil {
		t.Fatal("Shiko targeted a card with mana value 4")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !inExile(g, bolt) {
		t.Fatal("the targeted card was not exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	life := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want the copied Bolt's 3", life, opp.Life)
	}
	if !inExile(g, bolt) {
		t.Error("the original card left exile")
	}
}

// --- Snowmelt Stag -------------------------------------------------------------

func TestTdmS3SnowmeltStagIsFiveTwoOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stag := b12Push(g, me.ID, "Snowmelt Stag", "Creature — Elemental Elk", tdmS3SnowmeltStagOracle, 2, 5)
	if p, tough := effectivePower(t, g, stag), effectiveToughness(t, g, stag); p != 5 || tough != 2 {
		t.Errorf("on your turn the Stag is %d/%d, want 5/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, stag, "vigilance") {
		t.Error("the Stag has no vigilance")
	}
	advanceToEndStepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tough := effectivePower(t, g, stag), effectiveToughness(t, g, stag); p != 2 || tough != 5 {
		t.Errorf("on an opponent's turn the Stag is %d/%d, want 2/5", p, tough)
	}
}

func TestTdmS3SnowmeltStagCantBeBlockedForSevenMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stag := b12Push(g, me.ID, "Snowmelt Stag", "Creature — Elemental Elk", tdmS3SnowmeltStagOracle, 2, 5)
	advanceToMain(t, g)
	fillPoolColored(me, "U", 7)
	if err := g.ActivateCatalogAbility(me.ID, stag, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, stag, game.CantBeBlocked)
}

// --- Sonic Shrieker ------------------------------------------------------------

func TestTdmS3SonicShriekerHitsAPlayerWhoDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	handCard(opp, "Their Card", "Sorcery")
	hand, myLife, oppLife := opp.Hand.Size(), me.Life, opp.Life
	tdmSTCast(t, g, "Sonic Shrieker", "Creature — Dragon", tdmS3SonicShriekerOracle, 4, 4)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d -> %d, want -2", oppLife, opp.Life)
	}
	if me.Life != myLife+2 {
		t.Errorf("your life %d -> %d, want +2", myLife, me.Life)
	}
	if opp.Hand.Size() != hand-1 {
		t.Errorf("opponent hand %d -> %d, want one card discarded", hand, opp.Hand.Size())
	}
}

func TestTdmS3SonicShriekerHittingACreatureMakesNobodyDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	handCard(opp, "Their Card", "Sorcery")
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	myLife := me.Life
	tdmSTCast(t, g, "Sonic Shrieker", "Creature — Dragon", tdmS3SonicShriekerOracle, 4, 4)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil {
		t.Error("damage to a creature asked its controller to discard")
	}
	if me.Life != myLife+2 {
		t.Errorf("your life %d -> %d, want +2", myLife, me.Life)
	}
}

// --- Spectral Denial -----------------------------------------------------------

// tdmS3CastDenial has the opponent cast Spectral Denial at X=2 for {U}
// alone, against a spell the active player just cast.
func tdmS3CastDenial(t *testing.T, g *game.Game, victim uuid.UUID) error {
	t.Helper()
	opp := tdmS3Opp(g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	id := handCardFull(opp, "Spectral Denial", "Instant", "{X}{U}", tdmS3SpectralDenialOracle, []string{"U"})
	tmcB6Pool(t, g, opp, "{U}")
	return g.CastSpell(opp.ID, id, game.CastSpellParams{XValue: 2, Targets: cardRefs(victim), Strict: true})
}

func TestTdmS3SpectralDenialCostsLessPerBigCreatureButAsksTheFullX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	b12Creature(g, opp.ID, "Big One", "Creature — Beast", 4, 4)
	b12Creature(g, opp.ID, "Big Two", "Creature — Beast", 5, 5)
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)
	if err := tdmS3CastDenial(t, g, victim); err != nil {
		t.Fatalf("Spectral Denial at X=2 with two power-4 creatures did not cost {U}: %v", err)
	}
	passUntilTaxed(t, g, me.ID)
	c := payUnlessFor(g, me.ID)
	if c.PayCost != "{2}" {
		t.Fatalf("the counter asks for %q, want {2} (the reduction does not lower X)", c.PayCost)
	}
	answerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) || !me.Graveyard.Contains(victim) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
}

func TestTdmS3SpectralDenialCountsOnlyPowerFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	opp := tdmS3Opp(g)
	b12Creature(g, opp.ID, "Big One", "Creature — Beast", 4, 4)
	b12Creature(g, opp.ID, "Small", "Creature — Beast", 3, 3)
	seedBear(g, g.Seats[g.Turn.ActiveSeat].ID)
	b12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "Not Theirs", "Creature — Beast", 6, 6)
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)
	if err := tdmS3CastDenial(t, g, victim); err == nil {
		t.Error("Spectral Denial at X=2 was cast for {U} with one power-4 creature of its caster's")
	}
}

// --- Static Snare --------------------------------------------------------------

func TestTdmS3StaticSnareCostsLessPerAttackerAndExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	a := b12Creature(g, me.ID, "Attacker One", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Attacker Two", "Creature — Bear", 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	declareAttack(t, g, opp.ID, a, b)
	snare := game.Card{Name: "Static Snare", TypeLine: "Enchantment", ManaCost: "{4}{W}", OracleID: tdmS3StaticSnareOracle}
	tmcB6Pool(t, g, me, "{W}{C}")
	if _, err := tmcB6Cast(g, me, snare, game.CastSpellParams{}); err == nil {
		t.Fatal("Static Snare was cast for {1}{W} with two attackers")
	}
	tmcB6Pool(t, g, me, "{C}")
	id, err := tmcB6Cast(g, me, snare, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast for {2}{W} with two attackers: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(relic) {
		t.Fatal("the opponent's artifact was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Relic", opp.ID); n != 1 {
		t.Errorf("after the Snare left, %d Relics under their owner's control, want 1", n)
	}
}

// --- Stormscale Scion ----------------------------------------------------------

func TestTdmS3StormscaleScionStormsAndPumpsOtherDragons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := b12Creature(g, me.ID, "Old Dragon", "Creature — Dragon", 2, 2)
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "First Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	tdmSTCast(t, g, "Stormscale Scion", "Creature — Dragon", tdmS3StormscaleScionOracle, 4, 4)
	passPriorityAroundTable(t, g)
	scions := socB8Named(g, me.ID, "Stormscale Scion")
	if len(scions) != 2 {
		t.Fatalf("%d Scions, want the spell and one storm copy", len(scions))
	}
	// Two Scions: the other Dragon gets +2/+2, each Scion +1/+1 from the
	// other, and the Bear nothing.
	if p := effectivePower(t, g, dragon); p != 4 {
		t.Errorf("the other Dragon has power %d, want 4", p)
	}
	for _, s := range scions {
		if p := effectivePower(t, g, s.InstanceID); p != 5 {
			t.Errorf("a Scion has power %d, want 5", p)
		}
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
}

// --- Strategic Betrayal --------------------------------------------------------

func TestTdmS3StrategicBetrayalExilesTheirChoiceAndTheirGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	chosen := seedBear(g, opp.ID)
	kept := seedBear(g, opp.ID)
	gone := gyCard(opp, "Their Dead", "Creature — Bear", "{1}{G}", "")
	mine := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Strategic Betrayal", "Sorcery", tdmS3StrategicBetrayalOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	etChooseCards(t, g, opp.ID, 2, chosen)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(chosen) || !g.Exile.Contains(gone) {
		t.Error("the chosen creature and the graveyard were not exiled")
	}
	if !g.Battlefield.Contains(kept) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not chosen left the battlefield")
	}
}

func TestTdmS3StrategicBetrayalWithNoCreatureStillExilesTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := tdmS3Opp(g)
	gone := gyCard(opp, "Their Dead", "Sorcery", "{1}{B}", "")
	castCatalogSpell(t, g, "Strategic Betrayal", "Sorcery", tdmS3StrategicBetrayalOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, opp.ID) != nil {
		t.Error("an opponent with no creature was asked to choose one")
	}
	if !g.Exile.Contains(gone) || opp.Graveyard.Size() != 0 {
		t.Error("the opponent's graveyard was not exiled")
	}
}

// --- Sunpearl Kirin ------------------------------------------------------------

func TestTdmS3SunpearlKirinBouncesATokenAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tdmS3Library(me, 2)
	token := b12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	tdmSTCast(t, g, "Sunpearl Kirin", "Creature — Kirin", tdmS3SunpearlKirinOracle, 2, 1)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not returned")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn (the token ceased to exist)", hand, me.Hand.Size())
	}
}

func TestTdmS3SunpearlKirinBouncesACardWithoutDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tdmS3Library(me, 2)
	bear := seedBear(g, me.ID)
	theirs := seedBear(g, tdmS3Opp(g).ID)
	kirin := tdmSTCast(t, g, "Sunpearl Kirin", "Creature — Kirin", tdmS3SunpearlKirinOracle, 2, 1)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{"an opponent's creature": theirs, "the Kirin itself": kirin} {
		if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(id)); err == nil {
			t.Fatalf("the Kirin targeted %s", name)
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card was not returned to its owner's hand")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want only the returned card", hand, me.Hand.Size())
	}
}

// --- Surrak, Elusive Hunter ----------------------------------------------------

func TestTdmS3SurrakDrawsWhenAnOpponentTargetsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	tdmS3Library(me, 3)
	b12Push(g, me.ID, "Surrak, Elusive Hunter", "Legendary Creature — Human Warrior", tdmS3SurrakOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 2, 5)
	advanceToMain(t, g)
	// Your own Bolt at your own creature draws nothing.
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d -> %d after your own spell targeted your creature, want no draw", hand, me.Hand.Size())
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after an opponent targeted your creature, want one draw", hand, me.Hand.Size())
	}
}

func TestTdmS3SurrakDrawsWhenAnOpponentTargetsYourCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], tdmS3Opp(g)
	tdmS3Library(me, 3)
	b12Push(g, me.ID, "Surrak, Elusive Hunter", "Legendary Creature — Human Warrior", tdmS3SurrakOracle, 4, 3)
	bear := tdmSTCast(t, g, "Bear Spell", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	id := handCardFull(opp, "Syncopate", "Instant", "{X}{U}", syncopateOracle, []string{"U"})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{XValue: 0, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("Syncopate at the creature spell: %v", err)
	}
	passUntilTaxed(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after an opponent targeted your creature spell, want one draw", hand, me.Hand.Size())
	}
}

func TestTdmS3SurrakCantBeCountered(t *testing.T) {
	spec, ok := Lookup(tdmS3SurrakOracle)
	if !ok || !spec.CantBeCountered || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Errorf("Surrak: can't be countered %v, keywords %v", spec.CantBeCountered, spec.PrintedKeywords)
	}
}
