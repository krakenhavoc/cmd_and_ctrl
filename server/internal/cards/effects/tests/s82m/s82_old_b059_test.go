package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b059_test.go — Kaldheim's Saga batch 059 (#5505): each test plays
// the real catalog entry through the real chapter, cast and combat paths.

const (
	oldB059ForgingOracle   = "47a8f3a8-253b-4df9-9cfc-ff7df65d850c"
	oldB059FallOracle      = "54deeeef-be90-4975-9213-8355ddf3ce0c"
	oldB059FrostFireOracle = "5a7b89a6-d42d-4fd4-8754-d9c3e3df2118"
	oldB059ShowdownOracle  = "6107f9aa-3373-4166-b4b8-a26fd6d69c72"
	oldB059BearsOracle     = "697f3fd9-1577-4f69-bdf8-db1f673d3505"
	oldB059KardurOracle    = "6ca1d5ca-cc97-4783-9c0a-f47097f35ba8"
	oldB059RavenOracle     = "74206de7-470b-4256-9207-9fb46d3c4356"
	oldB059HeistOracle     = "78f99c66-cb50-49e3-ab2d-e7755818ffd7"
)

func TestOldB059CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB059ForgingOracle:   "Forging the Tyrite Sword",
		oldB059FallOracle:      "Fall of the Impostor",
		oldB059FrostFireOracle: "Battle of Frost and Fire",
		oldB059ShowdownOracle:  "Showdown of the Skalds",
		oldB059BearsOracle:     "The Bears of Littjara",
		oldB059KardurOracle:    "Kardur's Vicious Return",
		oldB059RavenOracle:     "The Raven's Warning",
		oldB059HeistOracle:     "The Trickster-God's Heist",
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

// oldB059Saga puts the Saga on the battlefield with `lore` counters, in the
// active player's main phase, so the next lore counter is chapter lore+1.
func oldB059Saga(t *testing.T, g *game.Game, me *game.Player, name, oracle string, lore int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Saga", OracleID: oracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: lore},
	})
}

// oldB059NextChapter adds the next lore counter, which triggers the chapter.
func oldB059NextChapter(g *game.Game, saga uuid.UUID) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
}

func oldB059Seats(g *game.Game) (me, opp *game.Player) {
	me, opp, _ = effectstest.BlbB1Seats(g)
	return me, opp
}

func TestOldB059ForgingTheTyriteSwordMakesTwoTreasuresThenSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB059Seats(g)
	saga := effectstest.CastCatalogSpell(t, g, "Forging the Tyrite Sword", "Enchantment — Saga", oldB059ForgingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("chapter I: %d Treasures, want 1", n)
	}
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("chapter II: %d Treasures, want 2", n)
	}
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Halvar, God of Battle", TypeLine: "Legendary Creature — God"},
		game.Card{Name: "Test Armor", TypeLine: "Artifact"},
	)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("chapter III opened no search: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Test Blade") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Halvar, God of Battle") == uuid.Nil {
		t.Error("the Equipment and Halvar must both be offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Test Armor") != uuid.Nil {
		t.Error("a Bear and a non-Equipment artifact must not be offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Halvar, God of Battle")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasNamed(me, "Halvar, God of Battle") {
		t.Error("Halvar did not go to hand")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestOldB059FallOfTheImpostorCountersThenExilesTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	saga := oldB059Saga(t, g, me, "Fall of the Impostor", oldB059FallOracle, 0)
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("chapter I: %d counters, want 1", n)
	}
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Fatalf("chapter II: %d counters, want 2", n)
	}
	big1 := effectstest.PushVanillaCreature(g, opp.ID, "Big One", 4, 4)
	big2 := effectstest.PushVanillaCreature(g, opp.ID, "Big Two", 4, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 1, 1)
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pickChoice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pickChoice == nil {
		t.Fatalf("chapter III asked nothing: %v", oldB059Kinds(g))
	}
	if err := g.ResolveTheirPermanents(pickChoice.ID, me.ID, []uuid.UUID{small}); err == nil {
		t.Fatal("a creature below the greatest power was accepted")
	}
	if err := g.ResolveTheirPermanents(pickChoice.ID, me.ID, []uuid.UUID{big2}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(big2) {
		t.Error("the chosen creature was not exiled")
	}
	if !g.Battlefield.Contains(big1) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(bear) {
		t.Error("another creature was exiled")
	}
}

func TestOldB059FallOfTheImpostorExilesALoneGreatestCreatureUnasked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 5, 5)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 6, 6)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 1, 1)
	saga := oldB059Saga(t, g, me, "Fall of the Impostor", oldB059FallOracle, 2)
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID) != nil {
		t.Fatal("a lone greatest creature was asked about")
	}
	if !g.Exile.Contains(big) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("only the opponent's greatest creature should be exiled")
	}
}

func TestOldB059BattleOfFrostAndFireSweepsScriesAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	giant := effectstest.B12Creature(g, me.ID, "Test Giant", "Creature — Giant", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	pw := effectstest.B12Creature(g, opp.ID, "Test Walker", "Planeswalker — Test", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pw {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 6}
			}
		}
	})
	saga := effectstest.CastCatalogSpell(t, g, "Battle of Frost and Fire", "Enchantment — Saga", oldB059FrostFireOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(giant) {
		t.Error("a Giant died to chapter I")
	}
	for _, id := range []uuid.UUID{bear, ogre} {
		if g.Battlefield.Contains(id) {
			t.Error("a non-Giant creature survived 4 damage")
		}
	}
	if n := effectstest.CounterCount(g, pw, game.CounterLoyalty); n != 2 {
		t.Errorf("the planeswalker has %d loyalty, want 2", n)
	}
	// II: scry 3.
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatalf("chapter II asked no scry: %+v", g.PendingChoices)
	}
}

func TestOldB059BattleOfFrostAndFireDrawsTwoDiscardsOneOnEachBigSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB059Seats(g)
	saga := oldB059Saga(t, g, me, "Battle of Frost and Fire", oldB059FrostFireOracle, 2)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Fatal("the Saga should be sacrificed after chapter III")
	}
	effectstest.CastFromHand(t, g, game.Card{Name: "Small Spell", TypeLine: "Sorcery", ManaCost: "{4}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	hand := effectstest.HandSize(me)
	for i := 0; i < 2; i++ {
		effectstest.CastFromHand(t, g, game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{5}"}, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.DiscardChoiceFor(g, me.ID) == nil {
			t.Fatalf("cast %d: no discard prompt after drawing two", i)
		}
		effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Each cast nets +2 drawn and -1 discarded (the spell itself was
	// pushed into the hand and cast again).
	if got := effectstest.HandSize(me); got != hand+2 {
		t.Errorf("hand %d -> %d after two big spells, want +2", hand, got)
	}
}

func TestOldB059ShowdownOfTheSkaldsExilesFourThenPutsCountersOnEachCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB059Seats(g)
	top := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	saga := effectstest.CastCatalogSpell(t, g, "Showdown of the Skalds", "Enchantment — Saga", oldB059ShowdownOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range top[:4] {
		if !g.Exile.Contains(id) || !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %d is not exiled and playable", i)
		}
	}
	if g.Exile.Contains(top[4]) {
		t.Error("the fifth card was exiled")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		effectstest.CastFromHand(t, g, game.Card{Name: "Cantrip", TypeLine: "Instant"}, game.CastSpellParams{})
		effectstest.AnswerPickTarget(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after two casts, want 2", n)
	}
}

func TestOldB059ShowdownOfTheSkaldsChapterThreeAlsoCountersEachCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB059Seats(g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	saga := oldB059Saga(t, g, me, "Showdown of the Skalds", oldB059ShowdownOracle, 2)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Fatal("the Saga should be sacrificed after chapter III")
	}
	effectstest.CastFromHand(t, g, game.Card{Name: "Cantrip", TypeLine: "Instant"}, game.CastSpellParams{})
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after chapter III's cast, want 1", n)
	}
}

func TestOldB059TheBearsOfLittjaraMakesAShapeshifterResizesItAndBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	saga := effectstest.CastCatalogSpell(t, g, "The Bears of Littjara", "Enchantment — Saga", oldB059BearsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var shape uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Shapeshifter" {
			shape = c.InstanceID
		}
	}
	if shape == uuid.Nil {
		t.Fatal("chapter I made no Shapeshifter")
	}
	if effectstest.EffectivePower(t, g, shape) != 2 || !effectstest.HasEffectiveKeyword(t, g, shape, "changeling") {
		t.Error("the token is not a 2/2 with changeling")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 5, 5)
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Changeling", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1})
	effectstest.PushVanillaCreature(g, me.ID, "Three", 3, 3)
	oldB059NextChapter(g, saga)
	effectstest.MscM2PickTargets(t, g, shape)
	if effectstest.EffectivePower(t, g, shape) != 4 {
		t.Fatalf("the Shapeshifter has power %d, want 4", effectstest.EffectivePower(t, g, shape))
	}
	if effectstest.EffectivePower(t, g, theirs) != 1 || effectstest.EffectivePower(t, g, bear) != 5 {
		t.Error("a creature that was not chosen changed size")
	}
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 10)
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok {
		t.Fatal("the 10-toughness creature died to 9 damage")
	}
	if c.DamageMarked != 9 {
		t.Errorf("%d damage marked, want 9 (the 4/4 and the 5/5, not the 3/3)", c.DamageMarked)
	}
}

func TestOldB059KardursViciousReturnSacrificesForDamageDiscardsAndReanimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	saga := effectstest.CastCatalogSpell(t, g, "Kardur's Vicious Return", "Enchantment — Saga", oldB059KardurOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	life := opp.Life
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	// II: every player discards.
	meHand, oppHand := effectstest.HandSize(me), effectstest.HandSize(opp)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if effectstest.DiscardChoiceFor(g, p.ID) != nil {
			effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != meHand-1 || effectstest.HandSize(opp) != oppHand-1 {
		t.Errorf("hands %d/%d -> %d/%d, want one fewer each", meHand, oppHand, effectstest.HandSize(me), effectstest.HandSize(opp))
	}
	// III: reanimate with a counter and haste.
	dead := effectstest.PushLibraryCard(me, game.Card{Name: "Dead Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	g.WithWriteLock(func() {
		z := g.FindCardZoneForEffect(dead)
		c, _ := z.Remove(dead)
		me.Graveyard.PushTop(c)
	})
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTarget(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature did not return")
	}
	if n := effectstest.CounterCount(g, dead, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("the creature has no haste")
	}
	seat := g.Turn.ActiveSeat
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	if effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("haste outlasted my next turn's start")
	}
}

func TestOldB059KardursViciousReturnDeclinedSacrificeDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Kardur's Vicious Return", "Enchantment — Saga", oldB059KardurOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || opp.Life != life {
		t.Error("declining the sacrifice still did something")
	}
}

func TestOldB059TheRavensWarningMakesABirdAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB059Seats(g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "The Raven's Warning", "Enchantment — Saga", oldB059RavenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var bird uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Bird" {
			bird = c.InstanceID
		}
	}
	if bird == uuid.Nil || !effectstest.HasEffectiveKeyword(t, g, bird, "flying") || effectstest.EffectivePower(t, g, bird) != 1 {
		t.Error("chapter I did not make a 1/1 Bird with flying")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestOldB059TheRavensWarningIgnoresCreaturesWithoutFlyingAndEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	groundling := effectstest.PushVanillaCreature(g, me.ID, "Groundling", 2, 2)
	saga := oldB059Saga(t, g, me, "The Raven's Warning", oldB059RavenOracle, 1)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	hand := effectstest.HandSize(me)
	effectstest.AttackWith(t, g, opp.ID, groundling)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand {
		t.Fatal("a creature without flying triggered chapter II")
	}
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	// The delayed trigger lasted only the turn it was made.
	if effectstest.HandSize(me) != hand {
		t.Fatalf("hand changed across turns: %d", effectstest.HandSize(me))
	}
}

func TestOldB059TheRavensWarningFlyersDrawOnceForTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	f1 := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Flyer One", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	f2 := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Flyer Two", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	saga := oldB059Saga(t, g, me, "The Raven's Warning", oldB059RavenOracle, 1)
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	hand := effectstest.HandSize(me)
	effectstest.AttackWith(t, g, opp.ID, f1, f2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d -> %d after two flyers connected, want +1", hand, got)
	}
	for _, c := range opp.Hand.Cards {
		if !c.IsKnownTo(me.ID) {
			t.Errorf("%s in the opponent's hand was not looked at", c.Name)
		}
	}
}

func TestOldB059TheRavensWarningPutsACardFromOutsideTheGameOnTop(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Wish Target", "Instant"), effectstest.OutsideCard("Other", "Sorcery"))
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Raven's Warning", TypeLine: "Enchantment — Saga", OracleID: oldB059RavenOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("chapter III asked nothing: %v", oldB059Kinds(g))
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 || len(c.ChooseCards) != 2 {
		t.Errorf("bounds %d..%d over %d cards, want 0..1 over both", c.ChooseMin, c.ChooseMax, len(c.ChooseCards))
	}
	libBefore := me.Library.Size()
	want := effectstest.SideCardID(me, "Wish Target")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{want}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	top, _ := me.Library.Top()
	if top.InstanceID != want || me.Library.Size() != libBefore+1 {
		t.Errorf("the top of the library is %q (size %d -> %d), want Wish Target on top", top.Name, libBefore, me.Library.Size())
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Other"}) {
		t.Errorf("sideboard %v, want only Other", got)
	}
}

func TestOldB059TheRavensWarningMayDeclineTheOutsideCard(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Wish Target", "Instant"))
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Raven's Warning", TypeLine: "Enchantment — Saga", OracleID: oldB059RavenOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	oldB059NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	libBefore := me.Library.Size()
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore || len(me.SideboardCards()) != 1 {
		t.Error("declining moved a card")
	}
}

func TestOldB059TheTricksterGodsHeistExchangesCreaturesOnAYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 5, 5)
	effectstest.CastCatalogSpell(t, g, "The Trickster-God's Heist", "Enchantment — Saga", oldB059HeistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MscM2PickTargets(t, g, mine, theirs)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != opp.ID {
		t.Error("my creature did not change hands")
	}
	if effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("their creature did not change hands")
	}
}

func TestOldB059TheTricksterGodsHeistDeclinedExchangeKeepsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 5, 5)
	effectstest.CastCatalogSpell(t, g, "The Trickster-God's Heist", "Enchantment — Saga", oldB059HeistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MscM2PickTargets(t, g, mine, theirs)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != me.ID || effectstest.ControllerOf(t, g, theirs) != opp.ID {
		t.Error("control changed after a no")
	}
}

func TestOldB059TheTricksterGodsHeistChapterTwoNeedsNonbasicNoncreaturePermanentsSharingAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	rock := effectstest.B12Permanent(g, me.ID, "Mine Rock", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	basic := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirCave := effectstest.B12Permanent(g, opp.ID, "Their Cave", "Land")
	myCave := effectstest.B12Permanent(g, me.ID, "Mine Cave", "Land")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	saga := oldB059Saga(t, g, me, "The Trickster-God's Heist", oldB059HeistOracle, 1)
	oldB059NextChapter(g, saga)
	// The first pick is offered the nonbasic, noncreature permanents only.
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("chapter II asked for no targets: %v", oldB059Kinds(g))
	}
	for _, bad := range []uuid.UUID{basic, bear} {
		if effectstest.HasID(pick.PickTargetCards, bad) {
			t.Errorf("an ineligible permanent %s was offered", bad)
		}
	}
	for _, ok := range []uuid.UUID{rock, theirRock, theirCave, myCave} {
		if !effectstest.HasID(pick.PickTargetCards, ok) {
			t.Errorf("an eligible permanent %s was not offered", ok)
		}
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: rock}); err != nil {
		t.Fatalf("first target: %v", err)
	}
	// The second target must share a card type with the first: an artifact.
	second := effectstest.LatestPickTarget(g, me.ID)
	if second == nil {
		t.Fatalf("no second target prompt: %v", oldB059Kinds(g))
	}
	if effectstest.HasID(second.PickTargetCards, theirCave) || effectstest.HasID(second.PickTargetCards, myCave) || effectstest.HasID(second.PickTargetCards, rock) {
		t.Error("a land, or the first target again, was offered as the second target")
	}
	if !effectstest.HasID(second.PickTargetCards, theirRock) {
		t.Fatal("the other artifact was not offered")
	}
	if err := g.ResolvePickTarget(second.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirRock}); err != nil {
		t.Fatalf("second target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, rock) != opp.ID || effectstest.ControllerOf(t, g, theirRock) != me.ID {
		t.Error("the artifacts did not trade hands")
	}
}

func TestOldB059TheTricksterGodsHeistChapterThreeDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB059Seats(g)
	saga := oldB059Saga(t, g, me, "The Trickster-God's Heist", oldB059HeistOracle, 2)
	mine, theirs := me.Life, opp.Life
	oldB059NextChapter(g, saga)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("life %d/%d -> %d/%d, want -3 and +3", mine, theirs, me.Life, opp.Life)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed")
	}
}

func oldB059Kinds(g *game.Game) []string {
	var out []string
	for _, c := range g.PendingChoices {
		if c != nil {
			out = append(out, string(c.Kind)+"/"+c.Chooser.String()[:4])
		}
	}
	return out
}
