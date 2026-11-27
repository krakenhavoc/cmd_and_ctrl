package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shriek of Dread — Instant {1}{B} (Planeshift, S82 OLD b643, #5615):
//
//	"Target creature gains fear until end of turn. (It can't be blocked except by
//	 artifact creatures and/or black creatures.)"
//
// A fear grant on the target until end of turn (CR 702.36, 611.2). A target that
// is no longer legal as the spell resolves is not affected (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee5a4902-f993-4fe6-92da-38636962c558",
		Name:         "Shriek of Dread",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "fear")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"fear"}, Label: "Shriek of Dread — fear"}.Apply(ctx)
		},
	})
}
