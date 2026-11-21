package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gild — Sorcery {3}{B} (Born of the Gods,
// #5559):
//
//	"Exile target creature. Create a Gold token. (It's an artifact with
//	 "Sacrifice this token: Add one mana of any color.")"
//
// The exile, then the Gold token (the catalog's predefined token,
// tokens.go). A target that has become illegal counters the whole
// spell, token included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5530fac-0bae-451d-921d-0dd4865ec91d",
		Name:         "Gild",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.GoldToken(), N: 1}.Apply(ctx)
		},
	})
}
