package game

import (
	"errors"
	"testing"
)

// activation_tap_test.go — S82 OLD b825 (#5635): ActivationAbility.Tap, the bit
// Serra Bestiary's "activated abilities with {T} in their costs can't be
// activated" reads. It is derived from the cost of the shape being activated,
// on the non-mana path (ActivationAbilityOf) and the mana path
// (ManaActivationAbilityOf) alike.

// tapCostRule is the restriction a Serra Bestiary contributes, written
// against the query alone: any battlefield ability whose cost has {T}.
func tapCostRule() ActivationRestriction {
	return ActivationRestriction{
		Label:   "Activated abilities with {T} in their costs can't be activated.",
		Forbids: func(q ActivationQuery) bool { return q.FromZone == ZoneBattlefield && q.Ability.Tap },
	}
}

func TestActivationAbilityTapFollowsTheCost(t *testing.T) {
	if !ActivationAbilityOf(ActivatedAbilityShape{Cost: AbilityCost{Tap: true}}).Tap {
		t.Error("a {T} cost did not set Tap")
	}
	if ActivationAbilityOf(ActivatedAbilityShape{Cost: AbilityCost{}}).Tap {
		t.Error("a cost without {T} set Tap")
	}
	if got := ManaActivationAbilityOf(ManaAbilityShape{TapCost: true, Label: "x"}); !got.Tap || !got.Mana || got.Label != "x" {
		t.Errorf("tap mana ability = %+v, want Tap and Mana", got)
	}
	if got := ManaActivationAbilityOf(ManaAbilityShape{Label: "x"}); got.Tap || !got.Mana {
		t.Errorf("non-tap mana ability = %+v, want Mana without Tap", got)
	}
}

func TestATapCostRestrictionStopsTapAbilitiesAndTapManaAbilitiesOnly(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	stubActivationRestrictions(t, totemOracle, []ActivationRestriction{tapCostRule()})
	stubActivatedFor(t, elfOracle,
		ActivatedAbilityShape{
			Label:  "{T}: Draw a card.",
			Cost:   AbilityCost{Tap: true},
			Effect: func(*Game, *StackItem) error { return nil },
		},
		ActivatedAbilityShape{
			Label:  "Free: Do nothing.",
			Cost:   AbilityCost{},
			Effect: func(*Game, *StackItem) error { return nil },
		})
	withCatalogHook(t, func(id string) []ManaAbilityShape {
		if id == elfOracle {
			return []ManaAbilityShape{
				{TapCost: true, Produced: "{G}", Label: "{T}: Add {G}"},
				{TapCost: false, ManaCost: "{1}", Produced: "{G}{G}", Label: "{1}: Add {G}{G}"},
			}
		}
		return nil
	})

	elf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)

	// Unrestricted: all of them work, so the refusals below are the rule's.
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability before the lock: %v", err)
	}
	me.ManaPool = nil
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].Tapped = false
		}
	})

	pushGateCard(g, "Lock Source", "Enchantment", totemOracle, me.ID)

	// A {T} ability that is not a mana ability.
	err := g.ActivateCatalogAbility(me.ID, elf, 0, ActivateAbilityParams{})
	var cant *CantActivateError
	if !errors.As(err, &cant) || cant.Reason != "Activated abilities with {T} in their costs can't be activated." {
		t.Fatalf("{T} ability: err = %v, want the printed clause", err)
	}
	// A {T} mana ability (a Forest's shape).
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Fatalf("{T} mana ability: err = %v, want ErrCantActivate", err)
	}
	// A non-{T} ability of each kind is untouched.
	if err := g.ActivateCatalogAbility(me.ID, elf, 1, ActivateAbilityParams{}); err != nil {
		t.Fatalf("non-{T} ability: %v", err)
	}
}
