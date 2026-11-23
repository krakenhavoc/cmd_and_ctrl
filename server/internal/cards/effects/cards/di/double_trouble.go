package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Double Trouble — {4}{R} Instant (Marvel's Spider-Man Eternal, #4460):
//
//	"Double the power of each creature you control until end of turn."
//
// CR 701.10b: each creature gets +X/+0 with X its power as the spell
// resolves. Every X is read before any is applied (doubling.go), and the set
// is the creatures you control now (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ef063d2-8702-43a3-a7d0-fc255f8512d3",
		Name:         "Double Trouble",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var mine []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() {
					mine = append(mine, c.InstanceID)
				}
			}
			return effects.DoublePowerOfEachUntilEOT(ctx, mine, "Double Trouble — double power")
		},
	})
}
