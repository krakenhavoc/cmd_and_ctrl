package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Path of Anger's Flame — Instant — Arcane {2}{R} (Saviors of Kamigawa,
// #5601):
//
//	"Creatures you control get +2/+0 until end of turn."
//
// The creatures are fixed as the spell resolves (CR 611.2c): one that
// enters later in the turn does not get the bonus. It ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e89db98-fc9f-4574-b5bc-ee421fa29428",
		Name:         "Path of Anger's Flame",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2,
				Label: "Path of Anger's Flame — +2/+0",
			}.Apply(ctx)
		},
	})
}
