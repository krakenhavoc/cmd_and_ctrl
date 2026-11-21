package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S82b11Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, index int, targets ...uuid.UUID) error {
	t.Helper()
	var refs []game.TargetRef
	for _, tg := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tg})
	}
	err := g.ActivateCatalogAbility(me.ID, id, index, game.ActivateAbilityParams{Strict: true, Targets: refs})
	if err == nil {
		PassPriorityAroundTable(t, g)
	}
	return err
}
