package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ojutai's Summons — Sorcery {3}{U}{U}:
//
//	"Create a 2/2 blue Djinn Monk creature token with flying.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "def45f3a-dba0-4d08-b086-5236d6e0edb1",
		Name:            "Ojutai's Summons",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 blue Djinn Monk with flying"), N: 1}.Apply(ctx)
		},
	})
}
