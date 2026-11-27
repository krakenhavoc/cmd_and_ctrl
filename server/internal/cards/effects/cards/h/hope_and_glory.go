package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hope and Glory — Instant {1}{W} (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Untap two target creatures. Each of them gets +1/+1 until end of turn."
//
// Exactly two targets; each one still legal as the spell resolves is
// untapped and then gets the pump (CR 608.2b), in printed order
// (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "506ea048-aca5-439a-a5b5-059ddda0672a",
		Name:         "Hope and Glory",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Hope and Glory — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
