package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warrior's Stand — Instant {1}{W}
// (Portal Three Kingdoms, #5622):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step."
//	"Creatures you control get +2/+2 until end of turn."
//
// The cast condition is Heavy Fog's: the declare attackers step, after a
// creature was declared attacking you (CR 508.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "1f08072d-7dec-4f0d-b465-dda62eb17630",
		Name:               "Warrior's Stand",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2, Label: "Warrior's Stand — creatures you control get +2/+2"}.Apply(ctx)
		},
	})
}
