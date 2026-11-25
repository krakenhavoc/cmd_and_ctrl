package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spireside Infiltrator — Creature — Human Rogue {2}{R}, 3/2 (Kaladesh,
// #5539):
//
//	"Whenever this creature becomes tapped, it deals 1 damage to each
//	 opponent."
//
// Moonfolk Puzzlemaker's "becomes tapped" trigger (the tap event and the
// attack that taps it, so a vigilance attack is not one) with a footprinted
// one-damage-to-each-opponent body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f46e9fa0-2ed7-4a94-8ad4-031e87a6da45",
		Name:         "Spireside Infiltrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Spireside Infiltrator — 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
