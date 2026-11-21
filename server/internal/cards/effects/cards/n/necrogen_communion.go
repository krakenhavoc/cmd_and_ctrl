package cardsn

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrogen Communion — Enchantment — Aura {1}{B}
// (Phyrexia: All Will Be One, #5479):
//
//	"Enchant creature you control
//	 Enchanted creature has toxic 2.
//	 When enchanted creature dies, return that card to the battlefield
//	 under your control."
//
// "Your" is the Aura's controller, not the creature's owner (CR 109.5), so
// the creature comes back under the Aura controller's control as a new
// object. A token has no card to return and does nothing. The Aura is in
// the graveyard by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff58459a-0015-4012-86c0-580da4e02676",
		Name:         "Necrogen Communion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("toxic 2")},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Necrogen Communion — return that card to the battlefield under your control",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					_, err := g.ReturnToBattlefieldForEffect(item.Trigger.Event.CardID, item.Controller, false)
					if errors.Is(err, game.ErrCardNotFound) {
						return nil
					}
					return err
				}),
		},
	})
}
