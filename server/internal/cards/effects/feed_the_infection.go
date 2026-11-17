package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feed the Infection — Sorcery {3}{B} (Phyrexia: All Will Be One, #5479):
//
//	"You draw three cards and you lose 3 life.
//	 Corrupted — Each opponent who has three or more poison counters loses 3
//	 life."
//
// Three steps in printed order, so each waits for a prompt the one before
// asked (CR 608.2c). Corrupted here is per opponent: each opponent with
// three or more poison counters loses 3 life, read as the step runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "405a2cf9-62a0-460c-bf2a-ce4c351b61ce",
		Name:         "Feed the Infection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Controller(), Amount: -3}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				for _, opp := range ctx.Opponents() {
					p := ctx.Game.PlayerByIDForEffect(opp)
					if p == nil || p.Poison < 3 {
						continue
					}
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -3); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
