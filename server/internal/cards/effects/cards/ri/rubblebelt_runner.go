package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rubblebelt Runner — Creature — Lizard Warrior {1}{R}{G}, 3/3 (Ravnica
// Allegiance):
//
//	"This creature can't be blocked by creature tokens."
//
// Duskwatch Hunter's block rule: a token can't be assigned to block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bfa51dd-daed-473a-8377-7bb34a2c3371",
		Name:         "Rubblebelt Runner",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.IsTokenPredicate(), "creature tokens"),
		},
	})
}
