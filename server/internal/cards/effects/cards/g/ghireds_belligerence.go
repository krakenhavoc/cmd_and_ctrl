package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghired's Belligerence — Sorcery {X}{R}{R} (Commander 2019, #5517):
//
//	"Ghired's Belligerence deals X damage divided as you choose among any
//	 number of target creatures. Whenever a creature dealt damage this way
//	 dies this turn, populate. (Create a token that's a copy of a creature
//	 token you control.)"
//
// Rolling Thunder's divided X over creatures only. Each creature that is
// actually dealt damage gets its own delayed trigger (CR 603.7a), so each one
// that dies this turn populates once; a creature whose damage was prevented
// gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d543cc69-4b8a-4c45-99e1-84143d663f25",
		Name:         "Ghired's Belligerence",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB153GhiredsBelligerence(ctx)
		},
	})
}
