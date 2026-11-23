package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vicious Conquistador — Creature — Vampire Soldier {B}, 1/2 (Foundations
// Jumpstart, S82, #5929):
//
//	"Whenever this creature attacks, each opponent loses 1 life."
//
// Bio-Asset Allocator's attack drain without the life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26b077df-7a05-459a-b5ff-bf27767bb519",
		Name:         "Vicious Conquistador",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Vicious Conquistador — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
