package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire's Bite — Instant {B} (Zendikar, S82 OLD batch 491, #5583):
//
//	"Kicker {2}{B}
//	 Target creature gets +3/+0 until end of turn. If this spell was kicked,
//	 that creature gains lifelink until end of turn."
//
// The pump is the target's; lifelink is granted until end of turn only when
// the spell was kicked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9e88796b-f37f-4f9f-9a6f-695db2801f0e",
		Name:          "Vampire's Bite",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{B}")},
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0)),
		Targets:       effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Vampire's Bite — +3/+0"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Vampire's Bite — lifelink"}.Apply(ctx)
		},
	})
}
