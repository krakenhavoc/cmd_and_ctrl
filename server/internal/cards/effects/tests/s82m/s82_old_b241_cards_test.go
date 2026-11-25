package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b241_cards_test.go — Hour of Devastation, batch 241 (S82 older
// sets, #5534). Every test plays the real catalog entry through the engine.

const (
	oldB241GildedCerodon = "93dd5f2e-aa13-4616-bfe6-466524b274ff"
	oldB241AvidReclaimer = "a94bb381-74e2-488f-896d-a017a2b568ab"
	oldB241Crypt         = "cc78776b-822b-4f11-8982-0805a25a9d36"
	oldB241Solemnity     = "f0b4331b-ddd3-4f11-ba16-67c651c8954c"
)

func TestOldB241EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB241GildedCerodon, oldB241AvidReclaimer, oldB241Crypt, oldB241Solemnity} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB241SolemnityStopsCountersOnPlayersAndFourTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.B12Permanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Ajani")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Fatalf("before Solemnity the bear has %d counters, want 1", n)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Solemnity", "Enchantment", oldB241Solemnity, false)
	g.RunStateChecksForTest()
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(walker, game.CounterLoyalty, 3)
		_ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 1)
		_ = g.AddPlayerCounterForEffect(opp.ID, game.CounterEnergy, 2)
	})
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want the 1 it had", n)
	}
	if opp.Counters[game.CounterPoison] != 0 || opp.Counters[game.CounterEnergy] != 0 {
		t.Errorf("a player got counters: %v", opp.Counters)
	}
	var loyalty int
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(walker); ok {
			loyalty = c.Counters[game.CounterLoyalty]
		}
	})
	if loyalty != 3 {
		t.Errorf("the planeswalker has %d loyalty, want 3: Solemnity doesn't name planeswalkers", loyalty)
	}
}

func TestOldB241GildedCerodonStopsABlockerOnlyWithADesert(t *testing.T) {
	for _, tc := range []struct {
		name   string
		desert func(g *game.Game, me *game.Player)
		want   bool
	}{
		{"no desert", func(*game.Game, *game.Player) {}, false},
		{"desert on the battlefield", func(g *game.Game, me *game.Player) {
			effectstest.B12Permanent(g, me.ID, "Test Desert", "Land — Desert")
		}, true},
		{"desert in the graveyard", func(_ *game.Game, me *game.Player) {
			c := game.NewCard("Test Desert", me.ID)
			c.TypeLine = "Land — Desert"
			me.Graveyard.PushTop(c)
		}, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		cerodon := effectstest.B12Push(g, me.ID, "Gilded Cerodon", "Creature — Beast", oldB241GildedCerodon, 4, 4)
		foe := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		tc.desert(g, me)
		if tc.want {
			effectstest.AttackWithTrigger(t, g, cerodon, opp.ID, foe)
		} else {
			effectstest.DeclareAttack(t, g, opp.ID, cerodon)
			if n := effectstest.TriggersOnStackFrom(g, cerodon); n != 0 {
				t.Errorf("%s: the trigger went on the stack", tc.name)
			}
		}
		if got := effectstest.MidB03CantBlock(t, g, foe); got != tc.want {
			t.Errorf("%s: target can't block = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestOldB241AvidReclaimerGainsTwoOnlyWithANissa(t *testing.T) {
	for _, tc := range []struct {
		subtype string
		want    int
	}{{"", 0}, {"Jace", 0}, {"Nissa", 2}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SeatsForCombat(g)
		druid := effectstest.B12Push(g, me.ID, "Avid Reclaimer", "Creature — Human Druid", oldB241AvidReclaimer, 2, 2)
		if tc.subtype != "" {
			effectstest.B12Permanent(g, me.ID, tc.subtype+" Test", "Legendary Planeswalker — "+tc.subtype)
		}
		before := effectstest.LifeOf(g, me.ID)
		if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
			t.Fatalf("%q: ActivateManaAbility: %v", tc.subtype, err)
		}
		if got := effectstest.LifeOf(g, me.ID) - before; got != tc.want {
			t.Errorf("%q: life changed by %d, want %d", tc.subtype, got, tc.want)
		}
		if me.ManaPool.Total() != 1 {
			t.Errorf("%q: pool holds %d mana, want 1", tc.subtype, me.ManaPool.Total())
		}
	}
}

func TestOldB241CryptOfTheEternalsGainsLifeAndFiltersMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	before := effectstest.LifeOf(g, me.ID)
	land := effectstest.PlayLandFromHand(t, g, "Crypt of the Eternals", oldB241Crypt)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID) - before; got != 1 {
		t.Errorf("entering gained %d life, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool holds %d, want 1", me.ManaPool.Total())
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	land = effectstest.PushCatalogPermanent(g, me.ID, "Crypt of the Eternals", "Land", oldB241Crypt, false)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("the filter ability worked with no mana to pay {1}")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(game.ManaToken{Color: "G"}) })
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1}, {T}: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.CountColor("G") != 0 {
		t.Errorf("the filter left %v, want one blue, black or red mana", me.ManaPool.Units())
	}
}
