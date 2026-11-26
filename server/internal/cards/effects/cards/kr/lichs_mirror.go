package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lich's Mirror — Artifact {5} (Shards of Alara, S82 OLD batch 509, #5588):
//
//	"If you would lose the game, instead shuffle your hand, your graveyard,
//	 and all permanents you own into your library, then draw seven cards and
//	 your life total becomes 20."
//
// game.ReplaceLoseGameShuffleAllDrawSevenLifeBecomes, read where every loss
// replacement is (game/loss_shuffle_all.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e81da4a7-4b03-4107-b7d9-4b635faa615d",
		Name:         "Lich's Mirror",
		Completeness: effects.CompletenessFull,
		GameActionReplacements: []game.GameActionReplacement{
			{Kind: game.ReplaceLoseGameShuffleAllDrawSevenLifeBecomes, Life: 20},
		},
	})
}
