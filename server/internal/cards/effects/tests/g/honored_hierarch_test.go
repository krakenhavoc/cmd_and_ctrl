package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Honored Hierarch (Magic Origins Clash Pack, #5550).

const honoredHierarchOracle = "ae42fa8a-b0f6-43dc-a4c6-e2531890ca3a"

func TestHonoredHierarchIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(honoredHierarchOracle)
	if !ok {
		t.Fatal("Honored Hierarch is not registered")
	}
	if spec.Name != "Honored Hierarch" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// Before it is renowned the Hierarch has neither vigilance nor a mana
// ability, so it cannot tap for mana.
func TestHonoredHierarchHasNothingUntilRenowned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hierarch := effectstest.PushRenownCard(g, me.ID, "Honored Hierarch", honoredHierarchOracle, 1, 1)
	if game.HasKeyword(effectstest.MustBattlefieldCard(t, g, hierarch), "vigilance") {
		t.Error("the Hierarch has vigilance before it is renowned")
	}
	if n := effectstest.GrantedManaRowsOf(g, hierarch); n != 0 {
		t.Errorf("the Hierarch has %d granted mana rows before it is renowned, want 0", n)
	}
	if err := g.ActivateManaAbility(me.ID, hierarch, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("an unrenowned Hierarch tapped for mana")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool holds %d mana, want none", me.ManaPool.Total())
	}
}

// Combat damage to a player makes it renowned with a +1/+1 counter, and
// then it has vigilance and taps for one mana of any color.
func TestHonoredHierarchBecomesRenownedAndTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hierarch := effectstest.PushRenownCard(g, me.ID, "Honored Hierarch", honoredHierarchOracle, 1, 1)
	effectstest.AttackWith(t, g, opp.ID, hierarch)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.RenownedNow(g, hierarch) || effectstest.PlusOneCounters(g, hierarch) != 1 {
		t.Fatalf("renowned %v, %d counters; want renowned with 1", effectstest.RenownedNow(g, hierarch), effectstest.PlusOneCounters(g, hierarch))
	}
	if !game.HasKeyword(effectstest.MustBattlefieldCard(t, g, hierarch), "vigilance") {
		t.Error("a renowned Hierarch has no vigilance")
	}
	if n := effectstest.GrantedManaRowsOf(g, hierarch); n != 1 {
		t.Fatalf("a renowned Hierarch has %d granted mana rows, want 1", n)
	}
	// It attacked before it had vigilance, so it is tapped.
	if err := effectstest.TapGrantedMana(t, g, me.ID, hierarch, "W"); err == nil {
		t.Error("a tapped Hierarch tapped for mana")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hierarch) })
	if err := effectstest.TapGrantedMana(t, g, me.ID, hierarch, "B"); err != nil {
		t.Fatalf("tap the renowned Hierarch: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want one {B}", got)
	}
	if !effectstest.MustBattlefieldCard(t, g, hierarch).Tapped {
		t.Error("the Hierarch is untapped after tapping for mana")
	}
}

// A renowned Hierarch attacks without tapping.
func TestHonoredHierarchAttacksWithVigilanceOnceRenowned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hierarch := effectstest.PushRenownCard(g, me.ID, "Honored Hierarch", honoredHierarchOracle, 1, 1)
	effectstest.MakeRenowned(t, g, hierarch)
	effectstest.AttackWith(t, g, opp.ID, hierarch)
	if effectstest.MustBattlefieldCard(t, g, hierarch).Tapped {
		t.Error("a renowned Hierarch tapped to attack")
	}
}

// The mana ability is the Hierarch's own: a summoning-sick Hierarch
// cannot pay {T} even once renowned (CR 302.6).
func TestHonoredHierarchCannotTapForManaWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hierarch := effectstest.PushCatalogPermanent(g, me.ID, "Honored Hierarch", "Creature — Human Druid", honoredHierarchOracle, true)
	effectstest.MakeRenowned(t, g, hierarch)
	if n := effectstest.GrantedManaRowsOf(g, hierarch); n != 1 {
		t.Fatalf("a renowned Hierarch has %d granted mana rows, want 1", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, hierarch, "G"); err == nil {
		t.Fatal("a summoning-sick Hierarch tapped for mana")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool holds %d mana after a refused activation, want none", me.ManaPool.Total())
	}
}
