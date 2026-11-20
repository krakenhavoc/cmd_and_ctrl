package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skirk Shaman — Creature — Goblin Shaman {1}{R}{R}, 2/2 (Planar Chaos, #5595):
//
//	"This creature can't be blocked except by artifact creatures and/or red
//	 creatures."
//
// A CR 509.1b block restriction on itself (CantBeBlockedExceptBy over
// OnSelf): a blocker must be an artifact or red as blocks are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78d5ce34-f160-4925-a4ce-25a3546d97c7",
		Name:         "Skirk Shaman",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), Or(Artifact(), OfColor("R")), "artifact creatures and/or red creatures"),
		},
	})
}
