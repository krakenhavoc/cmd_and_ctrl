package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// UpToCastXCreatureTargets is the TargetsFrom of an entering trigger
// whose clause is "up to X target creatures", X being the X the permanent
// was cast with (Card.CastX, CR 107.3m): Blitzball Stadium's support and
// Extraordinary Journey's exile.
//
// The clause is built when the trigger goes on the stack, from the
// source's cast provenance, so a permanent that was not cast (blinked,
// reanimated, put onto the battlefield) has X of zero and gets no clause
// at all — nothing to target, nothing to ask.
func UpToCastXCreatureTargets(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
	x := source.CastX()
	if x <= 0 {
		return nil
	}
	return TargetCreature("up to X target creatures").WithCount(0, x)
}
