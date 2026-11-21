package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wear Away — Instant — Arcane {G}{G}:
//
//	"Destroy target artifact or enchantment.
//	Splice onto Arcane {3}{G}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be7cd599-12b6-43b6-a22a-8a8f7dd1ecd2",
		Name:         "Wear Away",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{3}{G}"),
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
