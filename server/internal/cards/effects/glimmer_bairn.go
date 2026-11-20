package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmer Bairn — Creature — Ouphe {G}, 1/2 (Modern Horizons 2, #5501):
//
//	"Sacrifice a token: This creature gets +2/+2 until end of turn."
//
// Atog's sacrifice pump over a token (b08SacrificeAToken, Combine
// Chrysalis's cost): any token you control pays, creature or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "118cdf8b-44fb-4c3e-86c4-e7f4b1267646",
		Name:         "Glimmer Bairn",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a token: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    b08SacrificeAToken(),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Glimmer Bairn — +2/+2 until end of turn"),
		}},
	})
}
