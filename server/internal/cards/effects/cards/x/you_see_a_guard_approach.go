package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You See a Guard Approach — Instant {U} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Distract the Guard — Tap target creature.
//	 • Hide — Target creature you control gains hexproof until end of turn."
//
// The tap is any creature's; the hexproof is granted to the target as the
// spell resolves, and only a creature the caster controls can be chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49aa87c7-034b-409d-b37f-50db6a42e533",
		Name:         "You See a Guard Approach",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Distract the Guard — Tap target creature.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.TapTarget{Target: id}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Hide — Target creature you control gains hexproof until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "You See a Guard Approach — Hide"}.Apply(ctx)
				})), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof"))),
		),
	})
}
