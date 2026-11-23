package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Cheetah — Creature — Cat {2}{G}{W}, 2/2 (Invasion, #5616):
//
//	"Whenever this creature deals damage, you gain that much life."
//
// Exalted Angel's trigger without the morph: any damage the Cheetah
// deals, combat or not, to a player or to a permanent, read off the
// event's Amount (the damage actually dealt, after prevention). It is a
// triggered ability, not lifelink, so the life comes when it resolves.
// The life goes to the Cheetah's controller (CR 109.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a552edbe-bd5d-4281-8d9f-d03f4b10815d",
		Name:         "Horned Cheetah",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Horned Cheetah — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
