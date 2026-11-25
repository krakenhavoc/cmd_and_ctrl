package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Amonkhet batch 245 (#5535): the card draw and library cards.

const (
	akh245Shefet     = "14f081bb-dd7a-4419-af9b-f75d91e1a6d0"
	akh245Argument   = "686e0b86-121c-45be-842d-c83d3b6547f6"
	akh245Horror     = "6de20c2c-b8a8-4f12-897b-a2f55d7eed87"
	akh245Resolve    = "71c5718f-580d-4993-8492-81b7a557d10c"
	akh245Seeker     = "a0b2d446-fe90-40f8-8674-7421c425b489"
	akh245Serpent    = "d6f8c69f-e603-41f5-8f25-da8a102f7c8d"
	akh245Floodwater = "e89cce0e-4fd7-40fa-8a55-b07d72826956"
)

func TestAkh245CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		akh245Shefet: "Shefet Monitor", akh245Argument: "Compelling Argument", akh245Horror: "Horror of the Broken Lands",
		akh245Resolve: "Gideon's Resolve", akh245Seeker: "Seeker of Insight", akh245Serpent: "River Serpent",
		akh245Floodwater: "Floodwaters",
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

func TestAkh245ShefetMonitorCyclesAndFetchesABasicOrADesertOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Drawn", TypeLine: "Creature — Bear"},
		game.Card{Name: "Wasteland Cousin", TypeLine: "Land"},
		game.Card{Name: "Scorched Dunes", TypeLine: "Land — Desert"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	id, _ := effectstest.CycleFromHand(t, g, "Shefet Monitor", "Creature — Lizard", akh245Shefet, "{C}{C}{C}{G}")
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.SearchChoiceFor(g, me.ID)
	if search == nil {
		t.Fatalf("no search prompt after cycling: %+v", g.PendingChoices)
	}
	offered := map[uuid.UUID]bool{}
	for _, c := range search.SearchCards {
		offered[c] = true
	}
	if !offered[ids[2]] || !offered[ids[3]] || offered[ids[1]] || offered[ids[0]] {
		t.Errorf("offered %v: want the Desert and the Forest only", search.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, ids[2]) {
		t.Error("the Desert is not on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled Monitor is not in the graveyard")
	}
}

func TestAkh245ShefetMonitorMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Drawn", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CycleFromHand(t, g, "Shefet Monitor", "Creature — Lizard", akh245Shefet, "{C}{C}{C}{G}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, ids[1]) {
		t.Error("the Forest entered although the search found nothing")
	}
}

func TestAkh245CompellingArgumentMillsFiveAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Compelling Argument", "Sorcery", akh245Argument, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-5 || opp.Graveyard.Size() != yard+5 {
		t.Errorf("library %d (want %d), graveyard %d (want %d)", opp.Library.Size(), lib-5, opp.Graveyard.Size(), yard+5)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("my graveyard %d, want only the spell", me.Graveyard.Size())
	}
}

func TestAkh245CompellingArgumentCyclesForU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Compelling Argument", "Sorcery", akh245Argument, "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestAkh245HorrorOfTheBrokenLandsGrowsOnAnotherCardCycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	horror := effectstest.PushCatalogPermanent(g, me.ID, "Horror of the Broken Lands", "Creature — Horror", akh245Horror, false)
	p0, t0 := effectstest.EffectivePTOf(t, g, horror)
	// Cycling a card is one discard, so one trigger, not two.
	effectstest.CycleFromHand(t, g, "Floodwaters", "Sorcery", akh245Floodwater, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, horror); p != p0+2 || tough != t0+1 {
		t.Errorf("after a cycle %d/%d, want %d/%d", p, tough, p0+2, t0+1)
	}
}

func TestAkh245HorrorOfTheBrokenLandsGrowsOnAnotherCardDiscardedButNotAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	horror := effectstest.PushCatalogPermanent(g, me.ID, "Horror of the Broken Lands", "Creature — Horror", akh245Horror, false)
	p0, t0 := effectstest.EffectivePTOf(t, g, horror)
	effectstest.S58HandCards(opp, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, horror); p != p0 {
		t.Fatalf("an opponent's discard grew the Horror to %d power", p)
	}
	effectstest.S58HandCards(me, 2)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, horror); p != p0+2 || tough != t0+1 {
		t.Errorf("after a discard %d/%d, want %d/%d", p, tough, p0+2, t0+1)
	}
}

func TestAkh245HorrorDoesNotTriggerOnItsOwnCycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	id, _ := effectstest.CycleFromHand(t, g, "Horror of the Broken Lands", "Creature — Horror", akh245Horror, "{B}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled Horror is not in the graveyard")
	}
}

func TestAkh245GideonsResolveFetchesGideonAndPumpsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 3, 3)
	effectstest.SeedSearchLibrary(me)
	gideon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Gideon, Martial Paragon", TypeLine: "Legendary Planeswalker — Gideon"})
	effectstest.PayMana(t, g, me, "{C}{C}{C}{C}{W}")
	effectstest.CastFromHandForTest(t, g, me, "Gideon's Resolve", "Enchantment", "{4}{W}", akh245Resolve, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, gideon)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gideon) {
		t.Error("Gideon, Martial Paragon is not in hand")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("my Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("their Bear is %d/%d, want 3/3", p, tough)
	}
}

func TestAkh245SeekerOfInsightLootsOnlyAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	seeker := effectstest.PushCatalogPermanent(g, me.ID, "Seeker of Insight", "Creature — Human Wizard", akh245Seeker, false)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Seeker activated before a noncreature spell was cast")
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Seeker after a noncreature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() < hand {
		t.Errorf("hand %d after looting, started at %d", me.Hand.Size(), hand)
	}
}

func TestAkh245RiverSerpentCantAttackUntilFiveCardsAreInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "River Serpent", OracleID: akh245Serpent,
		TypeLine: "Creature — Serpent", Power: 5, Toughness: 5})
	me.Graveyard.Cards = nil
	for i := 0; i < 4; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	effectstest.WantAttackRefused(t, g, serpent, them.ID, "fewer than five cards in their graveyard")
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: me.ID, Controller: me.ID})
	if err := effectstest.DeclareResult(g, serpent, them.ID); err != nil {
		t.Errorf("attack with five cards in the graveyard: %v", err)
	}
}

func TestAkh245RiverSerpentCountsOnlyItsControllersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "River Serpent", OracleID: akh245Serpent,
		TypeLine: "Creature — Serpent", Power: 5, Toughness: 5})
	me.Graveyard.Cards = nil
	for i := 0; i < 6; i++ {
		them.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: them.ID, Controller: them.ID})
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	effectstest.WantAttackRefused(t, g, serpent, them.ID, "fewer than five cards in their graveyard")
}

func TestAkh245FloodwatersReturnsUpToTwoCreaturesAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Floodwaters", "Sorcery", akh245Floodwater, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, a) || effectstest.BattlefieldHas(g, b) || !effectstest.BattlefieldHas(g, c) {
		t.Errorf("A on battlefield %v, B %v, C %v: want A and B returned, C left", effectstest.BattlefieldHas(g, a), effectstest.BattlefieldHas(g, b), effectstest.BattlefieldHas(g, c))
	}
	if !opp.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("a returned creature is not in its owner's hand")
	}
}

func TestAkh245FloodwatersCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	id, _ := effectstest.CycleFromHand(t, g, "Floodwaters", "Sorcery", akh245Floodwater, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
}
