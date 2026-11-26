package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnatural Endurance — {B} Instant (Oath of the Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Target creature gets +2/+0 until end of turn. Regenerate it."
//
// The pump and the regeneration shield (CR 701.19a) both go on the target as
// the spell resolves; a target that is illegal by then gets neither (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63701aa6-d3b8-4af2-96a4-2dec27d6a661",
		Name:            "Unnatural Endurance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:         effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Unnatural Endurance — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.Regenerate{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
