package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diminish — Instant {U} (Magic 2011, #5579):
//
//	"Target creature has base power and toughness 1/1 until end of turn."
//
// Atomic Microsizer's layer 7b base power and toughness on a spell: the
// creature's counters and other pumps still apply on top (CR 613.4b,
// 613.4c), and the effect ends at cleanup. A target gone in response means
// the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e554ab4-8179-4d3e-84ba-9a67b51cff7f",
		Name:         "Diminish",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return untilEndOfTurn(ctx, id, nil,
				"Diminish — base power and toughness 1/1 until end of turn",
				game.SetBasePTMods(1, 1)...)
		})),
	})
}
