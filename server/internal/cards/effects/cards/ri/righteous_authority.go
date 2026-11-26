package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Righteous Authority — Enchantment — Aura {3}{W}{U} (Commander 2018, S82 OLD
// batch 193, #5526):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each card in its controller's hand.
//	 At the beginning of the draw step of enchanted creature's controller,
//	 that player draws an additional card."
//
// Empyrial Plate's pump (PumpAttachedPer, with DependsOnHandSize so every
// draw, discard or cast resizes the creature at once), but counting the hand
// of whoever controls the ENCHANTED creature, not the Aura's controller. The
// trigger is Super Intelligence's, on the draw step of the creature's
// controller: it fires in that player's draw step and that player draws the
// card, whoever controls the Aura. An Aura enchanting nothing on the
// battlefield neither pumps nor triggers.
//
// No simplification.
func init() {
	pump := effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
		host := g.AttachedHostOf(source)
		if host == nil {
			return 0
		}
		return effects.CardsInHand(g, host.Controller)
	})
	pump.DependsOnHandSize = true
	effects.Register(effects.Spec{
		OracleID:     "386faeef-f5cc-4ef9-b8cb-6ef06b33cbd0",
		Name:         "Righteous Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{pump},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				host := g.AttachedHostOf(source)
				return host != nil && ev.Actor != uuid.Nil && ev.Actor == host.Controller
			}, "Righteous Authority — that player draws an additional card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
