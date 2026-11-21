package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flight-Deck Coordinator — Creature — Human Soldier {2}{W}, 3/3 (Edge of
// Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more tapped
//	 creatures, you gain 2 life."
//
// Vaultguard Trooper's end-step trigger with a life gain in place of the
// discard. The "if" is an intervening if (CR 603.4): checked as the end
// step begins and again as the ability resolves, so untapping a creature
// in response leaves it doing nothing. The Coordinator itself counts
// toward the two tapped creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57927909-11a9-44ef-989d-1dcca1f6e23f",
		Name:         "Flight-Deck Coordinator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *effects.Context) bool {
					return effects.B24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2
				},
			}, effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Flight-Deck Coordinator — you gain 2 life", effects.Do(effects.GainLife{Amount: 2}))),
		},
	})
}
