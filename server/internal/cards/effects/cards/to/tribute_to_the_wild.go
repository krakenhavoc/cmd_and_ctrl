package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tribute to the Wild — Instant {1}{G} (Commander 2020, S82 OLD, #5513):
//
//	"Each opponent sacrifices an artifact or enchantment of their choice."
//
// The Eldest Reborn's edict over artifacts and enchantments: every opponent
// in turn order chooses one of their own, so a player with none loses
// nothing and a hexproof permanent is still chosen (it does not target).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba0fc2ca-7084-454a-9d4e-5cbca945a7ca",
		Name:         "Tribute to the Wild",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Or(effects.Artifact(), effects.Enchantment()),
				Label:            "an artifact or enchantment",
			}.Apply(ctx)
		},
	})
}
