package effects

import "github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "49aa87c7-034b-409d-b37f-50db6a42e533",
		Name:         "You See a Guard Approach",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Distract the Guard — Tap target creature.",
				TargetCreature("target creature"),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return TapTarget{Target: id}.Apply(ctx)
				})),
			ModeWithPurpose(ModeDoing("Hide — Target creature you control gains hexproof until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "You See a Guard Approach — Hide"}.Apply(ctx)
				})), ForTargets(BenefitUntilEOT(0, 0, 0, "hexproof"))),
		),
	})
}
