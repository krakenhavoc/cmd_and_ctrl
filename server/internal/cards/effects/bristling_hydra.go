package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristling Hydra — Creature — Hydra {2}{G}{G}, 4/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}{E}{E}: Put a +1/+1 counter on this creature. It gains
//	 hexproof until end of turn."
//
// ADR 0129 PR 1 (#1995). Both halves land on the Hydra that activated
// it; one that left and came back is a new object and gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3b23c58-0b7a-4fe4-a8e8-5320a7605724",
		Name:         "Bristling Hydra",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Bristling Hydra", 3),
		},
		Activated: []ActivatedAbility{{
			Label: "Pay {E}{E}{E}: Put a +1/+1 counter on this creature. It gains hexproof until end of turn.",
			Cost:  PayEnergy(3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := plusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("hexproof", "Bristling Hydra — hexproof until end of turn")(g, item)
			},
		}},
	})
}
