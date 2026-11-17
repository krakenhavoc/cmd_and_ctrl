package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for the S82 ONE batch 8 cards (Phyrexia: All Will Be One).

// s82OneB08Corrupted is corrupted as a cost predicate (Distorted
// Curiosity): an opponent of the caster has three or more poison counters.
func s82OneB08Corrupted(q game.CostQuery) bool {
	return rfReprintBCorrupted(q.Game, q.Controller)
}

// s82OneB08ManaValueAtMost reads, before anything happens to it, whether the
// object `id` (a spell on the stack, or a permanent) has mana value `limit`
// or less. "That spell's / permanent's mana value was 3 or less" is read as
// last known (CR 608.2h), so it must be read before the counter, bounce or
// destruction. An X spell counts its X on the stack and a permanent's X is 0
// (CR 202.3e). False when the object cannot be found.
func s82OneB08ManaValueAtMost(ctx *Context, id uuid.UUID, limit int) bool {
	card, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return false
	}
	mv, _ := ctx.Game.ManaValueForEffect(card)
	return mv <= limit
}

// s82OneB08ExpandTheSphereRest is the rest of Expand the Sphere after the
// lands are chosen: the unchosen cards go on the bottom in a random order,
// then proliferate once for each land fewer than two that was put onto the
// battlefield.
var s82OneB08ExpandTheSphereRest = PutFromLibraryTail("s82-one-b08/expand-the-sphere-rest",
	func(ctx *Context, res PutFromLibraryResult) error {
		if err := PutRestOnBottomInRandomOrder(ctx.Game, res); err != nil {
			return err
		}
		return proliferateTimesKeyed(ctx.Game, res.Player, res.Source, 2-len(res.Entered))
	})
