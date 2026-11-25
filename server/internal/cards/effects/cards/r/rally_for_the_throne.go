package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally for the Throne — Instant {2}{W} (Throne of Eldraine, #5516):
//
//	"Create two 1/1 white Human creature tokens.
//	 Adamant — If at least three white mana was spent to cast this spell, you
//	 gain 1 life for each creature you control."
//
// The tokens first, then the adamant rider counts the creatures the caster
// controls as it resolves, the new Humans included: a creature that is not
// there does not count (CR 608.2c). A waived payment reads as not spent
// (AdamantSpent), the weaker-than-printed direction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0a557383-596b-4b8f-b81f-21706fef7b63",
		Name:           "Rally for the Throne",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W"},
		Purpose:        game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 2}.Apply(ctx)
			}),
			effects.EldB1IfAdamantSpent("W", effects.OldB140GainLifePerCreature{}),
		),
	})
}
