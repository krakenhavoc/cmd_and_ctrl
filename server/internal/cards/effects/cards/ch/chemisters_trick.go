package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chemister's Trick — Instant {U}{R} (Return to Ravnica, #5566):
//
//	"Target creature you don't control gets -2/-0 until end of turn and attacks
//	 this turn if able.
//	 Overload {3}{U}{R} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Overloaded it names no target (CR 702.96b) and acts on each creature you don't
// control at resolution (CR 611.2c). The requirement binds a creature that can
// attack (CR 508.1d); one that can't is unaffected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daf22b20-3d87-429c-99fe-2f3d4c6fee24",
		Name:         "Chemister's Trick",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)),
		Targets:      effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{3}{U}{R}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, match, ok := effects.OldB424Subject(ctx, effects.OpponentControls())
			if !ok {
				return nil
			}
			if err := effects.UntilEndOfTurn(ctx, id, match, "Chemister's Trick — -2/-0", game.ModifyPTMod(-2, 0)); err != nil {
				return err
			}
			return effects.UntilEndOfTurn(ctx, id, match, "Chemister's Trick — attacks this turn if able", game.AddAttackRequirementMod(uuid.Nil))
		},
	})
}
