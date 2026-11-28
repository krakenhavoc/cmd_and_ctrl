package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hornet Cannon — Artifact {4} (Stronghold, S82 OLD batch 746, #5627):
//
//	"{3}, {T}: Create a 1/1 colorless Insect artifact creature token with
//	 flying and haste named Hornet. Destroy it at the beginning of the next
//	 end step."
//
// The token is the printed Hornet (a new template row). The tokens that
// actually landed are handed to the end-step destroy through the creation's
// keyed tail (OldB746DestroyTokensAtNextEndStep, CR 603.7), so a token a
// replacement kept off the battlefield is simply not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "595ae6f1-e66a-4774-943a-9d5368f4a3b4",
		Name:         "Hornet Cannon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a 1/1 colorless Insect artifact creature token with flying and haste named Hornet. Destroy it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: ctx.Controller(),
					Source:     ctx.Source(),
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 colorless Hornet Insect artifact with flying and haste")), Count: 1}},
				}, ctx.ItemTail(hornetCannonDestroy))
			},
		}},
	})
}

// hornetCannonDestroy is "destroy it at the beginning of the next end step".
// An on-disk key.
var hornetCannonDestroy = effects.OldB746DestroyTokensAtNextEndStep("hornet-cannon/destroy-the-hornet",
	"Hornet Cannon — destroy the Hornet")
