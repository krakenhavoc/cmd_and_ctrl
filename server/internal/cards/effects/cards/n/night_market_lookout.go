package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Night Market Lookout — Creature — Human Rogue {B}, 1/1 (Kaladesh, #5539):
//
//	"Whenever this creature becomes tapped, each opponent loses 1 life and you
//	 gain 1 life."
//
// Moonfolk Puzzlemaker's "becomes tapped" trigger (the tap event and the
// attack that taps it, so a vigilance attack is not one) with Blood Artist's
// drain as a footprinted body: each opponent loses 1 life, then you gain 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f5c5f59-1cc0-4a72-863c-c26db411f114",
		Name:         "Night Market Lookout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Night Market Lookout — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
