package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decode Transmissions — Sorcery {2}{B} (Edge of Eternities):
//
//	"You draw two cards and lose 2 life.
//	 Void — If a nonland permanent left the battlefield this turn or a
//	 spell was warped this turn, instead you draw two cards and each
//	 opponent loses 2 life."
//
// "Instead" swaps the second half only: you draw two cards either way, and
// void (void.go), read as the spell resolves, decides whether the loss is
// yours or every opponent's. Two steps, so the loss waits for a prompt the
// draw asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42510758-6167-420b-bb70-65f914b7a259",
		Name:         "Decode Transmissions",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.Void(ctx.Game) {
					return effects.EachOpponentLosesLife{N: 2}.Apply(ctx)
				}
				return effects.YouLoseLife(2)(ctx)
			}),
		),
	})
}
