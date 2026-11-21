package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transluminant — Creature — Dryad Shaman {1}{G}, 2/2 (Ravnica: Clue
// Edition, S82):
//
//	"{W}, Sacrifice this creature: Create a 1/1 white Spirit creature token
//	 with flying at the beginning of the next end step."
//
// The sacrifice is a cost; the token is a delayed trigger (CR 603.7) kept on
// the game, controlled by the ability's controller, so it still fires with
// the Dryad long gone. Scheduled during an end step it waits for the next
// one (Mishra's Bauble's "next" bookkeeping).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4d54a89-8409-4fbc-b01f-3b03f352820f",
		Name:         "Transluminant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{W}, Sacrifice this creature: Create a 1/1 white Spirit creature token with flying at the beginning of the next end step.",
			Cost:  effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			// Sacrificing itself at will is the sacrifice-outlet answer.
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ScheduleDelayedTrigger{
					Label: "Transluminant — create a 1/1 white Spirit creature token with flying",
					Body:  transluminantSpiritBody,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// transluminantSpiritBody is the delayed trigger: one Spirit for the
// ability's controller. An on-disk key: never renamed, never reused.
var transluminantSpiritBody = game.SimpleDelayedBody("transluminant/spirit-token", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("1/1 white Spirit with flying")), N: 1}.Apply(ctx)
})
