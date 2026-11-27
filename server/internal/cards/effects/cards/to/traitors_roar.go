package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Traitor's Roar — Sorcery {4}{B/R} (Shadowmoor, S82 older sets):
//
//	"Tap target untapped creature. It deals damage equal to its power to its controller.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Backlash's body (the shared tap-then-power-damage helper) as a sorcery with conspire.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "04e9ac31-3377-4a4c-9337-53ee0ada882a",
		Name:          "Traitor's Roar",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target untapped creature", effects.UntappedCreature()),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapTargetThenItDealsItsPowerToItsController(ctx.Game, item)
		},
	})
}
