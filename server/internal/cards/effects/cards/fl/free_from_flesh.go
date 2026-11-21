package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Free from Flesh — Instant {R} (Phyrexia: All Will Be One):
//
//	"Target creature gets +2/+2 until end of turn. Put two oil counters on it."
//
// Both halves read the same target as the spell resolves; with the target gone the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4a96cc7-8a4f-4007-8477-e1f30720db7b",
		Name:         "Free from Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Free from Flesh — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.AddCounter{Target: t.ID, Kind: effects.S82OneB09Oil, N: 2}.Apply(ctx)
		},
	})
}
