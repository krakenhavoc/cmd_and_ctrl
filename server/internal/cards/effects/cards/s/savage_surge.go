package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Surge — Instant {1}{G} (Theros, S82 OLD b389, #5561):
//
//	"Target creature gets +2/+2 until end of turn. Untap that creature."
//
// Riverguard's Reflexes without the keyword: the pump, then the untap, on
// the one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0acbc9a6-0de0-4096-b35c-1a586a1312ea",
		Name:         "Savage Surge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Savage Surge — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
