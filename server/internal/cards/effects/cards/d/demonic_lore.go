package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Lore — Enchantment {2}{B} (Commander Legends):
//
//	"When this enchantment enters, draw three cards.
//	 At the beginning of your end step, you lose 2 life for each card in
//	 your hand."
//
// The entry trigger is a plain draw. The end-step trigger counts the hand as
// it resolves, not as it triggers (Venser's Journal's read), and loses
// 2 life for each card, nothing at all with an empty hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47a9df32-7588-4430-a40e-56a9041e0360",
		Name:         "Demonic Lore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Demonic Lore — draw three cards", effects.Do(effects.DrawCards{N: 3})),
			effects.AtYourEndStep("Demonic Lore — you lose 2 life for each card in your hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2*p.Hand.Size())
				}))),
		},
	})
}
