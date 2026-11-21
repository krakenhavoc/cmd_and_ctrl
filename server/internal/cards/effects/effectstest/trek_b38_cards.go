package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared test helpers ------------------------------------------------------

// trekB38Named counts the battlefield permanents named `name` that
// `controller` controls, and how many of them are tapped.
func TrekB38Named(g *game.Game, controller uuid.UUID, name string) (n, tapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	return n, tapped
}

// trekB38Activate floats `mana` and activates ability `idx` strictly.
func TrekB38Activate(t *testing.T, g *game.Game, who, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		g.WithWriteLock(func() {
			if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
				t.Fatalf("AddManaForEffect(%s): %v", mana, err)
			}
		})
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who, id, idx, params)
}

// trekB38Attack declares the attackers at `opp`, locks the attack in and
// lets the attack triggers resolve.
func TrekB38Attack(t *testing.T, g *game.Game, opp uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
}
