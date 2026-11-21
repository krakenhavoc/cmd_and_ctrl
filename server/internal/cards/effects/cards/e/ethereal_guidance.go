package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethereal Guidance — Sorcery {2}{W} (Shadows over Innistrad, #5543):
//
//	"Creatures you control get +2/+1 until end of turn."
//
// On the Job's +2/+1 to the creatures you control until end of turn,
// without the Clue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e24c0d2-2333-425c-952e-9a13e6519cd3",
		Name:         "Ethereal Guidance",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1, Label: "Ethereal Guidance — +2/+1"},
		),
	})
}
