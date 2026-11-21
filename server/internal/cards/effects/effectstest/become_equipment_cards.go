package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// become_equipment_cards_test.go — #2562: a permanent that becomes an
// Equipment, or is given equip, by another effect (ADR 0093 amendment
// 2026-10-10). The Irencrag through a resolved effect that renames it,
// Gemcutter Buccaneer and Puresteel Paladin through statics. Every test
// drives the real trigger, activation and layer paths and asserts what a
// player sees: the name, the type line, the rows, the equipped creature's
// size.

// beGrantedRow is the index and ref of the granted activated row with
// this label on a permanent, or -1.
func BeGrantedRow(t *testing.T, g *game.Game, id uuid.UUID, label string) (int, string) {
	t.Helper()
	rows, origins := game.ActivatedAbilitiesWithOrigins(GaLayered(t, g, id))
	for i, o := range origins {
		if o.Granted() && rows[i].Label == label {
			return i, o.Ref
		}
	}
	return -1, ""
}

// beEquip activates a granted equip row onto `host` and lets it resolve.
func BeEquip(t *testing.T, g *game.Game, controller, equipment, host uuid.UUID, label string) {
	t.Helper()
	idx, ref := BeGrantedRow(t, g, equipment, label)
	if idx < 0 {
		t.Fatalf("%s has no granted %q row", equipment, label)
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	PassPriorityAroundTable(t, g)
}
