package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 2 — the utility lands. The Triomes are Xander's Lounge's
// template and are covered by the registry-wide checks; these tests
// drive the lands with an ability of their own.

const (
	cabalStrongholdOracle = "066cd584-773c-4623-be53-8f6feda5a26a"
	taintedPeakOracle     = "b2bae7fc-0668-4b34-9cd6-0d80aea52275"
	kessigWolfRunOracle   = "c6911265-54ef-4c16-bcf2-1ffb24b7d426"
	mirenOracle           = "03fe19bb-8e22-4030-8299-2ddd2d5a7eb2"
	nestingGroundsOracle  = "d27bb97d-286b-4947-8d7b-443e4df93319"
	castleGarenbrigOracle = "de75e5dd-8a52-406c-b55c-96d686885500"
	witchsClinicOracle    = "05899372-9784-4bdb-9c28-504c71fed906"
)

func TestCabalStrongholdCountsOnlyYourBasicSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stronghold := seedPermanentWithOracle(g, me.ID, "Cabal Stronghold", "Land", cabalStrongholdOracle)
	seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	seedManaLand(g, me.ID, "Watery Grave", "Land — Island Swamp", "U", "B")
	seedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	seedManaLand(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp", "B")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateManaAbility(me.ID, stronghold, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want two {B} (two basic Swamps, the {3} was spent)", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "B" {
			t.Errorf("token %v, want black", tok)
		}
	}
}

func TestTaintedPeakNeedsASwampForItsColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	peak := seedPermanentWithOracle(g, me.ID, "Tainted Peak", "Land", taintedPeakOracle)
	if err := g.ActivateManaAbility(me.ID, peak, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Tainted Peak made {B} or {R} with no Swamp")
	}
	seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	if err := g.ActivateManaAbility(me.ID, peak, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("with a Swamp: %v", err)
	}
	if pick := manaPickFor(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Errorf("mana pick = %+v, want a {B} or {R} choice", pick)
	}
}

func TestKessigWolfRunPumpsByXAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	run := seedPermanentWithOracle(g, me.ID, "Kessig Wolf Run", "Land", kessigWolfRunOracle)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := effectivePower(t, g, bear)
	if err := g.ActivateCatalogAbility(me.ID, run, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != before+2 {
		t.Errorf("power = %d, want %d", got, before+2)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestMirenGainsLifeEqualToTheSacrificedToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	miren := seedPermanentWithOracle(g, me.ID, "Miren, the Moaning Well", "Legendary Land", mirenOracle)
	ox := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ox", TypeLine: "Creature — Ox",
		Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, miren, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{ox},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestNestingGroundsMovesTheOnlyKindWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grounds := seedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", nestingGroundsOracle)
	from := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Source", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	to := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dest", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	src, _ := battlefieldCard(g, from)
	dst, _ := battlefieldCard(g, to)
	if src.Counters[game.CounterPlusOne] != 1 || dst.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters source=%v dest=%v, want one each", src.Counters, dst.Counters)
	}
}

func TestNestingGroundsAsksWhichKindWhenThereAreSeveral(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grounds := seedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", nestingGroundsOracle)
	from := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Source", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, "stun": 1},
	})
	to := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dest", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Kinds are offered in name order: "+1/+1" before "stun".
	answerOptionPick(t, g, me.ID, 1)
	src, _ := battlefieldCard(g, from)
	dst, _ := battlefieldCard(g, to)
	if src.Counters["stun"] != 0 || dst.Counters["stun"] != 1 || src.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters source=%v dest=%v, want the stun counter moved", src.Counters, dst.Counters)
	}
}

func TestNestingGroundsIsSorcerySpeedAndNeedsTwoDifferentPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grounds := seedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", nestingGroundsOracle)
	one := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "One", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: one}},
	}); err == nil {
		t.Fatal("the same permanent was accepted as both targets")
	}
}

func TestCastleGarenbrigMakesSixRestrictedGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castle := seedPermanentWithOracle(g, me.ID, "Castle Garenbrig", "Land", castleGarenbrigOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, castle, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 6 {
		t.Fatalf("pool = %v, want six {G}", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "G" || len(tok.Restrictions) != 1 {
			t.Errorf("token %+v, want restricted {G}", tok)
		}
	}
}

func TestWitchsClinicTargetsAnyPlayersCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	clinic := seedPermanentWithOracle(g, me.ID, "Witch's Clinic", "Land", witchsClinicOracle)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Test",
		Power: 2, Toughness: 2, Owner: them.ID, Controller: them.ID, IsCommander: true,
	})
	plain := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, clinic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	}); err == nil {
		t.Fatal("a non-commander was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, clinic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("no lifelink")
	}
}
