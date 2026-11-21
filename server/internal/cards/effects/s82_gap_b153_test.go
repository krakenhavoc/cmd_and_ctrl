package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b153_test.go — GAP batch 153 (Jumpstart 2022 green creatures,
// #5948), played through the catalog and the real stack.

const (
	gapB153Runadi         = "95dc1f55-9dad-4a35-ac5e-abc53d442ffb"
	gapB153Bandar         = "8e1eff5a-77e8-4ea2-96b8-547504f83b45"
	gapB153Brawler        = "2497cfe2-a85c-47ea-af7a-9a745d84f8b4"
	gapB153HuntCaller     = "d03ef35b-0609-46b5-ab8a-64d8007eab26"
	gapB153Shambler       = "f46c72b8-e2e8-42e6-a2e7-ffa5810a9bf3"
	gapB153Gibbon         = "0684771f-46ea-40a3-b021-e98a7d7cd79d"
	gapB153WickedWolf     = "19cede61-1982-47d1-9e0e-49ac6de00798"
	gapB153LightningBolt  = "4457ed35-7c10-48c8-9776-456485fdf070"
	gapB153CreatureTypeLn = "Creature — Bear"
)

// gapB153Cast puts a creature card with `cost` in p's hand, floats `mana` and
// casts it, returning its ID.
func gapB153Cast(t *testing.T, g *game.Game, p *game.Player, cost, mana string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Beast", TypeLine: gapB153CreatureTypeLn, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	floatForTest(g, p, mana)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the Beast: %v", err)
	}
	return id
}

func TestGapB153RunadiGivesBigCreaturesCountersByManaValueMinusFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	runadi := b12Push(g, me.ID, "Runadi, Behemoth Caller", "Legendary Creature — Cat Shaman", gapB153Runadi, 1, 3)
	six := gapB153Cast(t, g, me, "{4}{G}{G}", "GGGGGG")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, six, game.CounterPlusOne); got != 2 {
		t.Errorf("a mana value 6 creature entered with %d counters, want 2", got)
	}
	five := gapB153Cast(t, g, me, "{4}{G}", "GGGGG")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, five, game.CounterPlusOne); got != 1 {
		t.Errorf("a mana value 5 creature entered with %d counters, want 1", got)
	}
	four := gapB153Cast(t, g, me, "{3}{G}", "GGGG")
	if tr := triggerOnStack(g, runadi); tr != nil {
		t.Error("a mana value 4 creature spell triggered Runadi")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, four, game.CounterPlusOne); got != 0 {
		t.Errorf("a mana value 4 creature entered with %d counters, want 0", got)
	}
}

func TestGapB153RunadiGrantsHasteToCreaturesWithThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Runadi, Behemoth Caller", "Legendary Creature — Cat Shaman", gapB153Runadi, 1, 3)
	two := b12Creature(g, me.ID, "Two", gapB153CreatureTypeLn, 2, 2)
	three := b12Creature(g, me.ID, "Three", gapB153CreatureTypeLn, 2, 2)
	putCounters(t, g, two, game.CounterPlusOne, 2)
	putCounters(t, g, three, game.CounterPlusOne, 3)
	if p1g2HasKeyword(g, two, "haste") {
		t.Error("a creature with two counters has haste")
	}
	if !p1g2HasKeyword(g, three, "haste") {
		t.Error("a creature with three counters lacks haste")
	}
}

func TestGapB153RunadiTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	runadi := b12Push(g, me.ID, "Runadi, Behemoth Caller", "Legendary Creature — Cat Shaman", gapB153Runadi, 1, 3)
	if err := g.ActivateManaAbility(me.ID, runadi, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Runadi for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("the pool holds %d, want 1", got)
	}
}

func TestGapB153BandarEntersWithTwoCountersAndMovesThemOnUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bandar := mb1Enter(t, g, me, "Scrounging Bandar", "Creature — Cat Monkey", gapB153Bandar, 0, 0)
	if got := counterCount(g, bandar, game.CounterPlusOne); got != 2 {
		t.Fatalf("the Bandar entered with %d counters, want 2", got)
	}
	bear := b12Creature(g, me.ID, "Bear", gapB153CreatureTypeLn, 2, 2)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
	}
	rprB23AnswerNumber(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	if got := counterCount(g, bandar, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bandar has %d counters, want 1", got)
	}
}

func TestGapB153BrawlerDiscardsOnlyALandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	brawler := b12Push(g, me.ID, "Simian Brawler", "Creature — Ape Warrior", gapB153Brawler, 3, 3)
	spell, land := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("Simian Brawler discarded a nonland card")
	}
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("discard a land card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if live := apaLive(g, brawler); live.CurrentPower() != 4 || live.CurrentToughness() != 4 {
		t.Errorf("the Brawler is %d/%d, want 4/4", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB153HuntCallerPumpsAndTramplesTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	caller := b12Push(g, me.ID, "Spectral Hunt-Caller", "Creature — Wolf Spirit", gapB153HuntCaller, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", gapB153CreatureTypeLn, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", gapB153CreatureTypeLn, 2, 2)
	floatForTest(g, me, "GGGGGG")
	apaActivate(t, g, me, caller, 0, game.ActivateAbilityParams{})
	if live := apaLive(g, bear); live.CurrentPower() != 3 || live.CurrentToughness() != 3 || !p1g2HasKeyword(g, bear, "trample") {
		t.Errorf("my Bear is %d/%d (trample %v), want 3/3 with trample", live.CurrentPower(), live.CurrentToughness(), p1g2HasKeyword(g, bear, "trample"))
	}
	if live := apaLive(g, theirs); live.CurrentPower() != 2 || p1g2HasKeyword(g, theirs, "trample") {
		t.Errorf("an opponent's Bear was pumped")
	}
}

func TestGapB153ShamblerMakesAnInsectWhenACounteredCreatureIsTargetedByASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	advanceToMain(t, g)
	shambler := mb1Enter(t, g, me, "Swarm Shambler", "Creature — Fungus Beast", gapB153Shambler, 0, 0)
	if got := counterCount(g, shambler, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Shambler entered with %d counters, want 1", got)
	}
	plain := b12Creature(g, me.ID, "Plain", gapB153CreatureTypeLn, 3, 3)
	insects := func() int { n, _ := countTokensNamed(g, me.ID, "Insect"); return n }
	// A creature with no counter: nothing.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB153LightningBolt, cardTarget(plain))
	passPriorityAroundTable(t, g)
	if insects() != 0 {
		t.Fatalf("an Insect was made for a creature with no counter")
	}
	// The Shambler itself, targeted by an opponent's spell.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB153LightningBolt, cardTarget(shambler))
	passPriorityAroundTable(t, g)
	if insects() != 1 {
		t.Errorf("%d Insects after the Shambler was targeted, want 1", insects())
	}
}

func TestGapB153ShamblerIgnoresYourOwnSpellsAndAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shambler := mb1Enter(t, g, me, "Swarm Shambler", "Creature — Fungus Beast", gapB153Shambler, 0, 0)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB153LightningBolt, cardTarget(shambler))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Insect"); n != 0 {
		t.Errorf("%d Insects for its controller's own spell, want 0", n)
	}
}

func TestGapB153ShamblerGrowsForOneAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shambler := b12Push(g, me.ID, "Swarm Shambler", "Creature — Fungus Beast", gapB153Shambler, 0, 0)
	floatForTest(g, me, "G")
	apaActivate(t, g, me, shambler, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, shambler, game.CounterPlusOne); got != 1 {
		t.Errorf("the Shambler has %d counters, want 1", got)
	}
}

func TestGapB153GibbonPowerIsTheGreatestManaValueAmongYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gibbon := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Towering Gibbon", TypeLine: "Creature — Ape",
		OracleID: gapB153Gibbon, ManaCost: "{3}{G}", Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if live := apaLive(g, gibbon); live.CurrentPower() != 4 {
		t.Errorf("alone, the Gibbon's power is %d, want its own mana value 4", live.CurrentPower())
	}
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: gapB153CreatureTypeLn,
		ManaCost: "{6}{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: gapB153CreatureTypeLn,
		ManaCost: "{9}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if live := apaLive(g, gibbon); live.CurrentPower() != 7 || live.CurrentToughness() != 4 {
		t.Errorf("the Gibbon is %d/%d, want 7/4 (an opponent's creature does not count)", live.CurrentPower(), live.CurrentToughness())
	}
	mb1Kill(g, big)
	if live := apaLive(g, gibbon); live.CurrentPower() != 4 {
		t.Errorf("after the big creature left, the Gibbon's power is %d, want 4", live.CurrentPower())
	}
}

func TestGapB153WickedWolfFightsOnEntryAndSacrificesFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Victim", gapB153CreatureTypeLn, 2, 2)
	wolf := mb1Enter(t, g, me, "Wicked Wolf", "Creature — Wolf", gapB153WickedWolf, 3, 3)
	mb1Pick(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the fought creature survived 3 damage")
	}
	if live := apaLive(g, wolf); live.DamageMarked != 2 {
		t.Errorf("the Wolf took %d damage, want 2", live.DamageMarked)
	}
	// No Food: the ability cannot be paid.
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no Food to sacrifice")
	}
	food := b12Permanent(g, me.ID, "Food", "Artifact — Food")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("sacrifice a Food: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if got := counterCount(g, wolf, game.CounterPlusOne); got != 1 {
		t.Errorf("the Wolf has %d counters, want 1", got)
	}
	if !p1g2HasKeyword(g, wolf, "indestructible") {
		t.Error("the Wolf is not indestructible")
	}
	if live := apaLive(g, wolf); !live.Tapped {
		t.Error("the Wolf is not tapped")
	}
}
