package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flourishing Strike — Instant {1}{G}:
//
//	"Choose one —
//	 • Flourishing Strike deals 5 damage to target creature with flying.
//	 • Target creature gets +3/+3 until end of turn.
//	 Entwine {2}{G} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2}{G} more takes both, each bullet with
// its own target. Entwined on one flier, the damage is dealt first and
// the pump after (CR 702.42b), so the pump does not save it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "98c44298-7b6a-4f98-bcb7-1a63a4c29f87",
		Name:          "Flourishing Strike",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}{G}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Flourishing Strike deals 5 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+3 until end of turn.",
				effects.TargetCreature("target creature"), effects.BoostTheModesTarget(3, 3, "Flourishing Strike")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
		)),
	})
}
