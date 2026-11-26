package game

import "testing"

// block_tax_this_turn_test.go — War Cadence's turn-long block tax (S82 OLD
// b385, CR 509.1d): priced per blocking creature, paid as the declaration
// finishes, and gone with the turn.

func TestBlockTaxThisTurnChargesEachBlockerAndIsRefusedUnpaid(t *testing.T) {
	g, _, bear, wall1, wall2 := blockTaxTable(t, false)
	def := g.Seats[1]
	g.WithWriteLock(func() {
		if !g.BlockTaxThisTurnForEffect(bear, 2, "War Cadence") {
			t.Fatal("the tax was not registered")
		}
		if g.BlockTaxThisTurnForEffect(bear, 0, "War Cadence") {
			t.Fatal("a tax of {0} was registered")
		}
	})
	if err := block(t, g, wall1, bear, wall2, bear); err != nil {
		t.Fatalf("staging: %v", err)
	}
	if got := g.PriceBlockDeclaration(def.ID); got.Cost != "{2}{2}" || len(got.Lines) != 2 {
		t.Fatalf("price = %+v, want {2}{2} over two lines", got)
	}
	addMana(def, 3)
	blockTaxErr(t, g.FinishBlocks(def.ID), "{2}{2}")
	addMana(def, 1)
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("paid finish: %v", err)
	}
	if n := def.ManaPool.Total(); n != 0 {
		t.Errorf("pool = %d, want 0", n)
	}
}

func TestBlockTaxThisTurnModProblem(t *testing.T) {
	if blockTaxThisTurnModProblem(Mod{Kind: ModBlockTaxThisTurn}) == "" {
		t.Error("a {0} tax passed the check")
	}
	if p := blockTaxThisTurnModProblem(Mod{Kind: ModBlockTaxThisTurn, Amount: 3}); p != "" {
		t.Errorf("a sound mod failed the check: %s", p)
	}
}
