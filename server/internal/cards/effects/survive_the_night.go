package effects

import "github.com/google/uuid"

// Survive the Night — Instant {2}{W} (Shadows over Innistrad, #5543):
//
//	"Target creature gets +1/+0 and gains indestructible until end of
//	 turn. (Damage and effects that say "destroy" don't destroy it.)
//	 Investigate. (Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// +1/+0 and indestructible until end of turn on the one target, then
// investigate (CR 701.16), each step in printed order (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8fcc7ff-6274-4130-a3a5-99f545eeee30",
		Name:         "Survive the Night",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "indestructible")),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 1, Label: "Survive the Night — +1/+0"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Survive the Night — indestructible"}.Apply(ctx)
			}),
			Investigate{N: 1},
		),
	})
}
