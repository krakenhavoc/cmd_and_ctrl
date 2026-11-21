package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderbrine Preacher — Creature — Merfolk Cleric {1}{W}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature becomes tapped, you gain 2 life."
//
// Compassionate Healer's trigger with two life and no scry: "becomes
// tapped" is the tap event and the attack that taps it, so a vigilance
// attack is not one and crewing or paying a cost is (CR 701.26a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b4dcdb1-0017-494f-8d83-6706b188e575",
		Name:         "Wanderbrine Preacher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Wanderbrine Preacher — you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
