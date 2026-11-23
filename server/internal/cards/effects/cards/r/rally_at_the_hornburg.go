package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally at the Hornburg — Sorcery {1}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Create two 1/1 white Human Soldier creature tokens. Humans you control
//	 gain haste until end of turn."
//
// The tokens are made first, so the haste step, which fixes its set of Humans
// as it resolves (CR 611.2c), includes them. "Human" is read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d3f542a-de64-4de8-9de0-0e25004467f1",
		Name:         "Rally at the Hornburg",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Human")),
					Keywords: []string{"haste"},
					Label:    "Rally at the Hornburg — Humans gain haste",
				}.Apply(ctx)
			}),
		),
	})
}
