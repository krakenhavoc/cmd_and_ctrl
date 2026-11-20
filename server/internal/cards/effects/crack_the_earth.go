package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "eb107601-f4ff-4504-9e3f-3de63b0d9e6b",
		Name:         "Crack the Earth",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{Label: "a permanent"}.Apply(ctx)
		},
	})
}
