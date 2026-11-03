package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// plusOneCountersOnACreatureYouControl is the condition of "if one or more
// +1/+1 counters would be put on a creature you control" (CR 614.1a):
// a placement of at least one +1/+1 counter on a creature the
// replacement's controller controls, the source itself included.
// Shang-Chi, Martial Mentor doubles the count; Michelangelo, Weirdness to
// 11 adds one.
func plusOneCountersOnACreatureYouControl(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventCounter || ev.CounterName != game.CounterPlusOne || ev.CounterDelta <= 0 {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	return ok && target.IsCreature() && target.Controller == src.Controller
}
