package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "78d5ce34-f160-4925-a4ce-25a3546d97c7",
		Name:         "Skirk Shaman",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Or(effects.Artifact(), effects.OfColor("R")), "artifact creatures and/or red creatures"),
		},
	})
}
