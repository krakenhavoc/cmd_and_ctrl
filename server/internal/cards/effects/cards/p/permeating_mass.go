package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Permeating Mass — Creature — Spirit {G}, 1/3 (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Whenever this creature deals combat damage to a creature, that
//	 creature becomes a copy of this creature."
//
// Flayed Nim's trigger, naming the damaged creature as the object it was
// when the damage was dealt: one that died to the damage, or that left and
// came back, is not copied (CR 400.7). The copy has no stated duration
// (CR 611.2a, 707.2): it lasts, and it takes the Mass's copiable values as
// the trigger resolves, so a Mass that has left still gives them as it last
// existed.
//
// No simplification.
func init() {
	const label = "Permeating Mass — that creature becomes a copy of this creature"
	effects.Register(effects.Spec{
		OracleID:     "5e719329-612e-4ad5-ac95-6eb5328738ab",
		Name:         "Permeating Mass",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToACreature,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref := item.Params.Object
				c, ok := g.LookupCardForEffect(ref.ID)
				if !ok || c.ObjectEpoch != ref.Epoch {
					return nil
				}
				if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return effects.BecomeCopy{
					Targets:  []uuid.UUID{ref.ID},
					Of:       item.SourceCardID,
					Duration: effects.CopyIndefinite,
					Label:    label,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
