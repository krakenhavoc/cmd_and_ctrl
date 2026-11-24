package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b042_cards_test.go — Strixhaven: School of Mages, batch 042 (S82,
// #5502). Each test plays the real catalog entry through the engine.

const (
	oldB42LoreholdCommandOracle = "107becb1-f8a3-4777-b39f-a0456c2d580e"
	oldB42WoeEaterOracle        = "290323d7-d8cb-4706-87d9-b920a4e2f680"
	oldB42IntroOracle           = "60e36f86-e284-468e-8073-07f19a3ed65e"
	oldB42TeachingsOracle       = "6ac3abba-cf55-4d97-a253-bce45972b050"
	oldB42GoldenRatioOracle     = "724700ad-1e4e-4cbc-af7f-0b3b87ed4bc8"
	oldB42CultivatorOracle      = "73a63c92-6dce-4804-bcd4-3fcc942cf270"
	oldB42RushedRebirthOracle   = "8ee64f9d-54c4-4a75-a8ac-6bb0a060b960"
)

func TestOldB42EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB42LoreholdCommandOracle: "Lorehold Command",
		oldB42WoeEaterOracle:        "Daemogoth Woe-Eater",
		oldB42IntroOracle:           "Introduction to Annihilation",
		oldB42TeachingsOracle:       "Teachings of the Archaics",
		oldB42GoldenRatioOracle:     "Golden Ratio",
		oldB42CultivatorOracle:      "Quandrix Cultivator",
		oldB42RushedRebirthOracle:   "Rushed Rebirth",
	}
	for oracle, name := range want {
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

// oldB42Cast casts a catalog spell from the hand, paying `mana` exactly.
func oldB42Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	id := effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB42Hand(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func oldB42EmptyHand(p *game.Player) {
	for p.Hand.Size() > 0 {
		p.Hand.Cards = p.Hand.Cards[:0]
	}
}

func oldB42Teachings(t *testing.T, mine, theirs int) int {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for _, seat := range g.Seats {
		oldB42EmptyHand(seat)
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	oldB42Hand(me, mine)
	oldB42Hand(opp, theirs)
	// Teachings itself leaves the hand as it is cast, so it never counts.
	before := me.Hand.Size()
	oldB42Cast(t, g, me, "Teachings of the Archaics", "Sorcery — Lesson", "{2}{U}", oldB42TeachingsOracle, "CCU", game.CastSpellParams{})
	return me.Hand.Size() - before
}

func TestOldB42TeachingsDrawsByHowFarAnOpponentLeads(t *testing.T) {
	cases := []struct{ mine, theirs, drew int }{
		{3, 3, 0}, // a tie draws nothing
		{4, 3, 0},
		{3, 4, 2},
		{3, 6, 2},
		{3, 7, 3}, // at least four more: three cards, not five
		{0, 9, 3},
	}
	for _, c := range cases {
		if got := oldB42Teachings(t, c.mine, c.theirs); got != c.drew {
			t.Errorf("hands %d vs %d: drew %d, want %d", c.mine, c.theirs, got, c.drew)
		}
	}
}

func TestOldB42GoldenRatioCountsDifferentPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.B12Creature(g, me.ID, "Two A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Two B", "Creature — Bear", 2, 3)
	effectstest.B12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, me.ID, "Zero", "Creature — Wall", 0, 4)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 7, 7)
	before := me.Hand.Size()
	oldB42Cast(t, g, me, "Golden Ratio", "Sorcery", "{1}{G}{U}", oldB42GoldenRatioOracle, "CGU", game.CastSpellParams{})
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d cards, want 3 (powers 0, 2 and 3)", got)
	}
}

func TestOldB42IntroductionExilesAndItsControllerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.SeedLibrary(me, "A", "B", "C")
	target := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCC")
	effectstest.CastFromHandForTest(t, g, me, "Introduction to Annihilation", "Sorcery — Lesson", "{5}", oldB42IntroOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, target); ok {
		t.Error("the target is still on the battlefield")
	}
	if got := opp.Hand.Size() - theirHand; got != 1 {
		t.Errorf("its controller drew %d, want 1", got)
	}
	if got := me.Hand.Size() - myHand; got != 0 {
		t.Errorf("the caster drew %d, want 0", got)
	}
}

func TestOldB42IntroductionCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Introduction to Annihilation", TypeLine: "Sorcery — Lesson", ManaCost: "{5}",
		OracleID: oldB42IntroOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	if err == nil {
		t.Fatal("a land was a legal target")
	}
}

func TestOldB42QuandrixCultivatorFetchesAForestOrIslandUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(forest)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Quandrix Cultivator", "Creature — Turtle Druid", oldB42CultivatorOracle, 3, 4)
	effectstest.AnswerSearchByID(t, g, me.ID, forest.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, forest.InstanceID)
	if !ok {
		t.Fatal("the Forest did not enter")
	}
	if c.Tapped {
		t.Error("the Forest entered tapped; the card does not say tapped")
	}
}

func oldB42Creature(owner uuid.UUID, name, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost, Power: 1, Toughness: 1, Owner: owner, Controller: owner}
}

func TestOldB42RushedRebirthFetchesALesserCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, oldB42Creature(me.ID, "Victim", "{1}{G}"))
	cheap := oldB42Creature(me.ID, "Cheap", "{G}")
	equal := oldB42Creature(me.ID, "Equal", "{1}{G}")
	big := oldB42Creature(me.ID, "Big", "{4}{G}")
	for _, c := range []game.Card{big, equal, cheap} {
		me.Library.PushTop(c)
	}
	effectstest.CastCatalogSpell(t, g, "Rushed Rebirth", "Instant", oldB42RushedRebirthOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, cheap.InstanceID); ok {
		t.Fatal("nothing has died yet, but the search already happened")
	}
	effectstest.P1Destroy(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, cheap.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, cheap.InstanceID)
	if !ok || !c.Tapped {
		t.Errorf("the cheaper creature: on battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
	for _, other := range []uuid.UUID{equal.InstanceID, big.InstanceID} {
		if _, ok := effectstest.BattlefieldCard(g, other); ok {
			t.Error("a creature with an equal or greater mana value was fetched")
		}
	}
}

func TestOldB42RushedRebirthDoesNothingForACreatureThatDoesNotDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, oldB42Creature(me.ID, "Victim", "{1}{G}"))
	effectstest.CastCatalogSpell(t, g, "Rushed Rebirth", "Instant", oldB42RushedRebirthOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if len(g.PendingChoices) != 0 {
		t.Error("a search is pending though the creature never died")
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); !ok {
		t.Error("the creature left the battlefield")
	}
}

func TestOldB42WoeEaterSacrificesAtUpkeepAndPaysOffWhenItIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	oldB42Hand(opp, 2)
	woe := effectstest.PushCatalogPermanent(g, me.ID, "Daemogoth Woe-Eater", "Creature — Demon", oldB42WoeEaterOracle, false)
	hand, life, oppHand := me.Hand.Size(), me.Life, opp.Hand.Size()
	// Sacrificing it by any means pays off the second ability.
	g.WithWriteLock(func() {
		g.PlayerSacrificesForEffect(woe, me.ID, effects.SacrificeSpec("a creature", effects.Creature()), "test sacrifice")
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, woe)
	effectstest.PassPriorityAroundTable(t, g)
	for _, seat := range g.Seats[1:] {
		if seat.Hand.Size() > 0 {
			effectstest.AnswerDiscard(t, g, seat.ID, seat.Hand.Cards[0].InstanceID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if opp.Hand.Size() != oppHand-1 {
		t.Errorf("the opponent's hand %d, want %d", opp.Hand.Size(), oppHand-1)
	}
}

func TestOldB42WoeEaterDoesNotPayOffWhenItDiesAnotherWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	woe := effectstest.PushCatalogPermanent(g, me.ID, "Daemogoth Woe-Eater", "Creature — Demon", oldB42WoeEaterOracle, false)
	hand, life := me.Hand.Size(), me.Life
	effectstest.P1Destroy(t, g, woe)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("a destroyed Woe-Eater paid off: hand %d (was %d), life %d (was %d)", me.Hand.Size(), hand, me.Life, life)
	}
}

func TestOldB42WoeEaterUpkeepSacrificeMakesYouChooseACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.PushCatalogPermanent(g, me.ID, "Daemogoth Woe-Eater", "Creature — Demon", oldB42WoeEaterOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the chosen creature was not sacrificed")
	}
}

func TestOldB42LoreholdCommandTokenAndPumpIncludeTheNewSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1g3CastModal(t, g, "Lorehold Command", "Instant", oldB42LoreholdCommandOracle, []int{0, 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("the Bear's power is %d, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") || !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the Bear did not gain indestructible and haste")
	}
	var spirit *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Spirit" && g.Battlefield.Cards[i].IsToken() {
			spirit = &g.Battlefield.Cards[i]
		}
	}
	if spirit == nil {
		t.Fatal("no Spirit token")
	}
	if got := effectstest.EffectivePower(t, g, spirit.InstanceID); got != 4 {
		t.Errorf("the Spirit's power is %d, want 4 (3 +1/+0)", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, spirit.InstanceID, "haste") {
		t.Error("the Spirit made by the first bullet did not get haste")
	}
}

func TestOldB42LoreholdCommandDamagesAnyTargetAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	effectstest.P1g3CastModal(t, g, "Lorehold Command", "Instant", oldB42LoreholdCommandOracle, []int{2, 3},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0, Slot: 0},
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Mode: 0, Slot: 1})
	effectstest.PassPriorityAroundTable(t, g)
	victim := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	effectstest.AnswerSacrifice(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-3)
	}
	if me.Life != life+3 {
		t.Errorf("you are at %d, want %d", me.Life, life+3)
	}
}
