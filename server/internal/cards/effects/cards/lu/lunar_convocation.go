package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunar Convocation — Enchantment {W}{B} (EDHREC rank 3153):
//
//	"At the beginning of your end step, if you gained life this
//	 turn, each opponent loses 1 life.
//	 At the beginning of your end step, if you gained and lost life
//	 this turn, create a 1/1 black Bat creature token with flying.
//	 {1}{B}, Pay 2 life: Draw a card."
//
// The Orzhov lifegain-and-loss payoff. Two separate end-step
// triggers, each with an intervening-if (CR 603.4) checked when the
// end step begins AND again as the trigger resolves, read off the
// per-turn tally: a positive life change is a
// gain, a negative one — or damage to the player — a loss. Paying
// the draw ability's 2 life is losing life (CR 119.4), so the card
// enables its own Bat, as printed. When both fire the controller
// orders them (CR 603.3b), like any two differing triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb68d1e4-36aa-4a00-b671-8959e7d526d4",
		Name:         "Lunar Convocation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, "Lunar Convocation — each opponent loses 1 life", effects.B30LoseOneIfYouGainedLifeThisTurn),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30YourEndStepAndYouGainedAndLostLifeThisTurn(ev, source, g)
			}, "Lunar Convocation — create a 1/1 Bat with flying", effects.B30BatIfYouGainedAndLostLifeThisTurn),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Pay 2 life: Draw a card",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.PayLife(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
