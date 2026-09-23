package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_taxes_test.go — the card half of ADR 0080 (#1063). The RULES
// are proved in game/attack_tax_test.go; these tests prove that the
// three cards in the family say what is printed on them, and that the
// one line they differ on is the one the engine reads.

const (
	propagandaOracle     = "ea9709b6-4c37-4d5a-b04d-cd4c42e4f9dd"
	ghostlyPrisonOracle  = "e828b189-0e8f-43b8-b909-4c23e742e028"
	sphereOfSafetyOracle = "92f6c063-a740-4c3c-a60a-569fd298854d"
)

func TestAttackTaxCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		propagandaOracle:     "Propaganda",
		ghostlyPrisonOracle:  "Ghostly Prison",
		sphereOfSafetyOracle: "Sphere of Safety",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s is registered as %q", oracle, spec.Name)
		}
		if len(game.AttackTaxesForCard(game.Card{OracleID: oracle})) != 1 {
			t.Errorf("%s contributes %d attack taxes, want 1", name,
				len(game.AttackTaxesForCard(game.Card{OracleID: oracle})))
		}
	}
}

// taxTable parks a four-seat game on declare attackers with one 2/2
// under seat 0 and returns the game, the attacker and the defender.
func taxTable(t *testing.T) (*game.Game, uuid.UUID, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	return g, bear, opp
}

func TestPropagandaChargesTwoPerAttackingCreature(t *testing.T) {
	g, bear, opp := taxTable(t)
	pushCatalogPermanent(g, opp.ID, "Propaganda", "Enchantment", propagandaOracle, false)

	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("attacking into Propaganda with no mana: %v, want ErrAttackTaxUnpaid", err)
	}
	fillPool(g.Seats[0], 2)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking into Propaganda with {2}: %v", err)
	}
}

func TestGhostlyPrisonIsPropagandaInWhiteAndTheTwoStack(t *testing.T) {
	g, bear, opp := taxTable(t)
	pushCatalogPermanent(g, opp.ID, "Propaganda", "Enchantment", propagandaOracle, false)
	pushCatalogPermanent(g, opp.ID, "Ghostly Prison", "Enchantment", ghostlyPrisonOracle, false)

	price := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}})
	if price.Total.Generic != 4 {
		t.Errorf("Propaganda + Ghostly Prison = %d generic, want 4 (%q)", price.Total.Generic, price.Cost)
	}
	fillPool(g.Seats[0], 3)
	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("three mana against {4}: %v, want ErrAttackTaxUnpaid", err)
	}
	fillPool(g.Seats[0], 1)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("four mana against {4}: %v", err)
	}
}

// "Creatures can't attack YOU" — the clause Propaganda prints and
// Sphere of Safety does not. An attack on a planeswalker its
// controller controls is free under Propaganda and taxed under the
// Sphere.
func TestPropagandaLeavesPlaneswalkersFreeAndSphereOfSafetyDoesNot(t *testing.T) {
	g, bear, opp := taxTable(t)
	walker := pushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker-tax", false)

	pushCatalogPermanent(g, opp.ID, "Propaganda", "Enchantment", propagandaOracle, false)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: walker}}); !p.IsFree() {
		t.Errorf("Propaganda charged %q to attack its controller's planeswalker; it prints only \"you\"", p.Cost)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); p.Total.Generic != 2 {
		t.Errorf("Propaganda charged %d to attack its controller, want 2", p.Total.Generic)
	}

	pushCatalogPermanent(g, opp.ID, "Sphere of Safety", "Enchantment", sphereOfSafetyOracle, false)
	p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: walker}})
	if p.Total.Generic != 2 {
		// Two enchantments on their board, so the Sphere's X is 2.
		t.Errorf(`Sphere of Safety charged %d to attack a planeswalker, want 2 — it prints "or planeswalkers you control"`, p.Total.Generic)
	}
}

// Sphere of Safety's X is "the number of enchantments you control",
// counted at declaration and including the Sphere itself.
func TestSphereOfSafetyCountsTheDefendersEnchantments(t *testing.T) {
	g, bear, opp := taxTable(t)
	pushCatalogPermanent(g, opp.ID, "Sphere of Safety", "Enchantment", sphereOfSafetyOracle, false)

	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); p.Total.Generic != 1 {
		t.Errorf("a lone Sphere counts itself: %d, want 1", p.Total.Generic)
	}
	pushCatalogPermanent(g, opp.ID, "Propaganda", "Enchantment", propagandaOracle, false)
	// Two enchantments now: the Sphere charges {2}, Propaganda {2}.
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); p.Total.Generic != 4 {
		t.Errorf("Sphere({2}) + Propaganda({2}) = %d, want 4", p.Total.Generic)
	}
	// An enchantment on SOMEONE ELSE's board does not count.
	pushCatalogPermanent(g, g.Seats[2].ID, "Ghostly Prison", "Enchantment", ghostlyPrisonOracle, false)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); p.Total.Generic != 4 {
		t.Errorf(`"enchantments YOU control" counted a third seat's: %d, want 4`, p.Total.Generic)
	}
}

// An attack tax is not a restriction, and the two must not be
// confused: the creature is perfectly able to attack, and the engine
// lets it through the moment the tax is paid.
func TestAnAttackTaxDoesNotRestrictTheCreature(t *testing.T) {
	g, bear, opp := taxTable(t)
	pushCatalogPermanent(g, opp.ID, "Ghostly Prison", "Enchantment", ghostlyPrisonOracle, false)

	var restricted bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == bear {
				restricted = !game.CanAttack(c)
			}
		}
	})
	if restricted {
		t.Error("Ghostly Prison set a can't-attack restriction; it charges a cost instead")
	}
	// A third seat with no enchantment is attacked for free.
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatalf("attacking an untaxed seat: %v", err)
	}
}
