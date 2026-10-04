package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_pool1_test.go — the first pool PR after #2075 (ADR
// 0113 §4, owner decision 2): creatures whose undying or persist the
// engine already handles, with the rest of their text. One test per
// card; the keyword itself is held by undying_persist_test.go.

const (
	poolKitchenFinks       = "5470dcfa-4eff-43da-abf7-19922841f719"
	poolGeralfsMessenger   = "740f9740-aaa7-4061-86f5-85be79742541"
	poolGeralfsMindcrush   = "e8fa9455-92ac-46e0-bb34-4175c3c66fee"
	poolEvernightShade     = "9af1b6c4-295f-41d2-b3a0-0863798330d4"
	poolRestlessApparitn   = "fbd6dda7-ccee-4640-805d-78a9d60f45ef"
	poolGlenElendra        = "5f3f68b5-8c6a-4181-bb8a-9c73967d198a"
	poolKithkinSpellduster = "de8beb92-234b-4684-9185-c7077c8a5133"
	poolAerieOuphes        = "70cdd24a-8dc5-47b5-8729-9ebf486f4821"
	poolGrazingKelpie      = "a2db1035-008e-4de3-b5e5-2a0de83082b2"
	poolLesserMasticore    = "a6a257bd-ee19-4246-b997-5288e6fa41e1"
	poolSightlessGhoul     = "8791adf7-16b5-4692-9eaa-0d0c58867d9d"
	poolStormboundGeist    = "b5027d79-83e9-437a-baae-03cc3a13a608"
	poolHowlgeist          = "5a00eb94-6c0f-4f41-8c16-fc4e2f7cb9aa"
	poolThunderblust       = "30d961f7-aae2-4e5a-8f83-9dc17c7dee47"
	poolWoodfallPrimus     = "2f70f1bb-29fa-4abb-afc2-653acd0a08b9"
	poolPyreheartWolf      = "b97ccb69-e76c-4962-91ca-c7fd857140e8"
)

func TestUndyingPersistPool1Registered(t *testing.T) {
	want := map[string]string{
		poolKitchenFinks: "persist", poolGeralfsMessenger: "undying", poolGeralfsMindcrush: "undying",
		poolEvernightShade: "undying", poolRestlessApparitn: "persist", poolGlenElendra: "persist",
		poolKithkinSpellduster: "persist", poolAerieOuphes: "persist", poolGrazingKelpie: "persist",
		poolLesserMasticore: "persist", poolSightlessGhoul: "undying", poolStormboundGeist: "undying",
		poolHowlgeist: "undying", poolThunderblust: "persist", poolWoodfallPrimus: "persist",
		poolPyreheartWolf: "undying",
	}
	for oracle, kw := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
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

// poolPush puts a catalog creature on the battlefield with no entry.
func poolPush(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return b11Push(g, owner, name, "Creature — Test", oracle, power, toughness)
}

func poolActivate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) {
	t.Helper()
	floatForTest(g, p, mana)
	if err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

func cardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func powerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	p := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		p = c.PowerForComparison()
	})
	return p
}

// Kitchen Finks gains 2 as it enters, and again when persist returns it.
func TestKitchenFinksGainsLifeOnEachEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finks := poolPush(g, me.ID, "Kitchen Finks", poolKitchenFinks, 3, 2)
	life := me.Life
	b25Destroy(g, finks)
	settleProwess(t, g)
	if !onBattlefield(g, finks) || countersOn(g, finks, game.CounterMinusOne) != 1 {
		t.Fatal("persist did not return the Finks with a -1/-1 counter")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 from the second entry", life, me.Life)
	}
}

// Geralf's Messenger returns tapped, with its counter, and drains again.
func TestGeralfsMessengerReturnsTappedAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msg := poolPush(g, me.ID, "Geralf's Messenger", poolGeralfsMessenger, 3, 2)
	life := opp.Life
	b25Destroy(g, msg)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	settleProwess(t, g)
	if !b16Tapped(t, g, msg) || countersOn(g, msg, game.CounterPlusOne) != 1 {
		t.Error("undying should return it tapped with a +1/+1 counter")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, opp.Life)
	}
}

// Geralf's Mindcrusher mills five on each entry.
func TestGeralfsMindcrusherMillsFiveOnReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mc := poolPush(g, me.ID, "Geralf's Mindcrusher", poolGeralfsMindcrush, 5, 5)
	yard := opp.Graveyard.Size()
	b25Destroy(g, mc)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	settleProwess(t, g)
	if got := opp.Graveyard.Size() - yard; got != 5 {
		t.Errorf("milled %d, want 5", got)
	}
}

// Evernight Shade and Restless Apparition pump themselves.
func TestEvernightShadeAndRestlessApparitionPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shade := poolPush(g, me.ID, "Evernight Shade", poolEvernightShade, 1, 1)
	app := poolPush(g, me.ID, "Restless Apparition", poolRestlessApparitn, 2, 2)
	poolActivate(t, g, me, shade, "B")
	settleProwess(t, g)
	poolActivate(t, g, me, app, "WBW")
	settleProwess(t, g)
	if p := powerOf(t, g, shade); p != 2 {
		t.Errorf("Shade power %d, want 2", p)
	}
	if p := powerOf(t, g, app); p != 5 {
		t.Errorf("Apparition power %d, want 5", p)
	}
}

// Glen Elendra Archmage counters a noncreature spell and persists.
func TestGlenElendraArchmageCountersAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := poolPush(g, me.ID, "Glen Elendra Archmage", poolGlenElendra, 2, 2)
	spell := castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	poolActivate(t, g, me, mage, "U", cardRef(spell))
	settleProwess(t, g)
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the spell was not countered into the graveyard")
	}
	if !onBattlefield(g, mage) || countersOn(g, mage, game.CounterMinusOne) != 1 {
		t.Error("persist did not return the Archmage")
	}
}

// Kithkin Spellduster destroys an enchantment.
func TestKithkinSpelldusterDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duster := poolPush(g, me.ID, "Kithkin Spellduster", poolKithkinSpellduster, 2, 3)
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Enchantment", TypeLine: "Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})
	poolActivate(t, g, me, duster, "WW", cardRef(ench))
	settleProwess(t, g)
	if onBattlefield(g, ench) {
		t.Error("the enchantment survived")
	}
	if !onBattlefield(g, duster) {
		t.Error("persist did not return the Spellduster")
	}
}

// Aerie Ouphes deals its last-known power, 3 and then 2.
func TestAerieOuphesShootsFlyersWithItsLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ouphes := poolPush(g, me.ID, "Aerie Ouphes", poolAerieOuphes, 3, 3)
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 9, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	ground := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	poolActivate(t, g, me, ouphes, "", cardRef(big))
	settleProwess(t, g)
	if c, _ := battlefieldCard(g, big); c.DamageMarked != 3 {
		t.Fatalf("first shot: %d damage, want 3", c.DamageMarked)
	}
	floatForTest(g, me, "")
	if err := g.ActivateCatalogAbility(me.ID, ouphes, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(ground)}}); err == nil {
		t.Error("a creature without flying was accepted as a target")
	}
	poolActivate(t, g, me, ouphes, "", cardRef(big))
	settleProwess(t, g)
	if c, _ := battlefieldCard(g, big); c.DamageMarked != 5 {
		t.Errorf("second shot: %d total damage, want 5 (3 + 2)", c.DamageMarked)
	}
	if onBattlefield(g, ouphes) {
		t.Error("it died with a -1/-1 counter the second time and should stay dead")
	}
}

// Grazing Kelpie puts a graveyard card on the bottom of its owner's library.
func TestGrazingKelpieTucksAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kelpie := poolPush(g, me.ID, "Grazing Kelpie", poolGrazingKelpie, 2, 3)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Thing", TypeLine: "Creature — Zombie", Owner: opp.ID, Controller: opp.ID})
	poolActivate(t, g, me, kelpie, "G", cardRef(dead))
	settleProwess(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if lib := opp.Library.Cards; len(lib) == 0 || lib[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
}

// Lesser Masticore pings a creature for {4}.
func TestLesserMasticorePings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mc := poolPush(g, me.ID, "Lesser Masticore", poolLesserMasticore, 2, 2)
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	poolActivate(t, g, me, mc, "CCCC", cardRef(bear))
	settleProwess(t, g)
	if c, _ := battlefieldCard(g, bear); c.DamageMarked != 1 {
		t.Errorf("%d damage, want 1", c.DamageMarked)
	}
	if spec, _ := Lookup(poolLesserMasticore); spec.AdditionalCost == nil || spec.AdditionalCost.DiscardCards != 1 {
		t.Error("the discard additional cost is missing")
	}
}

// Sightless Ghoul can't block; Stormbound Geist blocks only flyers;
// Howlgeist can't be blocked by smaller creatures.
func TestUndyingBlockRules(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := poolPush(g, me.ID, "Sightless Ghoul", poolSightlessGhoul, 2, 2)
	geist := poolPush(g, me.ID, "Stormbound Geist", poolStormboundGeist, 2, 2)
	howl := poolPush(g, opp.ID, "Howlgeist", poolHowlgeist, 4, 2)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	bigMine := b16Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	tb := poolPush(g, me.ID, "Thunderblust", poolThunderblust, 7, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, tb); game.HasKeyword(&c, "trample") {
		t.Fatal("trample with no -1/-1 counter")
	}
	b25Destroy(g, tb)
	settleProwess(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, tb); !game.HasKeyword(&c, "trample") {
		t.Error("no trample after persist returned it with a -1/-1 counter")
	}
}

// Woodfall Primus destroys a noncreature permanent on entry.
func TestWoodfallPrimusDestroysOnReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	primus := poolPush(g, me.ID, "Woodfall Primus", poolWoodfallPrimus, 6, 6)
	art := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Artifact", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	b25Destroy(g, primus)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, art)
	settleProwess(t, g)
	if onBattlefield(g, art) {
		t.Error("the artifact survived the returned Primus")
	}
}

// Pyreheart Wolf gives the team menace when it attacks.
func TestPyreheartWolfGivesMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := poolPush(g, me.ID, "Pyreheart Wolf", poolPyreheartWolf, 1, 1)
	other := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, wolf)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, other); !game.HasKeyword(&c, "menace") {
		t.Error("another creature I control did not gain menace")
	}
}
