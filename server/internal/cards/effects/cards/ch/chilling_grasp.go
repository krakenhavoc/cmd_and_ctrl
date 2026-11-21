package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chilling Grasp — Instant {2}{U} (Eldritch Moon, #5541):
//
//	"Tap up to two target creatures. Those creatures don't untap during
//	 their controller's next untap step.
//	 Madness {3}{U} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Frost Breath's text and body (b751TapFreezeTargets: each still-legal
// target is tapped and skips its own controller's next untap step), plus
// the one Madness string (CR 702.35a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45c517d4-944c-43f6-8ead-a47965377575",
		Name:         "Chilling Grasp",
		Completeness: effects.CompletenessFull,
		Madness:      "{3}{U}",
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Chilling Grasp")
		},
	})
}
