package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withstand Death — Instant {G} (Scars of Mirrodin, #5577):
//
//	"Target creature gains indestructible until end of turn."
//
// Ephemeral Shields' grant without the convoke.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c462adf3-4ade-4b90-bdec-7247b080471f",
		Name:         "Withstand Death",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Withstand Death — indestructible"}.Apply(ctx)
		},
	})
}
