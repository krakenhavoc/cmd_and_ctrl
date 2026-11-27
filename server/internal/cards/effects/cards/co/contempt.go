package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contempt — Enchantment — Aura {1}{U} (Stronghold, S82 OLD b751, #5627):
//
//	"Enchant creature
//	 When enchanted creature attacks, return it and this Aura to their owners'
//	 hands at end of combat."
//
// An attack trigger on the Aura (WheneverEnchantedCreatureAttacks) that
// schedules a delayed trigger at the beginning of the end of combat step (CR
// 511.2, 603.7) naming both objects as they are now (EffectParams.Objects, CR
// 400.7). Each is returned only if it is still that object on the battlefield, so
// a creature that died or an Aura that fell off is not pulled back from
// anywhere, and the other still goes. Each goes to its OWNER's hand, so an
// opponent's creature returns to its owner.
//
// No simplification.
func init() {
	const label = "Contempt — return enchanted creature and this Aura to their owners' hands at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "811c8364-fafb-49c8-a5bf-7b6cdfceb475",
		Name:         "Contempt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				var objects []game.ObjectRef
				for _, id := range []uuid.UUID{item.Trigger.Event.CardID, item.SourceCardID} {
					if effects.OnBattlefield(g, id) {
						objects = append(objects, game.ObjectRef{ID: id, Epoch: g.ObjectEpochForEffect(id)})
					}
				}
				if len(objects) == 0 {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					At:     game.StepEndCombat,
					Label:  "Contempt — return the creature and the Aura to their owners' hands",
					Body:   effects.OldB751ReturnTheObjectsToHandsBody,
					Params: game.EffectParams{Objects: objects},
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
