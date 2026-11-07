package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanguine Syphoner — Creature — Vampire Warlock {1}{B}, 1/3
// (Foundations, #4465):
//
//	"Whenever this creature attacks, each opponent loses 1 life and you
//	 gain 1 life."
//
// Bio-Asset Allocator's granted attack drain, printed on the creature
// itself: every opponent loses 1 life, whichever one it attacked, then
// you gain 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b81a748b-f7a6-4ec4-8b18-40a1819745f6",
		Name:         "Sanguine Syphoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sanguine Syphoner — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
