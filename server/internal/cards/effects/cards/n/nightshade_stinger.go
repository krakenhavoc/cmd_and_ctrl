package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightshade Stinger — Creature — Faerie Rogue {B}, 1/1 (Jumpstart, #5510):
//
//	"Flying
//	 This creature can't block."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c57edca9-b71e-4ec1-84d5-4b178969ccb7",
		Name:            "Nightshade Stinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
