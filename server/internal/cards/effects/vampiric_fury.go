package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Fury — Instant {1}{R} (Foundations Jumpstart, #5929):
//
//	"Vampire creatures you control get +2/+0 and gain first strike until
//	 end of turn."
func init() {
	Register(Spec{
		OracleID:     "1e23457f-1ed9-4597-9f73-6f5b7206f5a1",
		Name:         "Vampiric Fury",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			vampires := And(Creature(), YouControl(), HasSubtype("Vampire"))
			if err := (BoostUntilEOT{Match: vampires, Power: 2, Label: "Vampiric Fury — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Match: vampires, Keywords: []string{"first strike"}, Label: "Vampiric Fury — first strike"}.Apply(ctx)
		},
	})
}
