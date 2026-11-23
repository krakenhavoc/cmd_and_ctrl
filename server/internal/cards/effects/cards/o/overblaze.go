package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overblaze — Instant — Arcane {3}{R}:
//
//	"Each time target permanent would deal damage to a permanent or
//	player this turn, it deals double that damage to that permanent or
//	player instead.
//	Splice onto Arcane {2}{R}{R}"
//
// ADR 0108 §3's multiplier pinned to the target (CR 614.1a), every
// instance this turn, any recipient. Splice is ADR 0202 (#1989): spliced
// onto an Arcane spell, this text runs after the spell's own with the
// spell as its source.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "543599e2-8312-43a6-ab26-0be17e690d4f",
		Name:         "Overblaze",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{2}{R}{R}"),
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.MultiplyDamage{Factor: 2, From: t.ID, Label: "Overblaze"}.Apply(ctx)
		},
	})
}
