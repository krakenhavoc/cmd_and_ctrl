package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trespasser's Curse — Enchantment — Aura Curse {1}{B} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant player
//	 Whenever a creature enchanted player controls enters, that player loses 1
//	 life and you gain 1 life."
//
// One trigger per creature that enters under the enchanted player's control
// (CR 603.2c), the player read as the creature enters. The life loss comes
// first and the gain follows it, as printed; both are for the Curse's
// controller and the enchanted player, even if the creature is gone by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2769751d-ca67-4d6e-93c9-9587367ae8ab",
		Name:         "Trespasser's Curse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller == source.AttachedTo.ID
			}, "Trespasser's Curse — that player loses 1 life and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if aura, ok := g.LookupCardForEffect(ctx.Source()); ok && aura.AttachedTo.Kind == game.TargetPlayer {
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), aura.AttachedTo.ID, -1); err != nil {
							return err
						}
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				}),
		},
	})
}
