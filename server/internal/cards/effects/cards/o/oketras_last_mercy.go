package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oketra's Last Mercy — Sorcery {1}{W}{W}:
//
//	"Your life total becomes equal to your starting life total. Lands
//	 you control don't untap during your next untap step."
//
// The life total becomes the table's starting total (40 in Commander)
// by gaining or losing the difference (CR 119.5), so "can't gain life"
// stops the gain and a life-gain payoff sees it. Then #3995's
// player-wide skip over the caster's lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5e0aa12-920d-4700-a287-5ca24933abb1",
		Name:         "Oketra's Last Mercy",
		Completeness: effects.CompletenessFull,
		// The lands' untap rider waits for the life change's prompt (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B31LifeBecomes(ctx, ctx.Controller(), effects.StartingLifeOf(ctx.Game))
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.LandsYouControlDontUntapNextStep(ctx)
			}),
		),
	})
}
