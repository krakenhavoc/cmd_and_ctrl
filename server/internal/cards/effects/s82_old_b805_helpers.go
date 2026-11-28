package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b805_helpers.go — S82 set OLD, batch 805 (Mirage, counters).

// OldB805InterveningIfSource narrows a trigger to an intervening "if" (CR
// 603.4) about its own source and controller: it triggers only while `test`
// holds for them as the event happens, and does nothing if it has stopped
// holding as the ability resolves. `name` words the condition for the
// footprint. (InterveningIfYou is the same for a condition about the
// controller alone.)
func OldB805InterveningIfSource(name string, test func(g *game.Game, source, controller uuid.UUID) bool, t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && test(g, source.InstanceID, source.Controller) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: name, Test: func(ctx *Context) bool {
		return test(ctx.Game, ctx.Source(), ctx.Controller())
	}}, t)
}
