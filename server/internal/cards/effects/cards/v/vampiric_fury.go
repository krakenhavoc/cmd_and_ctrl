package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Fury — Instant {1}{R} (Foundations Jumpstart, #5929):
//
//	"Vampire creatures you control get +2/+0 and gain first strike until
//	 end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e23457f-1ed9-4597-9f73-6f5b7206f5a1",
		Name:         "Vampiric Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			vampires := effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Vampire"))
			if err := (effects.BoostUntilEOT{Match: vampires, Power: 2, Label: "Vampiric Fury — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: vampires, Keywords: []string{"first strike"}, Label: "Vampiric Fury — first strike"}.Apply(ctx)
		},
	})
}
