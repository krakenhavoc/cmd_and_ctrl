package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anurid Barkripper — Creature — Frog Beast {1}{G}{G}, 2/2 (Judgment,
// #5611):
//
//	"Threshold — This creature gets +2/+2 as long as there are seven or
//	 more cards in your graveyard."
//
// Threshold is an ability word with no rules of its own (CR 207.2c): the
// layer 7c self-modifier (CR 613.4c) Krosan Beast uses, at +2/+2, read
// live on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cdd85c8-fa60-44ca-bc96-d38f029d2184",
		Name:         "Anurid Barkripper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{fraThresholdSelfPT(2, 2)},
	})
}
