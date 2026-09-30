package effects

import (
	"testing"

	"github.com/google/uuid"
)

// TestEnduringGlimmersReturnAsNonCreatureEnchantmentsOnce pins the
// shared dies trigger on every catalogued Enduring Glimmer: the first
// death returns the card under its owner's control as an enchantment
// that is not a creature; the second death (as an enchantment) does
// not return it.
func TestEnduringGlimmersReturnAsNonCreatureEnchantmentsOnce(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
	}{
		{"Enduring Vitality", "Enchantment Creature — Elk Glimmer", enduringVitalityOracle},
		{"Enduring Innocence", "Enchantment Creature — Sheep Glimmer", "98a389f4-2905-47f3-b60e-3d4afb3e5cb0"},
		{"Enduring Tenacity", "Enchantment Creature — Snake Glimmer", "98e698ae-1a69-469c-9cfb-0e3fedeb71d4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)

			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			passPriorityAroundTable(t, g)

			if !g.Battlefield.Contains(id) {
				t.Fatal("should be back on the battlefield after dying as a creature")
			}
			var isCreature, isEnchantment bool
			var owner, controller uuid.UUID
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.InstanceID == id {
						isCreature, isEnchantment = c.IsCreature(), c.IsEnchantment()
						owner, controller = c.Owner, c.Controller
					}
				}
			})
			if isCreature {
				t.Error("returned card must not be a creature")
			}
			if !isEnchantment {
				t.Error("returned card must be an enchantment")
			}
			if owner != me.ID || controller != me.ID {
				t.Errorf("owner/controller = %s/%s, want %s", owner, controller, me.ID)
			}

			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(id) {
				t.Error("dying a second time as an enchantment must not return it again")
			}
		})
	}
}

// TestEnduringVitalityStopsGrantingManaOnceItIsNotACreature: the
// returned enchantment is no longer a creature, so "creatures you
// control have {T}: add mana" no longer covers it.
func TestEnduringVitalityStopsGrantingManaOnceItIsNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Enduring Vitality", "Enchantment Creature — Elk Glimmer", enduringVitalityOracle, false)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Enduring Vitality should have returned")
	}
	var granted int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				granted = len(c.Effective().GrantedAbilities)
			}
		}
	})
	if granted != 0 {
		t.Errorf("returned enchantment still has %d granted abilities, want 0", granted)
	}
}
