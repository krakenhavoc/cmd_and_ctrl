package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b060_test.go — Kaldheim's Saga batch 060 (#5505): each test plays
// the real catalog entry through the real chapter, choice and combat paths.

const (
	oldB060ArniOracle     = "7cb47a7e-fe60-45a2-9c07-bd5c6ddca82d"
	oldB060InvasionOracle = "807e35c0-8ea1-44a5-abac-51fd516aad03"
	oldB060SeasonsOracle  = "a38ebf4a-064c-4517-a553-15f0a252c159"
	oldB060TrollsOracle   = "c8cdabc4-e147-491b-9c44-96e57f6ec3b4"
	oldB060HaraldOracle   = "d0cc7e79-69dc-4f9d-9eed-1848f8ddb1be"
	oldB060NikoOracle     = "e79bf038-071e-448e-ad95-234c207a9575"
	oldB060BloodskyOracle = "fd9cd2b0-3f43-41a6-b963-c781d0f285e8"
)

func TestOldB060CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB060ArniOracle:     "Arni Slays the Troll",
		oldB060InvasionOracle: "Invasion of the Giants",
		oldB060SeasonsOracle:  "The Three Seasons",
		oldB060TrollsOracle:   "Waking the Trolls",
		oldB060HaraldOracle:   "Harald Unites the Elves",
		oldB060NikoOracle:     "Niko Defies Destiny",
		oldB060BloodskyOracle: "The Bloodsky Massacre",
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

// oldB060Saga puts the Saga on the battlefield with `lore` counters in the
// active player's main phase, so the next lore counter is chapter lore+1.
func oldB060Saga(t *testing.T, g *game.Game, me *game.Player, name, oracle string, lore int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Saga", OracleID: oracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: lore},
	})
}

func oldB060NextChapter(g *game.Game, saga uuid.UUID) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
}

func oldB060Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// oldB060Decline answers every open target prompt of `who` with no target.
func oldB060Decline(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for effectstest.LatestPickTarget(g, who) != nil {
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, who).ID, who, nil); err != nil {
			t.Fatalf("declining a clause: %v", err)
		}
	}
}

func TestOldB060ArniSlaysTheTrollFightsThenCountersThenGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 3, 3)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	saga := oldB060Saga(t, g, me, "Arni Slays the Troll", oldB060ArniOracle, 0)
	oldB060NextChapter(g, saga)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature died to a 2-power fight")
	}

	oldB060NextChapter(g, saga)
	effectstest.AnswerPickTarget(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, other, game.CounterPlusOne); n != 2 {
		t.Errorf("chapter II put %d counters, want 2", n)
	}
	if n := me.ManaPool.CountColor("R"); n != 1 {
		t.Errorf("chapter II added %d {R}, want 1", n)
	}

	life := me.Life
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("chapter III: life %d -> %d, want +3 (greatest power 3)", life, me.Life)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestOldB060ArniChapterOneWithNoOpposingCreatureDoesNotFightAndChapterTwoStillAddsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 3, 3)
	saga := oldB060Saga(t, g, me, "Arni Slays the Troll", oldB060ArniOracle, 0)
	oldB060NextChapter(g, saga)
	effectstest.PickCard(t, g, me.ID, mine)
	oldB060Decline(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("power %d", p)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	g.WithWriteLock(func() { g.Battlefield.Remove(mine) })
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("R"); n != 1 {
		t.Errorf("chapter II with no creature added %d {R}, want 1", n)
	}
}

func TestOldB060InvasionOfTheGiantsRevealsAGiantForDamageThenDiscountsGiants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	giant := effectstest.PutInHand(me, game.Card{Name: "Hill Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 3, Toughness: 3})
	effectstest.PutInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	saga := oldB060Saga(t, g, me, "Invasion of the Giants", oldB060InvasionOracle, 1)
	life := opp.Life
	handBefore := len(me.Hand.Cards)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d -> %d, want a card drawn", handBefore, len(me.Hand.Cards))
	}
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no reveal prompt")
	}
	if !effectstest.HasID(c.ChooseCards, giant) || len(c.ChooseCards) != 1 {
		t.Errorf("the offered cards are %v, want only the Giant", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
	if !effectstest.InHandOf(me, giant) {
		t.Error("the revealed Giant left the hand")
	}
}

func TestOldB060InvasionOfTheGiantsAllowsDecliningTheReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	effectstest.PutInHand(me, game.Card{Name: "Hill Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 3, Toughness: 3})
	saga := oldB060Saga(t, g, me, "Invasion of the Giants", oldB060InvasionOracle, 1)
	life := opp.Life
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("declining the reveal still dealt damage: %d -> %d", life, opp.Life)
	}
}

func TestOldB060InvasionOfTheGiantsWithNoGiantAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	effectstest.PutInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	saga := oldB060Saga(t, g, me, "Invasion of the Giants", oldB060InvasionOracle, 1)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
		t.Errorf("a reveal prompt opened with no Giant in hand: %+v", c)
	}
}

func TestOldB060InvasionOfTheGiantsChapterThreeDiscountsOnlyTheNextGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	saga := oldB060Saga(t, g, me, "Invasion of the Giants", oldB060InvasionOracle, 2)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	giant := effectstest.HandCardForTest(me, "Hill Giant", "Creature — Giant", "")
	bear := effectstest.HandCardForTest(me, "Hill Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{giant, bear} {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].ManaCost = "{2}{R}"
					me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 3, 3
				}
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil || strings.Contains(err.Error(), "sorcery speed") {
		t.Fatalf("a non-Giant spell was discounted, or the cast failed for another reason: %v", err)
	}
	if err := g.CastSpell(me.ID, giant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Giant was not discounted by {2}: %v", err)
	}
}

func TestOldB060TheThreeSeasonsMillsReturnsSnowAndShufflesThreeFromEachGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Creature — Bear")
	}
	saga := oldB060Saga(t, g, me, "The Three Seasons", oldB060SeasonsOracle, 0)
	gy := len(me.Graveyard.Cards)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Graveyard.Cards) != gy+3 {
		t.Errorf("chapter I milled %d, want 3", len(me.Graveyard.Cards)-gy)
	}

	snowA := effectstest.PushGraveyardCardTyped(me, "Snow Land", "Snow Land")
	snowB := effectstest.PushGraveyardCardTyped(me, "Snow Bear", "Snow Creature — Bear")
	snowC := effectstest.PushGraveyardCardTyped(me, "Snow Beast", "Snow Creature — Beast")
	plain := effectstest.PushGraveyardCardTyped(me, "Plain Bear", "Creature — Bear")
	instant := effectstest.PushGraveyardCardTyped(me, "Snow Spell", "Snow Instant")
	oldB060NextChapter(g, saga)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatalf("no target prompt")
	} else {
		for _, bad := range []uuid.UUID{plain, instant} {
			if effectstest.HasID(p.PickTargetCards, bad) {
				t.Errorf("a non-snow-permanent card %s was offered", bad)
			}
		}
		for _, ok := range []uuid.UUID{snowA, snowB, snowC} {
			if !effectstest.HasID(p.PickTargetCards, ok) {
				t.Errorf("a snow permanent card %s was not offered", ok)
			}
		}
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, snowA, snowB)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, snowA) || !effectstest.InHandOf(me, snowB) {
		t.Error("the two chosen snow permanents did not return to hand")
	}
	if effectstest.InHandOf(me, snowC) {
		t.Error("a third card returned")
	}

	// Chapter III: three from each graveyard, owners shuffle them in.
	for i := 0; i < 5; i++ {
		effectstest.PushGraveyardCardTyped(opp, "Theirs", "Creature — Bear")
	}
	myGy := len(me.Graveyard.Cards)
	myLib, theirLib := len(me.Library.Cards), len(opp.Library.Cards)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.LatestChooseCardsFor(g, me.ID)
	if first == nil {
		t.Fatalf("chapter III asked nothing")
	}
	if first.ChooseMin != 3 || first.ChooseMax != 3 {
		t.Errorf("first prompt asks %d-%d cards, want exactly 3", first.ChooseMin, first.ChooseMax)
	}
	pickFrom := func(c *game.PendingChoice) []uuid.UUID { return append([]uuid.UUID(nil), c.ChooseCards[:3]...) }
	mineChosen := pickFrom(first)
	effectstest.AnswerChooseCards(t, g, me.ID, mineChosen...)
	second := effectstest.LatestChooseCardsFor(g, me.ID)
	if second == nil {
		t.Fatalf("no prompt for the second graveyard")
	}
	for _, id := range mineChosen {
		if effectstest.HasID(second.ChooseCards, id) {
			t.Error("the second prompt offers a card of the first graveyard")
		}
	}
	for _, id := range mineChosen {
		if !me.Graveyard.Contains(id) {
			t.Error("a card moved before every graveyard had been chosen from")
		}
	}
	theirChosen := pickFrom(second)
	effectstest.AnswerChooseCards(t, g, me.ID, theirChosen...)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Graveyard.Cards) != myGy-3+1 || len(me.Library.Cards) != myLib+3 {
		t.Errorf("my graveyard %d -> %d, library %d -> %d, want -3 (and the sacrificed Saga +1) and +3", myGy, len(me.Graveyard.Cards), myLib, len(me.Library.Cards))
	}
	if len(opp.Library.Cards) != theirLib+3 || len(opp.Graveyard.Cards) != 2 {
		t.Errorf("their library %d -> %d, graveyard %d, want +3 and 2 left", theirLib, len(opp.Library.Cards), len(opp.Graveyard.Cards))
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed")
	}
}

func TestOldB060WakingTheTrollsDestroysReanimatesAndMakesTrollsForTheLandGap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	theirLand := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Mountain", TypeLine: "Basic Land — Mountain"})
	saga := oldB060Saga(t, g, me, "Waking the Trolls", oldB060TrollsOracle, 0)
	oldB060NextChapter(g, saga)
	effectstest.PickCard(t, g, me.ID, theirLand)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirLand) {
		t.Error("chapter I did not destroy the land")
	}

	dead := effectstest.PushGraveyardCardTyped(opp, "Dead Swamp", "Basic Land — Swamp")
	nonland := effectstest.PushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	oldB060NextChapter(g, saga)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, nonland) || !effectstest.HasID(p.PickTargetCards, dead) {
		t.Fatalf("chapter II's target prompt is wrong: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, dead) != me.ID {
		t.Error("the land did not enter under my control")
	}

	// I now control 2 lands, the opponent 1: one Troll.
	oldB060NextChapter(g, saga)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	trolls := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Troll Warrior" {
			trolls++
			if c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(&c, "trample") || c.Controller != me.ID {
				t.Errorf("Troll Warrior token = %+v", c)
			}
		}
	}
	if trolls != 1 {
		t.Errorf("%d Trolls, want the difference of 1", trolls)
	}
}

func TestOldB060WakingTheTrollsMakesNoTrollsWhenTheOpponentHasAsManyLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Mountain", TypeLine: "Basic Land — Mountain"})
	saga := oldB060Saga(t, g, me, "Waking the Trolls", oldB060TrollsOracle, 2)
	oldB060NextChapter(g, saga)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Troll Warrior" {
			t.Error("a Troll was made with no land deficit")
		}
	}
}

func TestOldB060HaraldUnitesTheElvesMillsReanimatesCountersAndShrinksAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Creature — Bear")
	}
	elfInYard := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf Warrior")
	tyvar := effectstest.PushGraveyardCardTyped(me, "Tyvar Kell", "Legendary Planeswalker — Tyvar")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	saga := oldB060Saga(t, g, me, "Harald Unites the Elves", oldB060HaraldOracle, 0)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("chapter I asked nothing")
	}
	if !effectstest.HasID(c.ChooseCards, elfInYard) || !effectstest.HasID(c.ChooseCards, tyvar) || effectstest.HasID(c.ChooseCards, bear) {
		t.Errorf("offered %v, want the Elf and Tyvar cards only", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elfInYard)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elfInYard) {
		t.Error("the Elf did not come back")
	}

	elf := effectstest.B12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, elf, game.CounterPlusOne); n != 1 {
		t.Errorf("the Elf has %d counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, elfInYard, game.CounterPlusOne); n != 1 {
		t.Errorf("the returned Elf has %d counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, other, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Elf got %d counters", n)
	}
}

func TestOldB060HaraldUnitesTheElvesChapterThreeShrinksPerAttackingElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	saga := oldB060Saga(t, g, me, "Harald Unites the Elves", oldB060HaraldOracle, 2)
	elf := effectstest.B12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{elf, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.SncSettle(t, g, me.ID, victim)
	if p, tough := effectstest.PtOf(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("Victim is %d/%d, want 2/2 after one Elf attacked (the Bear is not an Elf)", p, tough)
	}
}

func TestOldB060NikoDefiesDestinyGainsForForetoldCardsAndReturnsAForetellCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	saga := oldB060Saga(t, g, me, "Niko Defies Destiny", oldB060NikoOracle, 0)
	card := effectstest.HandCardForTest(me, "Augury Raven", "Creature — Bird", oldB060AuguryRavenOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	life := me.Life
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("chapter I: life %d -> %d, want +2 for one foretold card", life, me.Life)
	}

	with := effectstest.PushGraveyardCardTyped(me, "Augury Raven", "Creature — Bird")
	without := effectstest.PushGraveyardCardTyped(me, "Plain Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == with {
				me.Graveyard.Cards[i].OracleID = oldB060AuguryRavenOracle
			}
		}
	})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("W"); n != 1 {
		t.Errorf("chapter II added %d {W}", n)
	}
	if n := me.ManaPool.CountColor("U"); n != 1 {
		t.Errorf("chapter II added %d {U}", n)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	oldB060NextChapter(g, saga)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || !effectstest.HasID(p.PickTargetCards, with) || effectstest.HasID(p.PickTargetCards, without) {
		t.Fatalf("chapter III's target prompt is wrong: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, with)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, with) {
		t.Error("the foretell card did not return")
	}
}

const oldB060AuguryRavenOracle = "ee05fc58-1356-4b2f-b9ce-bda8cc716aee"

func TestOldB060NikoDefiesDestinyManaPaysForForetellAndForetellSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	saga := oldB060Saga(t, g, me, "Niko Defies Destiny", oldB060NikoOracle, 1)
	raven := effectstest.HandCardForTest(me, "Augury Raven", "Creature — Bird", oldB060AuguryRavenOracle)
	plain := effectstest.HandCardForTest(me, "Plain Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == plain {
				me.Hand.Cards[i].ManaCost = "{1}{W}"
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
			}
		}
	})
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana paid for a spell without foretell")
	}
	if err := g.PerformSpecialAction(me.ID, raven, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("the restricted mana did not pay for foretelling: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left after the foretell", n)
	}
	// The same mana also pays for casting the foretold card ({1}{U}).
	g.WithWriteLock(func() { g.Turn.Seq++ })
	effectstest.AdvanceToMain(t, g)
	restriction := game.ManaRestrictAnyOf([]string{game.ManaRestrictForetell}, []string{game.ManaRestrictSpellForetell})
	me.ManaPool.AddMana(
		game.ManaToken{Color: "W", Restrictions: []string{restriction}},
		game.ManaToken{Color: "U", Restrictions: []string{restriction}},
	)
	if err := g.CastSpell(me.ID, raven, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err != nil {
		t.Fatalf("the restricted mana did not pay for casting a spell with foretell: %v", err)
	}
}

func TestOldB060TheBloodskyMassacreMakesABerserkerDrawsOnAttacksAndAddsKeptMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB060Seats(g)
	saga := oldB060Saga(t, g, me, "The Bloodsky Massacre", oldB060BloodskyOracle, 0)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Demon Berserker" {
			token = c.InstanceID
			if c.Power != 2 || c.Toughness != 3 || !game.HasKeyword(&c, "menace") || !c.HasColor("R") {
				t.Errorf("token = %+v", c)
			}
		}
	}
	if token == uuid.Nil {
		t.Fatal("chapter I made no Demon Berserker")
	}
	berserkerA := effectstest.B12Creature(g, me.ID, "Berserker A", "Creature — Human Berserker", 2, 2)
	berserkerB := effectstest.B12Creature(g, me.ID, "Berserker B", "Creature — Human Berserker", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Creature — Bear")
	}
	hand, life := len(me.Hand.Cards), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{berserkerA, berserkerB, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.SncSettle(t, g, me.ID, uuid.Nil)
	if len(me.Hand.Cards) != hand+2 || me.Life != life-2 {
		t.Errorf("two Berserkers attacked: hand %d -> %d, life %d -> %d, want +2 cards and -2 life", hand, len(me.Hand.Cards), life, me.Life)
	}
}

func TestOldB060TheBloodskyMassacreChapterThreeCountsBerserkersAndKeepsTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB060Seats(g)
	effectstest.B12Creature(g, me.ID, "Berserker A", "Creature — Human Berserker", 2, 2)
	effectstest.B12Creature(g, me.ID, "Berserker B", "Creature — Human Berserker", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	saga := oldB060Saga(t, g, me, "The Bloodsky Massacre", oldB060BloodskyOracle, 2)
	oldB060NextChapter(g, saga)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("R"); n != 2 {
		t.Fatalf("chapter III added %d {R}, want 2", n)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if n := me.ManaPool.CountColor("R"); n != 2 {
		t.Errorf("the mana emptied at the end of the phase: %d {R} left, want 2", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed")
	}
}
