package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impolite Entrance — Sorcery {R} (Lorwyn Eclipsed, #4458):
//
//	"Target creature gains trample and haste until end of turn.
//	 Draw a card."
//
// The keywords are a turn-long grant on the target as it is now, then the
// draw. With the target illegal as the spell resolves nothing happens, and
// no card is drawn (the ruling, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b77764fe-bc79-410a-995a-ac392f5b46b8",
		Name:         "Impolite Entrance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample", "haste")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample", "haste"},
				Label: "Impolite Entrance — trample and haste"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
