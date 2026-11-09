package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// archangel_of_tithes_test.go — #5375. The block-tax RULES are proved
// in game/block_tax_test.go (ADR 0221); these prove the card says what
// is printed: an attack tax while untapped, a block tax while
// attacking, {1} per creature each.

const archangelOfTithesOracle = "ff5caed4-0276-476d-8fae-edae2536df7f"

func TestArchangelOfTithesIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(archangelOfTithesOracle)
	if !ok {
		t.Fatal("Archangel of Tithes is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v caveats %v, want full with none", spec.Completeness, spec.Caveats)
	}
	c := game.Card{OracleID: archangelOfTithesOracle}
	if len(game.AttackTaxesForCard(c)) != 1 || len(game.BlockTaxesForCard(c)) != 1 {
		t.Error("want one attack tax and one block tax")
	}
}

func TestArchangelOfTithesUntappedTaxesAttacksOnItsController(t *testing.T) {
	g, bear, opp := taxTable(t)
	pushCatalogPermanent(g, opp.ID, "Archangel of Tithes", "Creature — Angel", archangelOfTithesOracle, false)

	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("attacking into an untapped Archangel with no mana: %v, want ErrAttackTaxUnpaid", err)
	}
	fillPool(g.Seats[0], 1)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking with {1}: %v", err)
	}
}

func TestArchangelOfTithesTappedTaxesNoAttacks(t *testing.T) {
	g, bear, opp := taxTable(t)
	angel := pushCatalogPermanent(g, opp.ID, "Archangel of Tithes", "Creature — Angel", archangelOfTithesOracle, false)
	findBattlefieldCardForTest(g, angel).Tapped = true

	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); !p.IsFree() {
		t.Fatalf("a tapped Archangel priced the attack at %q", p.Cost)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking past a tapped Archangel: %v", err)
	}
}

// tithesBlockTable: seat 0's Archangel (and a Bear) in declare
// attackers, seat 1 holding two 0/4 Walls.
func tithesBlockTable(t *testing.T, angelAttacks bool) (g *game.Game, bear uuid.UUID, def *game.Player, walls [2]uuid.UUID) {
	t.Helper()
	g, bear, def = taxTable(t)
	angel := pushCatalogPermanent(g, g.Seats[0].ID, "Archangel of Tithes", "Creature — Angel", archangelOfTithesOracle, false)
	walls[0] = pushVanillaCreature(g, def.ID, "Wall One", 0, 4)
	walls[1] = pushVanillaCreature(g, def.ID, "Wall Two", 0, 4)
	decls := []game.AttackDeclaration{{Attacker: bear, Target: def.ID}}
	if angelAttacks {
		decls = append(decls, game.AttackDeclaration{Attacker: angel, Target: def.ID})
	}
	if _, err := g.DeclareAttackers(decls); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, bear, def, walls
}

func TestArchangelOfTithesAttackingTaxesEachBlocker(t *testing.T) {
	g, bear, def, walls := tithesBlockTable(t, true)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: walls[0], Attacker: bear},
		{Blocker: walls[1], Attacker: bear},
	}); err != nil {
		t.Fatalf("staging two blocks: %v", err)
	}
	if p := g.PriceBlockDeclaration(def.ID); p.Cost != "{1}{1}" {
		t.Fatalf("two blockers cost %q, want {1}{1}", p.Cost)
	}
	fillPool(def, 1)
	if err := g.FinishBlocks(def.ID); !errors.Is(err, game.ErrBlockTaxUnpaid) {
		t.Fatalf("finishing two blocks with {1}: %v, want ErrBlockTaxUnpaid", err)
	}
	fillPool(def, 1)
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("finishing two blocks with {2}: %v", err)
	}
	if g.BlockDeclarationStatusOf(def.ID) != game.BlockDeclarationDeclared {
		t.Error("the paid declaration did not complete")
	}
}

func TestArchangelOfTithesAtHomeTaxesNoBlocks(t *testing.T) {
	g, bear, def, walls := tithesBlockTable(t, false)
	if err := g.DeclareBlocker(walls[0], bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("a block with the Archangel at home is free: %v", err)
	}
}
