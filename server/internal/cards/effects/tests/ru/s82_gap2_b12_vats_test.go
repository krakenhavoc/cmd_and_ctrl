package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_vats_test.go — S82 GAP2 batch 12: V.A.T.S., through the real
// cast, the set-rule announce gate and the CR 608.2b re-check.

const gap2B12Vats = "d077d54c-6f1b-4795-a241-6073b0dc1dc8"

func TestGap2B12VatsIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Vats)
	if !ok {
		t.Fatal("V.A.T.S. is not registered")
	}
	if spec.Name != "V.A.T.S." || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func gap2B12CastVats(t *testing.T, g *game.Game, targets ...[]game.TargetRef) error {
	t.Helper()
	var refs []game.TargetRef
	for _, r := range targets {
		refs = append(refs, r...)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("B", "B", "C", "C")...) })
	return effectstest.CastCatalogSpellErrStrict(t, g, "V.A.T.S.", "Instant", gap2B12Vats, refs)
}

func TestGap2B12VatsDestroysCreaturesWithEqualToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, opp.ID, "Two A", 3, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Two B", 1, 2)
	c := effectstest.PushVanillaCreature(g, opp.ID, "Three", 2, 3)
	if err := gap2B12CastVats(t, g, effectstest.CardTarget(a), effectstest.CardTarget(c)); err == nil {
		t.Fatal("creatures with different toughness were accepted")
	}
	if err := gap2B12CastVats(t, g, effectstest.CardTarget(a), effectstest.CardTarget(b)); err != nil {
		t.Fatalf("casting at two creatures with toughness 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the chosen creatures were not destroyed")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an unchosen creature was destroyed")
	}
}

// A creature pumped in response no longer has equal toughness with the
// others: it and every pick it disagrees with are illegal (CR 608.2b).
func TestGap2B12VatsToughnessChangeInResponseSavesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, opp.ID, "Two A", 3, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Two B", 1, 2)
	if err := gap2B12CastVats(t, g, effectstest.CardTarget(a), effectstest.CardTarget(b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, a).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("a creature was destroyed though the chosen creatures no longer had equal toughness")
	}
}

// A pick that left the battlefield does not save the others.
func TestGap2B12VatsAPickThatLeftLeavesTheRestDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, opp.ID, "Two A", 3, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Two B", 1, 2)
	if err := gap2B12CastVats(t, g, effectstest.CardTarget(a), effectstest.CardTarget(b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Destroy(t, g, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) {
		t.Error("the remaining pick survived")
	}
}
