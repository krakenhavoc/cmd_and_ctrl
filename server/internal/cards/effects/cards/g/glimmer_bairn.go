package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmer Bairn — Creature — Ouphe {G}, 1/2 (Modern Horizons 2, #5501):
//
//	"Sacrifice a token: This creature gets +2/+2 until end of turn."
//
// Atog's sacrifice pump over a token (b08SacrificeAToken, Combine
// Chrysalis's cost): any token you control pays, creature or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "118cdf8b-44fb-4c3e-86c4-e7f4b1267646",
		Name:         "Glimmer Bairn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a token: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.B08SacrificeAToken(),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Glimmer Bairn — +2/+2 until end of turn"),
		}},
	})
}
