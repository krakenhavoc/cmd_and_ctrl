package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowerfoot Swordmaster — Creature — Mouse Soldier {W}, 1/2 (Bloomburrow,
// #5230):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this spell. If
//	 you do, when this creature enters, create a 1/1 token copy of it.)
//	 Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, Mice you control get
//	 +1/+0 until end of turn."
//
// Offspring is the shared keyword (Offspring, OffspringToken), and the
// token copy has its own valiant ability. Valiant is the shared trigger
// (valiant.go). The pump is Hire a Crew's one-shot over a set fixed as it
// resolves (CR 611.2c): the Mice you control then, the Swordmaster among
// them, and not a Mouse that arrives later in the turn. A changeling is a
// Mouse.
//
// No simplification.
func init() {
	const label = "Flowerfoot Swordmaster — Valiant: Mice you control get +1/+0 until end of turn"
	Register(Spec{
		OracleID:      "523d075d-b577-4f24-8c14-9a76761a2eaf",
		Name:          "Flowerfoot Swordmaster",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Flowerfoot Swordmaster"),
			Valiant(label, Do(BoostUntilEOT{
				Match: And(Creature(), AnySubtype("Mouse"), YouControl()),
				Power: 1,
				Label: label,
			})),
		},
	})
}
