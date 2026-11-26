package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungal Sprouting — Sorcery {3}{G} (Magic 2013, S82 OLD, #5568):
//
//	"Create X 1/1 green Saproling creature tokens, where X is the greatest
//	 power among creatures you control."
//
// X is read as the spell resolves, through the layers (CR 608.2h, CR 613.4):
// with no creature, or only creatures of power 0, no token is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "303d4533-0c6b-4f93-80e8-b1f5926f4cd4",
		Name:         "Fungal Sprouting",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.S82OldB430SaprolingsForGreatestPower,
	})
}
