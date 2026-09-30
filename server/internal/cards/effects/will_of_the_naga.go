package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Will of the Naga — Instant {4}{U}{U}:
//
//	"Delve. Tap up to two target creatures. Those creatures don't untap during their controller's next untap step."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// Frost Breath's effect with delve. No simplification.
func init() {
	Register(Spec{
		OracleID:     "520b6637-0a9f-4dc4-846e-f1cd2b868263",
		Name:         "Will of the Naga",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b751TapFreezeTargets(ctx, uuid.Nil, "Will of the Naga")
		},
	})
}
