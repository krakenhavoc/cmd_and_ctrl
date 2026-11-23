package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "8df51012-b334-475f-b70d-af4f2b3840a0",
		Name:         "Sudden Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 3, 3))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Sudden Strength — +3/+3"}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
