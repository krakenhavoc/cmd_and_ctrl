package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imodane, the Pyrohammer — Legendary Creature {2}{R}{R}, 3/3 Human
// Knight (Wilds of Eldraine, #4471):
//
//	"Whenever an instant or sorcery spell you control that targets only a
//	 single creature deals damage to that creature, Imodane deals that
//	 much damage to each opponent."
//
// The condition reads the damage event and the spell that dealt it, still
// on the stack as it resolves: the source is a spell (an instant or a
// sorcery as it is on the stack, so an Adventure half counts), controlled
// by Imodane's controller, whose announced targets are all one creature,
// and the damage went to that creature. A spell that also targets a
// player, targets two different creatures, or has no target never
// triggers it, and neither does damage from an ability. Each damage event
// is its own trigger, so a spell dealing damage to that creature twice
// triggers it twice, and the amount is the damage dealt (after any
// prevention or replacement) carried on the trigger. Imodane deals it to
// each opponent at once, as one damage event per opponent (CR 120.3).
//
// No simplification.
func init() {
	const label = "Imodane, the Pyrohammer — that much damage to each opponent"
	Register(Spec{
		OracleID:     "d70d4686-2c35-4149-aed1-7e3b9e022702",
		Name:         "Imodane, the Pyrohammer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return std54SpellDamagedItsOnlyTarget(ev, source, g)
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, item.Params.Amount)
			},
		}},
	})
}
