package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foot Ninjas — Creature — Human Ninja {4}{W/B}{W/B}, 5/5 (TMT, #4457):
//
//	"Sneak {3}{W/B} (… It enters tapped and attacking.)
//	 When this creature enters, you gain 3 life."
//
// Sneak is Sneak(cost) (CR 702.190), hybrid symbol and all. The enter
// trigger is Angel of Mercy's and fires however it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "306652a8-9a9c-4b33-96d7-b59234d148e9",
		Name:             "Foot Ninjas",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{3}{W/B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Foot Ninjas — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
