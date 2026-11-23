package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Surprise — Instant {2}{R} (Foundations, #4465):
//
//	"Choose one —
//	 • Creatures you control get +2/+0 until end of turn.
//	 • Create two 1/1 red Goblin creature tokens."
//
// Mode one fixes its set of creatures as the spell resolves (CR 611.2c),
// so a creature that enters later is not pumped (Charge the Sanctum's
// first mode). Mode two makes two of Krenko's Goblins, which count for
// anything that counts Goblins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79cd09e8-d738-47c0-9473-b6281bec6935",
		Name:         "Goblin Surprise",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Creatures you control get +2/+0 until end of turn."),
			effects.ModeWithPurpose(effects.Mode("Create two 1/1 red Goblin creature tokens."), game.Purpose{Tokens: 2, CreatureTokens: 2}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 2,
					Label: "Goblin Surprise — +2/+0",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.CreateToken{Controller: item.Controller, Template: effects.RedGoblinToken(), N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
