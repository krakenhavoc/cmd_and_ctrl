package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e8fcc7ff-6274-4130-a3a5-99f545eeee30",
		Name:         "Survive the Night",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 1, Label: "Survive the Night — +1/+0"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Survive the Night — indestructible"}.Apply(ctx)
			}),
			effects.Investigate{N: 1},
		),
	})
}
