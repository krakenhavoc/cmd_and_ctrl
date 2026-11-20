package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pitiless Horde — Creature — Orc Berserker {2}{B}, 5/3 (Dragons of Tarkir,
// #5551):
//
//	"At the beginning of your upkeep, you lose 2 life.
//	 Dash {2}{B}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
//
// Dash is the shared alternative cost (Ragavan's, CR 702.109): cast for
// its dash cost it gains haste and returns to its owner's hand at the
// beginning of the next end step.
// The upkeep trigger is Benthic Djinn's: life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d5726b60-60f6-4481-bc76-759fbb55bb49",
		Name:             "Pitiless Horde",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{2}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Pitiless Horde — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
