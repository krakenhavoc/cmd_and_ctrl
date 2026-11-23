package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pip_5469.go — Fallout batch 1 (#5469) helpers its tests and the S82 GAP tests share.

const (
	Pipb1StrengthBobbleheadOracle = "bd0e22d1-4789-406d-9293-3ade30365b32"
	Pipb1CharismaBobbleheadOracle = "77897962-7d67-4fa8-858f-2f35a322d7c6"
)

// Pipb1Bobblehead puts a catalog Bobblehead onto the battlefield.
func Pipb1Bobblehead(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return PushCatalogPermanent(g, owner, name, "Artifact — Bobblehead", oracle, false)
}

// Pipb1Activate activates a catalog permanent's own ability with the
// mana gate on.
func Pipb1Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, index int, mana string, params game.ActivateAbilityParams) {
	t.Helper()
	if mana != "" {
		FloatMana(t, g, p, mana)
	}
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, id, index, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
}
