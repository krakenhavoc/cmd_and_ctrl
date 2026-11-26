package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarmborn Giant — Creature — Giant {2}{G}{G}, 6/6 (Journey into Nyx, S82 OLD,
// #5558):
//
//	"When you're dealt combat damage, sacrifice this creature.
//	 {4}{G}{G}: Monstrosity 2. (If this creature isn't monstrous, put two +1/+1
//	 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has reach."
//
// The sacrifice watches combat damage dealt to the Giant's controller by any
// source; if it has left or come back as a new object by the time the trigger
// resolves nothing is sacrificed (CR 400.7). Monstrosity and the reach are the
// shared constructors.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d09697f-8c5c-4bf9-811f-2f847ee374f0",
		Name:         "Swarmborn Giant",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{4}{G}{G}"), 2)},
		Static:       []game.StaticAbility{effects.MonstrousKeywords("reach")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, swarmbornGiantDealtCombatDamage,
				"Swarmborn Giant — sacrifice this creature",
				effects.Do(effects.Instruction(effects.SacrificeThisIfStillHere))),
		},
	})
}

// swarmbornGiantDealtCombatDamage is "when you're dealt combat damage": combat
// damage of any amount above zero dealt to the source's controller.
func swarmbornGiantDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Target == source.Controller
}
