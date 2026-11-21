package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Stasis Coffin — Legendary Artifact {3} (The Brothers' War):
//
//	"{2}, {T}, Exile The Stasis Coffin: You gain protection from
//	 everything until your next turn."
//
// The One Ring's shield (CR 702.16j) as the ability's whole effect. The
// exile is a cost, so the grant stands whatever becomes of the Coffin.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a46bfaf5-e631-4603-850a-f0074395b2f5",
		Name:         "The Stasis Coffin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Exile The Stasis Coffin: You gain protection from everything until your next turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GainPlayerKeyword{
					Player:   item.Controller,
					Keyword:  effects.ProtectionFromEverything,
					Label:    "The Stasis Coffin — protection from everything",
					Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
				}.Apply(ctx)
			},
		}},
	})
}
