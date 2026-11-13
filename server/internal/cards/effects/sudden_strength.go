package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Strength — Instant {3}{G} (Judgment, #5611):
//
//	"Target creature gets +3/+3 until end of turn.
//	 Draw a card."
//
// Auspicious Arrival's shape with a draw for the Clue: the pump ends at
// cleanup (CR 514.2) and is declared as a benefit, and the caster draws
// as the next step (CR 608.2c). A target that has gone counters the
// whole spell, so no card is drawn then (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8df51012-b334-475f-b70d-af4f2b3840a0",
		Name:         "Sudden Strength",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 3, 3))},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Sudden Strength — +3/+3"}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
