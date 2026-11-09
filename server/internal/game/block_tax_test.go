package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// block_tax_test.go — ADR 0221 / #5375: CR 509.1d's cost to block,
// paid as the defender's declaration completes. Archangel of Tithes
// proves the same rules from the catalog side in cards/effects.

const tithesTestOracle = "test-archangel-of-tithes"

// withCatalogBlockTaxes installs a CatalogBlockTaxes shim for the test.
func withCatalogBlockTaxes(t *testing.T, fn func(key string) []BlockTax) {
	t.Helper()
	prev := CatalogBlockTaxes
	CatalogBlockTaxes = fn
	t.Cleanup(func() { CatalogBlockTaxes = prev })
}

// tithesBlockTax is Archangel of Tithes' block half as the engine
// reads it.
func tithesBlockTax() BlockTax {
	return BlockTax{
		Label:     "As long as this creature is attacking, creatures can't block unless their controller pays {1} for each of those creatures.",
		AppliesTo: WhileSourceAttacking,
		ManaCost:  func(BlockTaxQuery) string { return "{1}" },
	}
}

// blockTaxTable is seat 0 attacking seat 1 with an Archangel (carrying
// the shimmed block tax) and a Bear, and seat 1 holding two Walls. The
// cursor is parked in declare blockers with seat 1 declaring.
func blockTaxTable(t *testing.T, angelAttacks bool) (g *Game, angel, bear, wall1, wall2 uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	withCatalogBlockTaxes(t, func(key string) []BlockTax {
		if key == tithesTestOracle {
			return []BlockTax{tithesBlockTax()}
		}
		return nil
	})
	angel = pushCombatant(t, g, g.Seats[0], "Archangel", 3, 5)
	findCard(g, angel).OracleID = tithesTestOracle
	bear = pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	wall1 = pushCombatant(t, g, g.Seats[1], "Wall One", 0, 4)
	wall2 = pushCombatant(t, g, g.Seats[1], "Wall Two", 0, 4)
	if angelAttacks {
		declareAttacks(t, g, angel, bear)
	} else {
		declareAttacks(t, g, bear)
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationPending {
		t.Fatalf("defender status = %q, want pending", got)
	}
	return g, angel, bear, wall1, wall2
}

func blockTaxErr(t *testing.T, err error, cost string) {
	t.Helper()
	if !errors.Is(err, ErrBlockTaxUnpaid) {
		t.Fatalf("err = %v, want ErrBlockTaxUnpaid", err)
	}
	var te *BlockTaxUnpaidError
	if !errors.As(err, &te) || te.Cost != cost {
		t.Fatalf("the refusal carries %#v, want cost %q", te, cost)
	}
}

func TestBlockTaxIsChargedPerBlockerWhileTheSourceAttacks(t *testing.T) {
	g, angel, bear, wall1, wall2 := blockTaxTable(t, true)
	def := g.Seats[1]
	// Staging is free and revisable: the price is not determined yet.
	if err := block(t, g, wall1, bear, wall2, angel); err != nil {
		t.Fatalf("staging two blocks: %v", err)
	}
	if got := g.PriceBlockDeclaration(def.ID); got.Cost != "{1}{1}" || len(got.Lines) != 2 {
		t.Fatalf("price = %+v, want {1}{1} over two lines", got)
	}
	// Unpaid: refused, nothing spent, still declaring, blocks staged.
	addMana(def, 1)
	blockTaxErr(t, g.FinishBlocks(def.ID), "{1}{1}")
	if g.BlockDeclarationStatusOf(def.ID) != BlockDeclarationPending {
		t.Fatal("a refused finish completed the declaration")
	}
	if n := def.ManaPool.Total(); n != 1 {
		t.Errorf("pool = %d, want 1 — a refused finish spends nothing (CR 509.1f)", n)
	}
	if c := findCard(g, wall1); c.BlockingTarget != bear {
		t.Error("a refused finish took the staged block back")
	}
	// Paid: completes, both blocks stand, the pool is empty.
	addMana(def, 1)
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("finish with {2} floating: %v", err)
	}
	if g.BlockDeclarationStatusOf(def.ID) != BlockDeclarationDeclared {
		t.Fatal("the paid finish did not complete the declaration")
	}
	if n := def.ManaPool.Total(); n != 0 {
		t.Errorf("pool = %d, want 0", n)
	}
	if findCard(g, wall1).BlockingTarget != bear || findCard(g, wall2).BlockingTarget != angel {
		t.Error("the paid blocks are not standing")
	}
}

func TestBlockTaxChargesNothingWhileTheSourceIsNotAttacking(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, false)
	def := g.Seats[1]
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	if p := g.PriceBlockDeclaration(def.ID); !p.IsFree() {
		t.Fatalf("price = %q with the Archangel at home, want free", p.Cost)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("an untaxed finish with an empty pool: %v", err)
	}
}

func TestNoBlocksUnderABlockTaxIsFree(t *testing.T) {
	g, _, _, _, _ := blockTaxTable(t, true)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatalf("declaring no blocks owes nothing: %v", err)
	}
}

func TestTheAutoTapperPaysTheBlockTax(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, true)
	def := g.Seats[1]
	land := pushBattlefieldForTest(g, def.ID, "Forest", "Basic Land — Forest", taxForestTestOracle)
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	// The zero posture is strict: an untapped land is not reached for.
	blockTaxErr(t, g.FinishBlocks(def.ID), "{1}")
	if err := g.FinishBlocksWith(def.ID, DeclareBlockersParams{AutoTap: true}); err != nil {
		t.Fatalf("finish with auto-tap and a Forest: %v", err)
	}
	if !findCard(g, land).Tapped {
		t.Error("the Forest was not tapped for the block tax")
	}
	if g.BlockDeclarationStatusOf(def.ID) != BlockDeclarationDeclared {
		t.Error("the auto-tapped finish did not complete")
	}
}

func TestALockedLandIsNotTappedForTheBlockTax(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, true)
	def := g.Seats[1]
	land := pushBattlefieldForTest(g, def.ID, "Forest", "Basic Land — Forest", taxForestTestOracle)
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	blockTaxErr(t, g.FinishBlocksWith(def.ID, DeclareBlockersParams{AutoTap: true, LockedSources: []uuid.UUID{land}}), "{1}")
	if findCard(g, land).Tapped {
		t.Error("a locked land was tapped")
	}
}

func TestAPendingDefendersPassPaysFromThePoolOrIsRefused(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, true)
	def := g.Seats[1]
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	passToDefender(t, g)
	blockTaxErr(t, g.PassPriority(), "{1}")
	if g.BlockDeclarationStatusOf(def.ID) != BlockDeclarationPending {
		t.Fatal("a refused pass completed the declaration")
	}
	addMana(def, 1)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the pass with {1} floating: %v", err)
	}
	if g.BlockDeclarationStatusOf(def.ID) != BlockDeclarationDeclared || def.ManaPool.Total() != 0 {
		t.Error("the paid pass did not complete and charge the declaration")
	}
}

func TestTheStepEndingTakesBackBlocksThePoolCannotPayFor(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, true)
	def := g.Seats[1]
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if c := findCard(g, wall1); c.BlockingTarget != uuid.Nil {
		t.Error("an unpaid block survived the step ending")
	}
	evs := blockersDeclaredEvents(g, def.ID)
	if len(evs) != 1 || evs[0].Amount != 0 {
		t.Errorf("want the declaration announced as none, got %+v", evs)
	}
}

func TestTheStepEndingPaysFromThePoolWhenItCan(t *testing.T) {
	g, _, bear, wall1, _ := blockTaxTable(t, true)
	def := g.Seats[1]
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	addMana(def, 1)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if findCard(g, wall1).BlockingTarget != bear {
		t.Error("a block the pool could pay for was taken back")
	}
	if n := def.ManaPool.Total(); n != 0 {
		t.Errorf("pool = %d, want 0 — the tax was charged", n)
	}
}

// CR 509.1c: a requirement is not owed through a block tax. A "blocks
// each combat if able" Wall under an attacking Archangel may stay home.
func TestARequirementIsNotOwedThroughABlockTax(t *testing.T) {
	g, _, _, wall1, _ := blockTaxTable(t, true)
	withBlockRequirement(t, g, wall1, BlockRequirementBlocks)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatalf("finishing with the taxed Wall at home: %v", err)
	}
}

// The control: without the tax the same requirement is owed.
func TestARequirementIsOwedWithoutTheBlockTax(t *testing.T) {
	g, _, _, wall1, _ := blockTaxTable(t, false)
	withBlockRequirement(t, g, wall1, BlockRequirementBlocks)
	blockRequirementErr(t, g.FinishBlocks(g.Seats[1].ID))
}

func TestAnAbilityRemovedArchangelTaxesNoBlocks(t *testing.T) {
	g, angel, bear, wall1, _ := blockTaxTable(t, true)
	g.WithWriteLock(func() {
		c := findCard(g, angel)
		c.OracleID = "" // its catalog key gone, as under CR 613.1f
	})
	if err := block(t, g, wall1, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	if p := g.PriceBlockDeclaration(g.Seats[1].ID); !p.IsFree() {
		t.Errorf("price = %q, want free", p.Cost)
	}
}
