package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ovinize — Instant {1}{U} (Dominaria Remastered, #5948):
//
//	"Until end of turn, target creature loses all abilities and has base
//	 power and toughness 0/1."
//
// One continuous effect: layer 6 removes the abilities, layer 7b sets the
// base power and toughness (CR 613.1), so counters and later pumps still add.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98799336-a61e-4e35-ac30-51edb81301f6",
		Name:         "Ovinize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.LosesAbilitiesAndBasePTUntilEOT(ctx, "Ovinize", 0, 1)
		},
	})
}
