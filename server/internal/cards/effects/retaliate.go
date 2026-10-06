package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retaliate — Instant {2}{W}{W}:
//
//	"Destroy all creatures that dealt damage to you this turn."
//
// The set is read as the spell resolves from the per-turn record of
// which creature objects dealt damage to its controller (#2149), so a
// creature that dealt damage and was then flickered is a new object
// and is spared (CR 400.7), and a creature that never damaged you
// survives whatever else it did. Regeneration is allowed: the card
// does not say otherwise.
func init() {
	Register(Spec{
		OracleID:     "37f5d022-2989-4b44-9bca-dc6b606afe10",
		Name:         "Retaliate",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), dealtDamageToPlayerThisTurn(item.Controller))}.Apply(ctx)
		},
	})
}
