package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lace with Moonglove — Instant {2}{G} (Conspiracy: Take the Crown, S82
// older sets batch 271, #5540):
//
//	"Target creature gains deathtouch until end of turn.
//	 Draw a card."
//
// Accelerate's shape with deathtouch: the grant ends at cleanup (CR 514.2),
// and the two sentences are steps (CR 608.2c), so a target that has gone by
// resolution only skips the grant and the card is still drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07beecfd-1720-42a1-93f8-f06604c3c228",
		Name:         "Lace with Moonglove",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Lace with Moonglove — deathtouch until end of turn"}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
