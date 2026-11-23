package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Purifier — Creature — Vampire Cleric {1}{W}{B}, 2/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"Lifelink
//	 At the beginning of your end step, if you gained life this turn, you
//	 may pay {2}. If you do, draw a card."
//
// Lifelink is printed. The trigger has an intervening "if" (CR 603.4):
// CondLifeGained reads the tally of life you gained this turn (lifelink
// included, a gain and a loss do not cancel) as the end step begins and
// again as the trigger resolves. The payment is MayPay's, asked at
// resolution, and the draw is in the paid branch so a declined or unpayable
// {2} draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d28eac97-f38a-4d2a-aeae-e758bd3ecf4e",
		Name:            "Markov Purifier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondLifeGained{},
				effects.AtYourEndStep("Markov Purifier — you may pay {2}; if you do, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayPay{
							Chooser:  item.Controller,
							Cost:     "{2}",
							Question: "Markov Purifier — pay {2} to draw a card?",
							Then:     markovPurifierPaid,
						}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}

// markovPurifierPaid is the paid half, keyed (#4264). Never renamed.
var markovPurifierPaid = effects.PayUnlessBranches("pay-unless/markov-purifier-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.DrawCards{N: 1}.Apply(ctx) }, nil)
