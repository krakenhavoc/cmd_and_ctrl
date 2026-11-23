package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surge of Zeal — Instant {R}:
//
//	"Radiance — Target creature and each other creature that shares a
//	 color with it gain haste until end of turn."
//
// The creatures are read as the spell resolves (radianceSet, #5736),
// whoever controls them, and gain haste as one effect locked now
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24dca209-9781-4bde-ad61-18c20e8ccbd8",
		Name:         "Surge of Zeal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.RadianceSet(ctx, effects.Creature())
			if len(ids) == 0 {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Match: effects.B10OneOf(ids), Keywords: []string{"haste"}, Label: "Surge of Zeal — haste"}.Apply(ctx)
		},
	})
}
