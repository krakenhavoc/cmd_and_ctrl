package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Germinating Wurm — Creature — Plant Wurm {4}{G}, 5/5 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, you gain 2 life.
//	 Warp {1}{G}"
//
// Amateur Hero's entry trigger with Mechanozoa's warp offer: cast for
// {1}{G}, the Wurm still gains you 2 life as it enters and is exiled at
// the beginning of the next end step, to be cast from exile on a later
// turn (CR 702.185a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8735c62d-1508-4bdc-a5d0-f7c68d892a5e",
		Name:             "Germinating Wurm",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Germinating Wurm — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
