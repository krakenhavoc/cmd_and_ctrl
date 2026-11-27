package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trenching Steed — Creature — Horse Rebel {3}{W}, 2/3 (Prophecy, S82 OLD batch
// 674, #5617):
//
//	"Sacrifice a land: This creature gets +0/+3 until end of turn."
//
// The sacrifice is the cost (Spitting Spider's B08SacrificeALand); the ability
// resolves into Ridgeline Rager's self pump with the other number. It may be
// activated any number of times, each paying its own land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31136103-535a-4d11-819a-1cbd57ad48a4",
		Name:         "Trenching Steed",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: This creature gets +0/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 3}},
			Cost:    effects.B08SacrificeALand(),
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 3, "Trenching Steed — +0/+3"),
		}},
	})
}
