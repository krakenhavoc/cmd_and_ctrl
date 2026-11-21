package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flip the Switch — Instant {2}{U} (Midnight Hunt):
//
//	"Counter target spell unless its controller pays {4}. Create a 2/2
//	 black Zombie creature token with decayed."
//
// The counter is Complicate's counterTheTargetSpellUnlessPaid: the
// spell's controller is asked as this resolves and the token waits for
// the answer (ResolveSteps, CR 608.2c). The token is made whether or
// not they paid. A spell that is gone leaves this with no legal target,
// so it does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a28eb946-c7f2-4c90-ab6c-b194e290e33d",
		Name:         "Flip the Switch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CounterTheTargetSpellUnlessPaid("Flip the Switch", "{4}")(ctx.Game, ctx.Item)
			}),
			effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1},
		),
	})
}
