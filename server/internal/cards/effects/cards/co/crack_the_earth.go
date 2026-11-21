package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crack the Earth — Sorcery — Arcane {R} (Betrayers of Kamigawa, #5602):
//
//	"Each player sacrifices a permanent of their choice."
//
// An edict over any permanent: each player, the caster included, picks
// one of their own (EachPlayerSacrifices with no Match), in APNAP order.
// A player with no permanents sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb107601-f4ff-4504-9e3f-3de63b0d9e6b",
		Name:         "Crack the Earth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{Label: "a permanent"}.Apply(ctx)
		},
	})
}
