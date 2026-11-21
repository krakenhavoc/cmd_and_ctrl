package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Assembly — Artifact Creature — Thopter {6}, 5/5:
//
//	"Flying
//	 At the beginning of your upkeep, if you control no Thopters other
//	 than this creature, return this creature to its owner's hand and
//	 create five 1/1 colorless Thopter artifact creature tokens with
//	 flying."
//
// An intervening "if" (CR 603.4): checked when the upkeep begins and
// again as the trigger resolves, so a Thopter that arrives in response
// stops it. The tokens are made even if the Assembly has already left
// (it does as much as it can, CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "007c110d-4364-41e5-a8bc-6ac05ecbbd00",
		Name:            "Thopter Assembly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.S82RprB09NoOtherThopter(g, source.Controller, source.InstanceID)
			}, "Thopter Assembly — return this creature to its owner's hand and create five 1/1 Thopters",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.S82RprB09NoOtherThopter(g, item.Controller, item.SourceCardID) {
						return nil
					}
					return thopterAssemblyBody(g, item)
				}),
		},
	})
}

var thopterAssemblyBody = effects.Do(
	effects.Instruction(func(ctx *effects.Context) error {
		return effects.BounceToHand{Target: ctx.Item.SourceCardID}.Apply(ctx)
	}),
	effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 5},
)
