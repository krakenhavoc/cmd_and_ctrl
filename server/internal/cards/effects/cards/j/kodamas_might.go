package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kodama's Might — Instant — Arcane {G}:
//
//	"Target creature gets +2/+2 until end of turn.
//	Splice onto Arcane {G}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c027a8ef-6cdd-4c14-9d4f-b21c63312874",
		Name:         "Kodama's Might",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{G}"),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Kodama's Might — +2/+2"}.Apply(ctx)
		},
	})
}
