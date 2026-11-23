package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCabalStrongholdCountsOnlyYourBasicSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stronghold := effectstest.SeedPermanentWithOracle(g, me.ID, "Cabal Stronghold", "Land", effectstest.CabalStrongholdOracle)
	effectstest.SeedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	effectstest.SeedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	effectstest.SeedManaLand(g, me.ID, "Watery Grave", "Land — Island Swamp", "U", "B")
	effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	effectstest.SeedManaLand(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp", "B")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateManaAbility(me.ID, stronghold, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want two {B} (two basic Swamps, the {3} was spent)", me.ManaPool)
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "B" {
			t.Errorf("token %v, want black", tok)
		}
	}
}

func TestTaintedPeakNeedsASwampForItsColoredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	peak := effectstest.SeedPermanentWithOracle(g, me.ID, "Tainted Peak", "Land", effectstest.TaintedPeakOracle)
	if err := g.ActivateManaAbility(me.ID, peak, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Tainted Peak made {B} or {R} with no Swamp")
	}
	effectstest.SeedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	if err := g.ActivateManaAbility(me.ID, peak, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("with a Swamp: %v", err)
	}
	if pick := effectstest.ManaPickFor(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Errorf("mana pick = %+v, want a {B} or {R} choice", pick)
	}
}

func TestKessigWolfRunPumpsByXAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	run := effectstest.SeedPermanentWithOracle(g, me.ID, "Kessig Wolf Run", "Land", effectstest.KessigWolfRunOracle)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := effectstest.EffectivePower(t, g, bear)
	if err := g.ActivateCatalogAbility(me.ID, run, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != before+2 {
		t.Errorf("power = %d, want %d", got, before+2)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestMirenGainsLifeEqualToTheSacrificedToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	miren := effectstest.SeedPermanentWithOracle(g, me.ID, "Miren, the Moaning Well", "Legendary Land", effectstest.MirenOracle)
	ox := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ox", TypeLine: "Creature — Ox",
		Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, miren, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{ox},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestNestingGroundsMovesTheOnlyKindWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grounds := effectstest.SeedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", effectstest.NestingGroundsOracle)
	from := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Source", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	to := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dest", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	src, _ := effectstest.BattlefieldCard(g, from)
	dst, _ := effectstest.BattlefieldCard(g, to)
	if src.Counters[game.CounterPlusOne] != 1 || dst.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters source=%v dest=%v, want one each", src.Counters, dst.Counters)
	}
}

func TestNestingGroundsAsksWhichKindWhenThereAreSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grounds := effectstest.SeedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", effectstest.NestingGroundsOracle)
	from := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Source", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, "stun": 1},
	})
	to := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dest", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds are offered in name order: "+1/+1" before "stun".
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	src, _ := effectstest.BattlefieldCard(g, from)
	dst, _ := effectstest.BattlefieldCard(g, to)
	if src.Counters["stun"] != 0 || dst.Counters["stun"] != 1 || src.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters source=%v dest=%v, want the stun counter moved", src.Counters, dst.Counters)
	}
}

func TestNestingGroundsIsSorcerySpeedAndNeedsTwoDifferentPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grounds := effectstest.SeedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", effectstest.NestingGroundsOracle)
	one := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "One", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: one}},
	}); err == nil {
		t.Fatal("the same permanent was accepted as both targets")
	}
}

func TestCastleGarenbrigMakesSixRestrictedGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	castle := effectstest.SeedPermanentWithOracle(g, me.ID, "Castle Garenbrig", "Land", effectstest.CastleGarenbrigOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, castle, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 6 {
		t.Fatalf("pool = %v, want six {G}", me.ManaPool)
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "G" || len(tok.Restrictions) != 1 {
			t.Errorf("token %+v, want restricted {G}", tok)
		}
	}
}

func TestWitchsClinicTargetsAnyPlayersCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	clinic := effectstest.SeedPermanentWithOracle(g, me.ID, "Witch's Clinic", "Land", effectstest.WitchsClinicOracle)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Test",
		Power: 2, Toughness: 2, Owner: them.ID, Controller: them.ID, IsCommander: true,
	})
	plain := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("no lifelink")
	}
}
