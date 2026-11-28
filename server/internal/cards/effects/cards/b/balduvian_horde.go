package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balduvian Horde — Creature — Human Barbarian {2}{R}{R} 5/5 (Alliances, S82
// OLD batch 819):
//
//	"When this creature enters, sacrifice it unless you discard a card at
//	 random."
//
// "Unless you discard" is "you may discard a card at random; if you don't,
// sacrifice it" (CR 118.12a): a player with an empty hand cannot, and so
// sacrifices it.
func init() {
	const label = "Balduvian Horde — sacrifice it unless you discard a card at random"
	effects.Register(effects.Spec{
		OracleID:     "fd2a94d3-901d-4b46-aa9b-a30ed6dea6bd",
		Name:         "Balduvian Horde",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				cost, action := effects.RandomDiscardPayment(1).Prompt()
				return effects.UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     cost,
					Action:   action,
					Question: label,
					Else:     balduvianHordeSacrifice,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var balduvianHordeSacrifice = effects.PayUnlessBranches("pay-unless/balduvian-horde-sacrifice", nil,
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.SacrificePermanent{Target: r.Source}.Apply(ctx)
	})
