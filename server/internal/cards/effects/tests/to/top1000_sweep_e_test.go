package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_e_test.go — the fifth slice of the missing top-1,000
// cards (#3654): Elspeth Storm Slayer, Descent into Avernus, Vesuva
// and Ashaya, Soul of the Wild.

const (
	sweepEElspethOracle = "f78af825-023a-42e9-8374-5c52303a1417"
	sweepEDescentOracle = "1313da2f-92bb-4862-9153-373972cc1520"
	sweepEVesuvaOracle  = "4001b868-ada1-43f4-92e2-27ab0e80c913"
	sweepEAshayaOracle  = "162572f2-1757-42e9-bd97-e6bd9a762c0e"
)

func TestSweepEElspethDoublesTokensAndPlusOneMakesTwoSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Storm Slayer", "Legendary Planeswalker — Elspeth", sweepEElspethOracle, 5)

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldier tokens = %d, want 2 (her +1 is doubled by her own static)", n)
	}
	if got := effectstest.LoyaltyCount(g, elspeth); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestSweepEElspethZeroCountersEveryCreatureAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Storm Slayer", "Legendary Planeswalker — Elspeth", sweepEElspethOracle, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		v := effectstest.S58p6View(t, g, id)
		if v.Counters["+1/+1"] != 1 {
			t.Errorf("%s has %d counters, want 1", v.Name, v.Counters["+1/+1"])
		}
		if !game.HasKeyword(&v, "flying") {
			t.Errorf("%s did not gain flying", v.Name)
		}
	}
	v := effectstest.S58p6View(t, g, theirs)
	if v.Counters["+1/+1"] != 0 || game.HasKeyword(&v, "flying") {
		t.Error("an opponent's creature was affected")
	}
}

func TestSweepEElspethMinusThreeDestroysABigOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Storm Slayer", "Legendary Planeswalker — Elspeth", sweepEElspethOracle, 5)
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Elf",
		ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})

	err := g.ActivateCatalogAbility(me.ID, elspeth, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}})
	if err == nil {
		t.Fatal("a mana value 2 creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, elspeth, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, big) {
		t.Error("the mana value 3 creature survived")
	}
	if !effects.OnBattlefield(g, small) {
		t.Error("the small creature was destroyed")
	}
}

func TestSweepEDescentIntoAvernusEscalatesEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avernus := effectstest.PushCatalogPermanent(g, me.ID, "Descent into Avernus", "Enchantment", sweepEDescentOracle, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("expected seat 0's upkeep, got seat %d", g.Turn.ActiveSeat)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AllCountersOn(t, g, avernus)["descent"]; got != 2 {
		t.Fatalf("descent counters %d, want 2", got)
	}
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Errorf("my Treasures %d, want 2", got)
	}
	if got := effectstest.CountTreasures(g, opp.ID); got != 2 {
		t.Errorf("their Treasures %d, want 2 (each player)", got)
	}
}

func TestSweepEVesuvaCopiesALandAndEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")

	id := effectstest.CastCatalogSpell(t, g, "Vesuva", "Land", sweepEVesuvaOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, forest)

	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Forest" {
		t.Errorf("name %q, want Forest", got.Name)
	}
	if !got.Tapped {
		t.Error("the copy did not enter tapped")
	}
}

func TestSweepEVesuvaDeclinedEntersUntappedAsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")

	id := effectstest.CastCatalogSpell(t, g, "Vesuva", "Land", sweepEVesuvaOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)

	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Vesuva" || got.Tapped {
		t.Errorf("declined: name %q tapped %v, want an untapped Vesuva", got.Name, got.Tapped)
	}
}

func TestSweepEAshayaMakesNontokenCreaturesForestLandsAndCountsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ashaya := effectstest.PushCatalogPermanent(g, me.ID, "Ashaya, Soul of the Wild", "Legendary Creature — Elemental", sweepEAshayaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })

	v := effectstest.S58p6View(t, g, bear)
	if !v.IsLand() || !v.HasSubtype("Forest") {
		t.Errorf("the nontoken Bear is not a Forest land: %v / %v", v.Effective().Types, v.Effective().Subtypes)
	}
	if got := effectstest.ManaAbilityProduced(t, g, bear); len(got) != 1 || got[0] != "{G}" {
		t.Errorf("the Bear's mana abilities %v, want [{G}]", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			tv := effectstest.S58p6View(t, g, c.InstanceID)
			if tv.IsLand() {
				t.Error("a creature token became a land")
			}
		}
	}
	// Lands you control: Ashaya, the Bear and the Forest.
	a := effectstest.S58p6View(t, g, ashaya)
	if a.CurrentPower() != 3 || a.CurrentToughness() != 3 {
		t.Errorf("Ashaya is %d/%d, want 3/3", a.CurrentPower(), a.CurrentToughness())
	}
}
