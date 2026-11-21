package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venser's Journal — Artifact — Book {5} (Scars of Mirrodin, #5577):
//
//	"You have no maximum hand size.
//	 At the beginning of your upkeep, you gain 1 life for each card in
//	 your hand."
//
// Curiosity Crafter's NoMaxHandSize beside an upkeep trigger that counts
// its controller's hand as it resolves, not as it triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f7907327-4df3-4bb2-a274-03d9e266af2f",
		Name:          "Venser's Journal",
		Completeness:  effects.CompletenessFull,
		NoMaxHandSize: true,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Venser's Journal — gain 1 life for each card in your hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
						return nil
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: p.Hand.Size()}.Apply(ctx)
				}))),
		},
	})
}
