package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestUndyingPersistPool1Registered(t *testing.T) {
	want := map[string]string{
		effectstest.PoolKitchenFinks: "persist", effectstest.PoolGeralfsMessenger: "undying", effectstest.PoolGeralfsMindcrush: "undying",
		effectstest.PoolEvernightShade: "undying", effectstest.PoolRestlessApparitn: "persist", effectstest.PoolGlenElendra: "persist",
		effectstest.PoolKithkinSpellduster: "persist", effectstest.PoolAerieOuphes: "persist", effectstest.PoolGrazingKelpie: "persist",
		effectstest.PoolLesserMasticore: "persist", effectstest.PoolSightlessGhoul: "undying", effectstest.PoolStormboundGeist: "undying",
		effectstest.PoolHowlgeist: "undying", effectstest.PoolThunderblust: "persist", effectstest.PoolWoodfallPrimus: "persist",
		effectstest.PoolPyreheartWolf: "undying",
	}
	for oracle, kw := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		has := false
		for _, k := range spec.PrintedKeywords {
			has = has || k == kw
		}
		if !has {
			t.Errorf("%s does not print %s: %v", spec.Name, kw, spec.PrintedKeywords)
		}
	}
}

// Kitchen Finks gains 2 as it enters, and again when persist returns it.
func TestKitchenFinksGainsLifeOnEachEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	finks := effectstest.PoolPush(g, me.ID, "Kitchen Finks", effectstest.PoolKitchenFinks, 3, 2)
	life := me.Life
	effectstest.B25Destroy(g, finks)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, finks) || effectstest.CountersOn(g, finks, game.CounterMinusOne) != 1 {
		t.Fatal("persist did not return the Finks with a -1/-1 counter")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 from the second entry", life, me.Life)
	}
}

// Geralf's Messenger returns tapped, with its counter, and drains again.
func TestGeralfsMessengerReturnsTappedAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msg := effectstest.PoolPush(g, me.ID, "Geralf's Messenger", effectstest.PoolGeralfsMessenger, 3, 2)
	life := opp.Life
	effectstest.B25Destroy(g, msg)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.SettleProwess(t, g)
	if !effectstest.B16Tapped(t, g, msg) || effectstest.CountersOn(g, msg, game.CounterPlusOne) != 1 {
		t.Error("undying should return it tapped with a +1/+1 counter")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, opp.Life)
	}
}

// Geralf's Mindcrusher mills five on each entry.
func TestGeralfsMindcrusherMillsFiveOnReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mc := effectstest.PoolPush(g, me.ID, "Geralf's Mindcrusher", effectstest.PoolGeralfsMindcrush, 5, 5)
	yard := opp.Graveyard.Size()
	effectstest.B25Destroy(g, mc)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.SettleProwess(t, g)
	if got := opp.Graveyard.Size() - yard; got != 5 {
		t.Errorf("milled %d, want 5", got)
	}
}

// Evernight Shade and Restless Apparition pump themselves.
func TestEvernightShadeAndRestlessApparitionPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shade := effectstest.PoolPush(g, me.ID, "Evernight Shade", effectstest.PoolEvernightShade, 1, 1)
	app := effectstest.PoolPush(g, me.ID, "Restless Apparition", effectstest.PoolRestlessApparitn, 2, 2)
	effectstest.PoolActivate(t, g, me, shade, "B")
	effectstest.SettleProwess(t, g)
	effectstest.PoolActivate(t, g, me, app, "WBW")
	effectstest.SettleProwess(t, g)
	if p := effectstest.PowerOf(t, g, shade); p != 2 {
		t.Errorf("Shade power %d, want 2", p)
	}
	if p := effectstest.PowerOf(t, g, app); p != 5 {
		t.Errorf("Apparition power %d, want 5", p)
	}
}

// Glen Elendra Archmage counters a noncreature spell and persists.
func TestGlenElendraArchmageCountersAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.PoolPush(g, me.ID, "Glen Elendra Archmage", effectstest.PoolGlenElendra, 2, 2)
	spell := effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PoolActivate(t, g, me, mage, "U", effectstest.CardRef(spell))
	effectstest.SettleProwess(t, g)
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the spell was not countered into the graveyard")
	}
	if !effects.OnBattlefield(g, mage) || effectstest.CountersOn(g, mage, game.CounterMinusOne) != 1 {
		t.Error("persist did not return the Archmage")
	}
}

// Kithkin Spellduster destroys an enchantment.
func TestKithkinSpelldusterDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duster := effectstest.PoolPush(g, me.ID, "Kithkin Spellduster", effectstest.PoolKithkinSpellduster, 2, 3)
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Enchantment", TypeLine: "Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PoolActivate(t, g, me, duster, "WW", effectstest.CardRef(ench))
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, ench) {
		t.Error("the enchantment survived")
	}
	if !effects.OnBattlefield(g, duster) {
		t.Error("persist did not return the Spellduster")
	}
}

// Aerie Ouphes deals its last-known power, 3 and then 2.
func TestAerieOuphesShootsFlyersWithItsLastKnownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ouphes := effectstest.PoolPush(g, me.ID, "Aerie Ouphes", effectstest.PoolAerieOuphes, 3, 3)
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 9, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	ground := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PoolActivate(t, g, me, ouphes, "", effectstest.CardRef(big))
	effectstest.SettleProwess(t, g)
	if c, _ := effectstest.BattlefieldCard(g, big); c.DamageMarked != 3 {
		t.Fatalf("first shot: %d damage, want 3", c.DamageMarked)
	}
	effectstest.FloatForTest(g, me, "")
	if err := g.ActivateCatalogAbility(me.ID, ouphes, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.CardRef(ground)}}); err == nil {
		t.Error("a creature without flying was accepted as a target")
	}
	effectstest.PoolActivate(t, g, me, ouphes, "", effectstest.CardRef(big))
	effectstest.SettleProwess(t, g)
	if c, _ := effectstest.BattlefieldCard(g, big); c.DamageMarked != 5 {
		t.Errorf("second shot: %d total damage, want 5 (3 + 2)", c.DamageMarked)
	}
	if effects.OnBattlefield(g, ouphes) {
		t.Error("it died with a -1/-1 counter the second time and should stay dead")
	}
}

// Grazing Kelpie puts a graveyard card on the bottom of its owner's library.
func TestGrazingKelpieTucksAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kelpie := effectstest.PoolPush(g, me.ID, "Grazing Kelpie", effectstest.PoolGrazingKelpie, 2, 3)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Thing", TypeLine: "Creature — Zombie", Owner: opp.ID, Controller: opp.ID})
	effectstest.PoolActivate(t, g, me, kelpie, "G", effectstest.CardRef(dead))
	effectstest.SettleProwess(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if lib := opp.Library.Cards; len(lib) == 0 || lib[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
}

// Lesser Masticore pings a creature for {4}.
func TestLesserMasticorePings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mc := effectstest.PoolPush(g, me.ID, "Lesser Masticore", effectstest.PoolLesserMasticore, 2, 2)
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PoolActivate(t, g, me, mc, "CCCC", effectstest.CardRef(bear))
	effectstest.SettleProwess(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.DamageMarked != 1 {
		t.Errorf("%d damage, want 1", c.DamageMarked)
	}
	if spec, _ := effects.Lookup(effectstest.PoolLesserMasticore); spec.AdditionalCost == nil || spec.AdditionalCost.DiscardCards != 1 {
		t.Error("the discard additional cost is missing")
	}
}

// Sightless Ghoul can't block; Stormbound Geist blocks only flyers;
// Howlgeist can't be blocked by smaller creatures.
func TestUndyingBlockRules(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.PoolPush(g, me.ID, "Sightless Ghoul", effectstest.PoolSightlessGhoul, 2, 2)
	geist := effectstest.PoolPush(g, me.ID, "Stormbound Geist", effectstest.PoolStormboundGeist, 2, 2)
	howl := effectstest.PoolPush(g, opp.ID, "Howlgeist", effectstest.PoolHowlgeist, 4, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	bigMine := effectstest.B16Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		get := func(id uuid.UUID) game.Card { c, _ := g.LookupCardForEffect(id); return c }
		gh, ge, hg, fl, og := get(ghoul), get(geist), get(howl), get(flyer), get(bigMine)
		if gh.Effective().Restrictions&game.CantBlock == 0 {
			t.Error("Sightless Ghoul can block")
		}
		if r := g.BlockPairRefusalLocked(&fl, &ge); !r.Legal() {
			t.Errorf("Stormbound Geist may not block a flyer: %q", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(&hg, &ge); r.Legal() {
			t.Error("Stormbound Geist blocked a creature without flying")
		}
		if r := g.BlockPairRefusalLocked(&hg, &og); !r.Legal() {
			t.Errorf("a 4-power creature may not block a 4-power Howlgeist: %q", r.Reason)
		}
	})
}

// Thunderblust has trample only with a -1/-1 counter.
func TestThunderblustTramplesAfterPersist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tb := effectstest.PoolPush(g, me.ID, "Thunderblust", effectstest.PoolThunderblust, 7, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := effectstest.BattlefieldCard(g, tb); game.HasKeyword(&c, "trample") {
		t.Fatal("trample with no -1/-1 counter")
	}
	effectstest.B25Destroy(g, tb)
	effectstest.SettleProwess(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := effectstest.BattlefieldCard(g, tb); !game.HasKeyword(&c, "trample") {
		t.Error("no trample after persist returned it with a -1/-1 counter")
	}
}

// Woodfall Primus destroys a noncreature permanent on entry.
func TestWoodfallPrimusDestroysOnReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	primus := effectstest.PoolPush(g, me.ID, "Woodfall Primus", effectstest.PoolWoodfallPrimus, 6, 6)
	art := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Artifact", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.B25Destroy(g, primus)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, art)
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, art) {
		t.Error("the artifact survived the returned Primus")
	}
}

// Pyreheart Wolf gives the team menace when it attacks.
func TestPyreheartWolfGivesMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.PoolPush(g, me.ID, "Pyreheart Wolf", effectstest.PoolPyreheartWolf, 1, 1)
	other := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, wolf)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := effectstest.BattlefieldCard(g, other); !game.HasKeyword(&c, "menace") {
		t.Error("another creature I control did not gain menace")
	}
}
