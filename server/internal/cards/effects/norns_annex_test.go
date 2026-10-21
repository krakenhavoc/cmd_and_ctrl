package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// norns_annex_test.go — #4198. The card says {W/P} per attacking
// creature; these tests prove each symbol is paid with {W} or with 2
// life, at the attacker's choice, and that the engine refuses a claim
// the cost or CR 119.4 does not allow.

const nornsAnnexOracle = "9a1fbe72-4a17-42be-8e23-d7d30a5e59c1"

func nornsTable(t *testing.T, attackers int) (*game.Game, []uuid.UUID, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < attackers; i++ {
		ids = append(ids, pushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	pushCatalogPermanent(g, opp.ID, "Norn's Annex", "Artifact", nornsAnnexOracle, false)
	return g, ids, opp
}

func nornsDecls(ids []uuid.UUID, target uuid.UUID) []game.AttackDeclaration {
	out := make([]game.AttackDeclaration, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.AttackDeclaration{Attacker: id, Target: target})
	}
	return out
}

func TestNornsAnnexIsFullAndTaxesOnePhyrexianWhitePerCreature(t *testing.T) {
	spec, ok := Lookup(nornsAnnexOracle)
	if !ok || spec.Name != "Norn's Annex" {
		t.Fatalf("Norn's Annex is not registered: %+v", spec)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats %v: want Full with none", spec.Completeness, spec.Caveats)
	}
	g, ids, opp := nornsTable(t, 3)
	p := g.PriceAttackDeclaration(nornsDecls(ids, opp.ID))
	if p.Cost != "{W/P}{W/P}{W/P}" {
		t.Errorf("three attackers cost %q, want {W/P}{W/P}{W/P}", p.Cost)
	}
	if p.Total.PhyrexianSymbols() != 3 {
		t.Errorf("tax prints %d Phyrexian symbols, want 3", p.Total.PhyrexianSymbols())
	}
}

func TestNornsAnnexPaysWithWhiteMana(t *testing.T) {
	g, ids, opp := nornsTable(t, 1)
	me := g.Seats[0]
	if err := g.DeclareAttacker(ids[0], opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("no mana and no claim: %v, want ErrAttackTaxUnpaid", err)
	}
	before := me.Life
	fillPoolColored(me, "W", 1)
	if err := g.DeclareAttacker(ids[0], opp.ID); err != nil {
		t.Fatalf("attacking with {W}: %v", err)
	}
	if me.Life != before {
		t.Errorf("life %d -> %d: paying with mana must cost no life", before, me.Life)
	}
}

func TestNornsAnnexPaysWithTwoLifePerSymbol(t *testing.T) {
	g, ids, opp := nornsTable(t, 3)
	me := g.Seats[0]
	before := me.Life
	fillPoolColored(me, "W", 1)

	// Two symbols with life, one with the {W}.
	_, err := g.DeclareAttackersWith(nornsDecls(ids, opp.ID), game.DeclareAttackersParams{PhyrexianLife: 2})
	if err != nil {
		t.Fatalf("three attackers, one {W}, two symbols with life: %v", err)
	}
	if me.Life != before-4 {
		t.Errorf("life %d -> %d, want 4 paid for two symbols", before, me.Life)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, the {W} should have paid the third symbol", got)
	}
}

func TestNornsAnnexRefusesMoreLifeThanThereAreSymbols(t *testing.T) {
	g, ids, opp := nornsTable(t, 2)
	me := g.Seats[0]
	before := me.Life
	_, err := g.DeclareAttackersWith(nornsDecls(ids, opp.ID), game.DeclareAttackersParams{PhyrexianLife: 3})
	if !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("claiming 3 of 2 symbols: %v, want ErrAttackTaxUnpaid", err)
	}
	if me.Life != before {
		t.Errorf("a refused declaration cost %d life", before-me.Life)
	}
}

func TestNornsAnnexCannotPayLifeBelowNothing(t *testing.T) {
	g, ids, opp := nornsTable(t, 2)
	me := g.Seats[0]
	me.Life = 3 // CR 119.4: one symbol is 2 life, two would be 4.
	_, err := g.DeclareAttackersWith(nornsDecls(ids, opp.ID), game.DeclareAttackersParams{PhyrexianLife: 2})
	if !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("paying 4 life from 3: %v, want ErrAttackTaxUnpaid", err)
	}
	if me.Life != 3 {
		t.Errorf("life = %d after the refusal, want 3", me.Life)
	}
}

// "Or planeswalkers you control": an attack on its controller's
// planeswalker pays the same {W/P}.
func TestNornsAnnexTaxesAttacksOnPlaneswalkersToo(t *testing.T) {
	g, ids, opp := nornsTable(t, 1)
	walker := pushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker-norn", false)
	p := g.PriceAttackDeclaration(nornsDecls(ids, walker))
	if p.Cost != "{W/P}" {
		t.Errorf("attacking a planeswalker costs %q, want {W/P}", p.Cost)
	}
	// A third seat is untaxed.
	if p := g.PriceAttackDeclaration(nornsDecls(ids, g.Seats[2].ID)); !p.IsFree() {
		t.Errorf("attacking another seat cost %q, want nothing", p.Cost)
	}
}
