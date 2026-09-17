package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Breath —
//
// "Tap up to two target creatures. Those creatures don't untap during their
// controller's next untap step."
func init() {
	Register(Spec{
		OracleID:     "382097b3-f753-493c-bde4-101c0538feb4",
		Name:         "Frost Breath",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b751TapFreezeTargets(ctx, uuid.Nil, "Frost Breath")
		},
	})
}
