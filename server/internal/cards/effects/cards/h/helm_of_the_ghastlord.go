package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helm of the Ghastlord — Enchantment — Aura {3}{U/B} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 As long as enchanted creature is blue, it gets +1/+1 and has "Whenever this
//	 creature deals damage to an opponent, draw a card."
//	 As long as enchanted creature is black, it gets +1/+1 and has "Whenever this
//	 creature deals damage to an opponent, that player discards a card.""
//
// Shield of the Oversoul's colour-conditional statics, with the granted ability a
// bundle (Tandem Lookout's draw trigger) applied only while the enchanted
// creature is that colour. The draw is the creature controller's; the discard is
// the damaged player's choice (CR 701.9).
//
// No simplification.
func init() {
	const drawGrant = "helm-of-the-ghastlord/draw"
	const discardGrant = "helm-of-the-ghastlord/discard"
	blue := effects.B43AttachedHasColor("U")
	black := effects.B43AttachedHasColor("B")
	effects.Register(effects.Spec{
		OracleID:     "66fa528f-0e3a-450f-9035-1ba5a8a921e6",
		Name:         "Helm of the Ghastlord",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{
			{
				Key: drawGrant,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisDealsDamageToAnOpponent("Helm of the Ghastlord — draw a card", effects.Do(effects.DrawCards{N: 1})),
				},
				Text: "Whenever this creature deals damage to an opponent, draw a card.",
			},
			{
				Key: discardGrant,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisDealsDamageToAnOpponent("Helm of the Ghastlord — that player discards a card", func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil || g.PlayerByIDForEffect(item.Trigger.Event.Target) == nil {
							return nil
						}
						return effects.DiscardCards{Player: item.Trigger.Event.Target, N: 1}.Apply(effects.NewContext(g, item))
					}),
				},
				Text: "Whenever this creature deals damage to an opponent, that player discards a card.",
			},
		},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(blue, 1, 1),
			effects.GrantAbilities(blue, drawGrant),
			effects.B43PumpAttachedWhen(black, 1, 1),
			effects.GrantAbilities(black, discardGrant),
		},
	})
}
