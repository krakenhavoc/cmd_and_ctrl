package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobgoblinLordDamage activates the Lord at the opponent and returns
// the life it cost them.
func HobgoblinLordDamage(t *testing.T, g *game.Game, me *game.Player, opp *game.Player, lord uuid.UUID) int {
	t.Helper()
	before := opp.Life
	B06AddMana(me, "R")
	B16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: B16TargetPlayer(opp.ID)})
	return before - opp.Life
}
