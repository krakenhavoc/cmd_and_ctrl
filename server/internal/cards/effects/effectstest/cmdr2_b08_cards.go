package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B08Damage deals noncombat damage and lets the batch settle, so the
// "whenever this is dealt damage" triggers go on the stack.
func Cmdr2B08Damage(g *game.Game, source, target uuid.UUID, n int) {
	B27Damage(g, source, target, n)
	g.RunStateChecksForTest()
}
