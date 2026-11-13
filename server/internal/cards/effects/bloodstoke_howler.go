package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodstoke Howler — Creature — Beast {5}{R}, 3/4 (Legions, #5609):
//
//	"Morph {6}{R}
//	 When this creature is turned face up, Beast creatures you control get
//	 +3/+0 until end of turn."
//
// The morph trigger (CR 708.8) pumps the Beasts its controller controls
// as it resolves, the Howler included, and the set is locked then
// (CR 611.2c): a Beast that arrives later is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "41fa3954-119b-4ff2-ad56-9105482f6b7c",
		Name:             "Bloodstoke Howler",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{6}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Bloodstoke Howler — Beast creatures you control get +3/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(OfCreatureType("Beast"), YouControl()),
						Power: 3,
						Label: "Bloodstoke Howler — +3/+0",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
