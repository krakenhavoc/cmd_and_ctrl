package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Executioner's Swing — Instant {W}{B} (Gatecrash, S82 OLD #5565):
//
//	"Target creature that dealt damage this turn gets -5/-5 until end of
//	 turn."
//
// The target clause reads the turn's damage events for a creature that DEALT
// damage, combat or not, after prevention (MshA6DealtDamageThisTurn), at
// announce and again as the spell resolves (CR 608.2b): a creature that has
// not dealt damage is not a legal target, and one that left and came back is a
// new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd9ab645-d8ed-4cdd-8d8e-53383b5501ba",
		Name:         "Executioner's Swing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature that dealt damage this turn", effects.MshA6DealtDamageThisTurn()),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: -5, Toughness: -5, Label: "Executioner's Swing — -5/-5"}.Apply(ctx)
		},
	})
}
