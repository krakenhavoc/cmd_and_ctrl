package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ApaActivateNoPoolCheck has p activate row index of source (strict) and
// passes priority around the table until the stack is empty or a choice is
// owed.
func ApaActivateNoPoolCheck(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(p.ID, source, index, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("%s activates row %d: %v", p.Name, index, err)
	}
	PassPriorityAroundTable(t, g)
}
