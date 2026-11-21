package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tribute to Hunger — Instant {2}{B} (Foundations, #4465):
//
//	"Target opponent sacrifices a creature of their choice. You gain life
//	 equal to that creature's toughness."
//
// Consuming Vapors' edict (edictGainToughness) aimed at an opponent: the
// life waits for the sacrifice and reads the creature's toughness as it
// last existed on the battlefield (CR 608.2h, the ruling). A player with
// no creature sacrifices nothing and you gain nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc5a501c-e16c-4d4b-a510-70a14272473e",
		Name:         "Tribute to Hunger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EdictGainToughness(ctx, "Tribute to Hunger")
		},
	})
}
