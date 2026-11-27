package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Numbing Dose — Enchantment — Aura {3}{U}{U}:
//
//	"Enchant artifact or creature
//	 Enchanted permanent doesn't untap during its controller's untap step.
//	 At the beginning of the upkeep of enchanted permanent's controller, that
//	 player loses 1 life."
//
// "That player" is whoever controls the permanent as the upkeep begins, so the
// drain follows it if it changes hands.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "e42d087e-c999-4f9a-bdc5-4327e8ce94b6",
		Name:                  "Numbing Dose",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantPermanent(effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				host := g.AttachedHostOf(source)
				return host != nil && ev.Actor == host.Controller
			}, "Numbing Dose — that player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -1)
				}),
		},
	})
}
