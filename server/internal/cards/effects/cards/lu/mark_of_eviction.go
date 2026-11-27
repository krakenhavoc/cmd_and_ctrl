package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mark of Eviction — Enchantment — Aura {U} (Ravnica: City of Guilds, S82 OLD
// batch 581, #5600):
//
//	"Enchant creature
//	 At the beginning of your upkeep, return enchanted creature and all Auras
//	 attached to that creature to their owners' hands."
//
// "Your" is the Aura's controller, who need not control the creature. The
// creature and every Aura attached to it as the trigger resolves (Mark of
// Eviction itself among them) go to their owners' hands in one simultaneous
// move (CR 400.7); a trigger whose Aura is not attached to anything does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9193641-a7f4-40d3-ac62-15015acece4e",
		Name:         "Mark of Eviction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mark of Eviction — return enchanted creature and all Auras attached to it to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					host := effects.AttachedHostFor(g, item.SourceCardID)
					if host == nil {
						return nil
					}
					ids := []uuid.UUID{host.InstanceID}
					for _, a := range g.BattlefieldCardsForEffect() {
						if a.IsAttachedTo(host.InstanceID) && a.IsAura() {
							ids = append(ids, a.InstanceID)
						}
					}
					return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
				}),
		},
	})
}
