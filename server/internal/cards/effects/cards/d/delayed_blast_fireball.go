package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delayed Blast Fireball — Instant {1}{R}{R} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Delayed Blast Fireball deals 2 damage to each opponent and each creature
//	 they control. If this spell was cast from exile, it deals 5 damage to each
//	 opponent and each creature they control instead.
//	 Foretell {4}{R}{R}"
//
// Tectonic Hazard's one-sided sweep at 2, or at 5 when the spell's cast zone
// (stamped at announce, CR 601.2a) is exile: a foretold card is cast from
// exile, and so is one cast from exile by any other permission.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "99915b5a-5092-4900-8824-6de2aab1b2c1",
		Name:           "Delayed Blast Fireball",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{4}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if item.CastFromZone == game.ZoneExile {
				n = 5
			}
			return effects.DamageEachOpponentAndEachCreatureTheyControl(ctx, n)
		},
	})
}
