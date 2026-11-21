package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You See a Pair of Goblins — Instant {2}{R} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Charge Them — Creatures you control get +2/+0 until end of turn.
//	 • Befriend Them — Create two 1/1 red Goblin creature tokens."
//
// Ambitious Assault's mass pump over the creatures you control as the spell
// resolves, or Dragon Fodder's two Goblins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77d016e3-f652-4827-b49a-8044f6d2d45b",
		Name:         "You See a Pair of Goblins",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Charge Them — Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "You See a Pair of Goblins — Charge Them"}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			effects.ModeWithPurpose(effects.ModeDoing("Befriend Them — Create two 1/1 red Goblin creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.RedGoblinToken(), N: 2}.Apply(ctx)
				}), game.Purpose{Tokens: 2, CreatureTokens: 2}),
		),
	})
}
