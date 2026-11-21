package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b3_cards_test.go — Star Trek, batch 3 (#4973). Each test plays the
// real catalog entry and asserts what its text says.

const (
	trekB3JanewayOracle  = "b572be16-ad13-4f9f-bcf6-085f81e5292c"
	trekB3HolodeckOracle = "3f04c38b-d6bd-405f-b69c-009df87801dd"
	trekB3MeasureOracle  = "727e98cc-b295-47f9-9d3d-bcf38c9f56d2"
	trekB3MilesOracle    = "31e3f39f-8b18-4316-984f-4d57d2d480ca"
	trekB3NeedsOracle    = "984e0426-10ba-4c17-b26e-ce21100ec398"
	trekB3OberthOracle   = "fb5e53cf-e12e-469b-b9ea-9debc967e77e"
	trekB3PeaceOracle    = "b19ce6f2-7778-4779-8173-973eb7ee0b02"
	trekB3PicardOracle   = "4b36f8e2-ba7d-4bc7-890f-dd6ab73eef17"
	trekB3AdjunctOracle  = "be3dd0c8-32c2-4a56-9290-1fc9e366bae2"
)

func TestTrekB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB3JanewayOracle:  "Hologram Janeway",
		trekB3HolodeckOracle: "Malfunctioning Holodeck",
		trekB3MeasureOracle:  "The Measure of a Man",
		trekB3MilesOracle:    "Miles O'Brien, Transporter Ops",
		trekB3NeedsOracle:    "The Needs of the Many",
		trekB3OberthOracle:   "Oberth-Class Starship",
		trekB3PeaceOracle:    "Peace in Our Time",
		trekB3PicardOracle:   "Picard, Leading by Example",
		trekB3AdjunctOracle:  "Primary Unimatrix Adjunct",
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

// trekB3PickTargetOpen reports whether a pick_target prompt is open.
func trekB3PickTargetOpen(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}

// --- Hologram Janeway ---------------------------------------------------

func TestTrekB3HologramJanewayCountersWhenYourCreaturesTotalEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	janeway := effectstest.B27Push(g, me.ID, "Hologram Janeway", "Legendary Artifact Creature — Illusion Officer", trekB3JanewayOracle, "{2}{W}", 3, 3, "W")
	big := effectstest.PushVanillaCreature(g, me.ID, "Big Ally", 5, 5)
	home := effectstest.PushVanillaCreature(g, me.ID, "Home Guard", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(big, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, home)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, home); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
	if effectstest.PlusOnes(g, janeway) != 0 || effectstest.PlusOnes(g, big) != 0 {
		t.Error("a creature that was not the target got a counter")
	}
}

func TestTrekB3HologramJanewayNeedsEightTotalPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	janeway := effectstest.B27Push(g, me.ID, "Hologram Janeway", "Legendary Artifact Creature — Illusion Officer", trekB3JanewayOracle, "{2}{W}", 3, 3, "W")
	ally := effectstest.PushVanillaCreature(g, me.ID, "Ally", 4, 4)
	// An opponent's power is not "creatures you control".
	effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 9, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ally, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if trekB3PickTargetOpen(g) || effectstest.TriggersOnStackFrom(g, janeway) != 0 {
		t.Fatal("Janeway triggered with 7 total power")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, ally) != 0 || effectstest.PlusOnes(g, janeway) != 0 {
		t.Error("a counter was placed with 7 total power")
	}
}

// --- Malfunctioning Holodeck ------------------------------------------

func TestTrekB3MalfunctioningHolodeckExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	deck := effectstest.CastFromHand(t, g, game.Card{Name: "Malfunctioning Holodeck", TypeLine: "Artifact", OracleID: trekB3HolodeckOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("a permanent you control was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	effectstest.B27Kill(g, deck)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Their Bear") == uuid.Nil {
		t.Error("the exiled creature did not return when the Holodeck left")
	}
}

func TestTrekB3MalfunctioningHolodeckHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deck := effectstest.PushCatalogPermanent(g, me.ID, "Malfunctioning Holodeck", "Artifact", trekB3HolodeckOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Naturalize", TypeLine: "Instant", OracleID: effectstest.NaturalizeOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(deck)}); err != nil {
		t.Fatalf("an opponent's Naturalize at the Holodeck: %v", err)
	}
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment asked of the opponent")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(deck) {
		t.Error("the declined ward did not counter the Naturalize")
	}
}

// --- The Measure of a Man ---------------------------------------------

func TestTrekB3MeasureOfAManExilesInvestigatesAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Android", 3, 3)
	saga := effectstest.CastFromHand(t, g, game.Card{Name: "The Measure of a Man", TypeLine: "Enchantment — Saga", OracleID: trekB3MeasureOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("chapter I did not exile the creature")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Clue") == uuid.Nil {
		t.Fatal("chapter II did not investigate")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Their Android")
	if back == uuid.Nil {
		t.Fatal("chapter III did not return the exiled card")
	}
	if c := effectstest.LayeredCard(t, g, back); c.Controller != opp.ID {
		t.Errorf("the card returned under %v, want its owner %v", c.Controller, opp.ID)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestTrekB3MeasureOfAManExilesANontokenArtifactNotAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	rock := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	effectstest.CastFromHand(t, g, game.Card{Name: "The Measure of a Man", TypeLine: "Enchantment — Saga", OracleID: trekB3MeasureOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: goblin}); err == nil {
		t.Fatal("a token was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(goblin) || g.Battlefield.Contains(rock) {
		t.Fatal("want the nontoken artifact exiled and the token untouched")
	}
	_ = me
}

// --- Miles O'Brien, Transporter Ops ------------------------------------

func TestTrekB3MilesBouncesThenDrawsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	miles := effectstest.CastHoldCreature(t, g, "Miles O'Brien, Transporter Ops", "Legendary Creature — Human Artificer", trekB3MilesOracle, 3, 3)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: miles}); err == nil {
		t.Fatal("Miles was accepted as the other target")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Fatal("the artifact was not returned to its owner's hand")
	}
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d at the end step, want one card drawn", hand, got)
	}
}

func TestTrekB3MilesDrawsOnlyForYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Miles O'Brien, Transporter Ops", "Legendary Creature — Human Artificer", trekB3MilesOracle, "{2}{W}", 3, 3, "W")
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(theirs) })
	if !opp.Hand.Contains(theirs) {
		t.Fatal("the opponent's creature was not bounced")
	}
	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || opp.Hand.Size() != theirHand {
		t.Errorf("hands %d/%d -> %d/%d: a permanent returned to an opponent's hand draws nothing",
			hand, theirHand, me.Hand.Size(), opp.Hand.Size())
	}
}

func TestTrekB3MilesDrawsInAnOpponentsEndStepToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Miles O'Brien, Transporter Ops", "Legendary Creature — Human Artificer", trekB3MilesOracle, "{2}{W}", 3, 3, "W")
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(mine) })
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d in an opponent's end step, want one card", hand, got)
	}
}

// --- The Needs of the Many --------------------------------------------

func TestTrekB3NeedsOfTheManyProtectsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, me.ID, "Redshirt", 1, 1)
	keeper := effectstest.PushVanillaCreature(g, me.ID, "Spock", 2, 2)
	src := effectstest.PushVanillaCreature(g, opp.ID, "Khan", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	id := effectstest.B13HandCard(me, "The Needs of the Many", "Instant", trekB3NeedsOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no sacrifice was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, keeper, "hexproof") || !effectstest.HasEffectiveKeyword(t, g, keeper, "indestructible") {
		t.Error("creatures you control did not gain hexproof and indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, src, "hexproof") {
		t.Error("an opponent's creature gained hexproof")
	}
	var hexproof bool
	g.WithWriteLock(func() { hexproof = g.PlayerHasKeywordLocked(me, game.KeywordHexproof) })
	if !hexproof {
		t.Error("you did not gain hexproof")
	}
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Errorf("life %d -> %d, want all damage to you prevented", life, me.Life)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("The Needs of the Many was not exiled as it resolved")
	}
}

// --- Oberth-Class Starship --------------------------------------------

func TestTrekB3OberthFetchesABasicPlainsAndFliesAtSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.PushLibraryCard(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Sacred Foundry", TypeLine: "Land — Mountain Plains"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	ship := effectstest.CastFromHand(t, g, game.Card{Name: "Oberth-Class Starship", TypeLine: "Artifact — Spacecraft", OracleID: trekB3OberthOracle, ManaCost: "{1}{W}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(plains) {
		t.Fatal("the basic Plains was not put into your hand")
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("an unstationed Spacecraft is a creature")
	}
	crew := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Crew", TypeLine: "Creature — Human", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || !c.IsArtifact() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 2 {
		t.Errorf("at 6+ Oberth is a 2/2 flying artifact creature; creature=%v flying=%v power=%d",
			c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower())
	}
}

// --- Peace in Our Time --------------------------------------------------

// trekB3ToBeginCombatOf walks to `seat`'s beginning of combat, declining
// every Peace in Our Time question asked on the way.
func trekB3ToBeginCombatOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepBeginCombat && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			if !errors.Is(err, game.ErrChoicePending) {
				t.Fatalf("AdvanceStep: %v", err)
			}
			trekB3DeclineEveryConfirm(t, g)
		}
	}
	t.Fatalf("never reached seat %d's beginning of combat", seat)
}

func trekB3DeclineEveryConfirm(t *testing.T, g *game.Game) {
	t.Helper()
	for _, p := range g.Seats {
		if effectstest.ConfirmChoiceFor(g, p.ID) != nil {
			effectstest.AnswerMayChoice(t, g, p.ID, false)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestTrekB3PeaceInOurTimeDrawingStopsTheAttackAndAddsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peace := effectstest.PushCatalogPermanent(g, me.ID, "Peace in Our Time", "Enchantment", trekB3PeaceOracle, false)
	theirBear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	trekB3ToBeginCombatOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Fatal("the enchantment's controller was asked on an opponent's turn")
	}
	hand := opp.Hand.Size()
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the opponent's hand %d -> %d, want one card drawn", hand, opp.Hand.Size())
	}
	if got := effectstest.CounterOn(g, peace, "peace"); got != 1 {
		t.Errorf("peace counters = %d, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirBear, me.ID); err == nil {
		t.Error("the player who drew attacked anyway")
	}
}

func TestTrekB3PeaceInOurTimeDecliningLetsThemAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peace := effectstest.PushCatalogPermanent(g, me.ID, "Peace in Our Time", "Enchantment", trekB3PeaceOracle, false)
	theirBear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	trekB3ToBeginCombatOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	hand := opp.Hand.Size()
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand || effectstest.CounterOn(g, peace, "peace") != 0 {
		t.Error("a declined question drew a card or added a counter")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirBear, me.ID); err != nil {
		t.Errorf("the player who declined could not attack: %v", err)
	}
}

func TestTrekB3PeaceInOurTimeWinsAtTwelveCounters(t *testing.T) {
	for _, tc := range []struct {
		counters int
		win      bool
	}{{11, false}, {12, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		peace := effectstest.PushCatalogPermanent(g, me.ID, "Peace in Our Time", "Enchantment", trekB3PeaceOracle, false)
		effectstest.AddCountersForTest(g, peace, "peace", tc.counters)
		for i := 0; i < 400 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				if !errors.Is(err, game.ErrChoicePending) {
					t.Fatalf("AdvanceStep: %v", err)
				}
				trekB3DeclineEveryConfirm(t, g)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		won := g.State != game.StateActive && g.Outcome != nil && g.Outcome.Winner == me.ID
		if won != tc.win {
			t.Errorf("%d peace counters: won = %v (state %v, outcome %+v), want %v", tc.counters, won, g.State, g.Outcome, tc.win)
		}
	}
}

// --- Picard, Leading by Example ----------------------------------------

func TestTrekB3PicardPumpsAnotherCreatureByHisPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	picard := effectstest.B27Push(g, me.ID, "Picard, Leading by Example", "Legendary Creature — Human Officer", trekB3PicardOracle, "{1}{W}", 2, 2, "W")
	effectstest.AddCountersForTest(g, picard, game.CounterPlusOne, 1)
	ensign := effectstest.PushVanillaCreature(g, me.ID, "Ensign", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ensign), Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {1}")
	}
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(picard), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Picard targeting himself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ensign), Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, picard).Tapped {
		t.Error("Picard did not tap")
	}
	if p := effectstest.EffectivePower(t, g, ensign); p != 4 {
		t.Errorf("the Ensign's power = %d, want 1 + 3 (Picard's power)", p)
	}
}

func TestTrekB3PicardActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	picard := effectstest.B27Push(g, me.ID, "Picard, Leading by Example", "Legendary Creature — Human Officer", trekB3PicardOracle, "{1}{W}", 2, 2, "W")
	ensign := effectstest.PushVanillaCreature(g, me.ID, "Ensign", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ensign), Strict: true}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("activation in combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
}

// --- Primary Unimatrix Adjunct -----------------------------------------

func TestTrekB3AdjunctGivesYourArtifactCreaturesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adjunct := effectstest.B27Push(g, me.ID, "Primary Unimatrix Adjunct", "Artifact Creature — Borg", trekB3AdjunctOracle, "{3}{W}", 2, 3, "W")
	drone := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drone", TypeLine: "Artifact Creature — Borg", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	human := effectstest.PushVanillaCreature(g, me.ID, "Human", 2, 2)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Drone", TypeLine: "Artifact Creature — Borg", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	for id, want := range map[uuid.UUID]bool{adjunct: true, drone: true, human: false, rock: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "double strike"); got != want {
			t.Errorf("%v: double strike = %v, want %v", id, got, want)
		}
	}
}
