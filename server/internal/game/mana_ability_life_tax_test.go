package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// #5661: "Mana abilities of this land cost an additional 1 life to
// activate", driven through a stubbed static so the engine contract is
// pinned without a catalog card (Thran Portal's test is in
// cards/effects).

const lifeTaxOracle = "test-mana-ability-life-tax"

// withManaLifeTax makes every permanent with lifeTaxOracle charge
// `life` more for its mana abilities, and makes a permanent with
// removerOracle (when one is named) strip every land's abilities in
// layer 6 at its own timestamp.
func withManaLifeTax(t *testing.T, life int, removerOracle string) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case lifeTaxOracle:
			return []StaticAbility{{
				Layer:     Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.ManaAbilityLifeTax += life
				},
			}}
		case removerOracle:
			return []StaticAbility{{
				Layer:            Layer6Ability,
				RemovesAbilities: true,
				AppliesTo:        func(target *Card, _ *Game, _ *Card) bool { return target.IsLand() },
				Apply:            func(*Characteristic, *Card, *Game, *Card) {},
			}}
		}
		return nil
	})
}

// The surcharge lands on the intrinsic row (CR 305.6) and on a row the
// card carries, with whatever life that row already costs.
func TestManaAbilityLifeTaxIsAddedToEveryRow(t *testing.T) {
	withManaLifeTax(t, 1, "")
	g := newActiveGame(t)
	me := g.Seats[0]
	taxed := pushTypedTestCard(g, Card{
		Name: "Taxed Forest", TypeLine: "Land — Forest", OracleID: lifeTaxOracle,
		Owner: me.ID, Controller: me.ID,
		ManaAbilities: []ManaAbilityShape{{Label: "{T}, Pay 1 life: Add {B}", TapCost: true, LifeCost: 1, Produced: "{B}"}},
	})
	plain := pushTypedTestCard(g, Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})

	rows := ManaAbilitiesForCard(layeredBattlefieldCard(t, g, taxed))
	if len(rows) != 2 || rows[0].LifeCost != 2 || rows[1].LifeCost != 1 {
		t.Fatalf("rows %+v, want the carried row at 2 life and the intrinsic {G} at 1", rows)
	}
	for _, ab := range ManaAbilitiesForCard(layeredBattlefieldCard(t, g, plain)) {
		if ab.LifeCost != 0 {
			t.Errorf("an untaxed Forest's row costs %d life", ab.LifeCost)
		}
	}
	// The shared intrinsic row was copied, not written to.
	for _, shared := range intrinsicLandRows {
		for _, ab := range shared {
			if ab.LifeCost != 0 {
				t.Fatalf("the shared intrinsic row %q now costs %d life", ab.Label, ab.LifeCost)
			}
		}
	}

	life := me.Life
	if err := g.ActivateManaAbility(me.ID, taxed, 1, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life-1 || len(me.ManaPool) != 1 {
		t.Errorf("life %d (want %d), pool %+v", me.Life, life-1, me.ManaPool)
	}
}

// CR 119.4: checked before anything is paid.
func TestManaAbilityLifeTaxRefusesAPlayerWhoCannotPay(t *testing.T) {
	withManaLifeTax(t, 1, "")
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTypedTestCard(g, Card{Name: "Taxed Forest", TypeLine: "Land — Forest", OracleID: lifeTaxOracle, Owner: me.ID, Controller: me.ID})
	me.Life = 0
	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("ActivateManaAbility at 0 life: %v, want ErrInvalidParam", err)
	}
	if c := layeredBattlefieldCard(t, g, id); c.Tapped || len(me.ManaPool) != 0 {
		t.Error("a refused activation tapped the land or made mana")
	}
}

// CR 613.1f: the surcharge is one of the permanent's own abilities, so a
// removal takes it whichever of the two has the later timestamp.
func TestManaAbilityLifeTaxGoesWithThePermanentsAbilities(t *testing.T) {
	const remover = "test-remover"
	for _, removerFirst := range []bool{true, false} {
		withManaLifeTax(t, 1, remover)
		g := newActiveGame(t)
		me := g.Seats[0]
		land := Card{Name: "Taxed Forest", TypeLine: "Land — Forest", OracleID: lifeTaxOracle, Owner: me.ID, Controller: me.ID}
		rem := Card{Name: "Remover", TypeLine: "Enchantment", OracleID: remover, Owner: me.ID, Controller: me.ID}
		var id uuid.UUID
		if removerFirst {
			pushTypedTestCard(g, rem)
			id = pushTypedTestCard(g, land)
		} else {
			id = pushTypedTestCard(g, land)
			pushTypedTestCard(g, rem)
		}
		rows := ManaAbilitiesForCard(layeredBattlefieldCard(t, g, id))
		if len(rows) != 1 || rows[0].LifeCost != 0 {
			t.Errorf("remover first %v: rows %+v, want the intrinsic {G} for no life", removerFirst, rows)
		}
	}
}
