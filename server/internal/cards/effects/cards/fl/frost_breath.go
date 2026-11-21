package cardsfl

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Breath —
//
// "Tap up to two target creatures. Those creatures don't untap during their
// controller's next untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "382097b3-f753-493c-bde4-101c0538feb4",
		Name:         "Frost Breath",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Frost Breath")
		},
	})
}
