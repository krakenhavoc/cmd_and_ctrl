package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guildpact Informant — Creature — Faerie Rogue {2}{U}, 1/1 (War of the Spark, #5520):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player or planeswalker, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Grateful Apparition's combat-damage proliferate, on a flyer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "330f97ee-6bb2-4a4b-a8bc-411659dc4260",
		Name:            "Guildpact Informant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Guildpact Informant — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
