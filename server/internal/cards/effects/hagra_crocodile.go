package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hagra Crocodile — Creature — Crocodile {3}{B}, 3/1 (Zendikar, #5583):
//
//	"This creature can't block.
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// The can't-block restriction is a static on the creature itself
// (Charging Slateback's). Canopy Baloth's landfall pump: a land
// entering under its controller's control, the boost on the creature
// while it is still the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8bca3213-d8e1-4915-bd9a-9a291cdec1cd",
		Name:         "Hagra Crocodile",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			Landfall("Hagra Crocodile — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Hagra Crocodile — +2/+2 until end of turn")),
		},
	})
}
