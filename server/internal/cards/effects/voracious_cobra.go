package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Cobra — Creature — Snake {2}{R}{G}, 2/2 (Invasion, #5616):
//
//	"First strike
//	 Whenever this creature deals combat damage to a creature, destroy
//	 that creature."
//
// First strike is a printed keyword. The trigger is Stinkweed Imp's
// (thisDealtCombatDamageToACreature, damagedCreatureRef and
// destroyDamagedCreature): the damaged creature is named as the object
// it was when the damage was dealt (CR 510.2), so one that died to the
// damage, or left and came back, is not destroyed. The destruction can
// be regenerated, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b2e1105-5af0-4f34-bfe9-15a2f6881edd",
		Name:            "Voracious Cobra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: thisDealtCombatDamageToACreature,
			Key:       voraciousCobraDestroyLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, voraciousCobraDestroyLabel)
				item.Params.Object = damagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: destroyDamagedCreature,
		}},
	})
}

const voraciousCobraDestroyLabel = "Voracious Cobra — destroy the creature it dealt combat damage to"
