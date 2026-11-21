package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choose Your Weapon — Instant {2}{G} (Adventures in the Forgotten Realms,
// #5499):
//
//	"Choose one —
//	 • Two-Weapon Fighting — Double target creature's power and toughness
//	   until end of turn.
//	 • Archery — This spell deals 5 damage to target creature with flying."
//
// Doubling is Bulk Up's one-time snapshot (the Oracle ruling): +N/+N for the
// N it reads as the spell resolves, power and toughness each as they then
// are, counters and pumps included. The damage is Flourishing Strike's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3acb58d-3fb1-4602-8575-d78b716b0ea5",
		Name:         "Choose Your Weapon",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Two-Weapon Fighting — Double target creature's power and toughness until end of turn.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					c, ok := ctx.Game.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: c.CurrentPower(), Toughness: c.CurrentToughness(), Label: "Choose Your Weapon — Two-Weapon Fighting"}.Apply(ctx)
				})), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
			effects.ModeWithPurpose(effects.ModeDoing("Archery — This spell deals 5 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})), effects.ForTargets(effects.DamageToTarget(0, 5))),
		),
	})
}
