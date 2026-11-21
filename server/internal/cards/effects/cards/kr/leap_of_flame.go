package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leap of Flame — Instant {U}{R} (Guildpact, #5599):
//
//	"Replicate {U}{R} (When you cast this spell, copy it for each time
//	 you paid its replicate cost. You may choose new targets for the
//	 copies.)
//	 Target creature gets +1/+0 and gains flying and first strike until
//	 end of turn."
//
// Replicate is Shattering Spree's pair (replicateMana and Replicate(),
// CR 702.56a), twenty payments a cast at most. Each copy pumps and
// grants to its own still-legal target (pumpAndGrantFirstTarget), so two
// copies on one creature give it +2/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "cf3b1070-2c08-49b9-aa78-fe676fab38aa",
		Name:          "Leap of Flame",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "flying", "first strike")),
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{U}{R}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 0, "Leap of Flame — +1/+0, flying and first strike", "flying", "first strike")(ctx.Game, item)
		},
	})
}
