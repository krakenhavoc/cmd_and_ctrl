package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Conquistador — Creature — Vampire Soldier {B}, 1/2 (Foundations
// Jumpstart, S82, #5929):
//
//	"Whenever this creature attacks, each opponent loses 1 life."
//
// Bio-Asset Allocator's attack drain without the life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26b077df-7a05-459a-b5ff-bf27767bb519",
		Name:         "Vicious Conquistador",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Vicious Conquistador — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
