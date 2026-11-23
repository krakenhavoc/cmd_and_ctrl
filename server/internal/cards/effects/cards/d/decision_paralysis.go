package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Decision Paralysis — Instant {3}{U}
// (Amonkhet, #5535):
//
//	"Tap up to two target creatures. Those creatures don't untap during their controller's next untap step."
//
// Frost Breath's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e09ceaf-9005-4b06-ac5a-5e2250d4b663",
		Name:         "Decision Paralysis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Decision Paralysis")
		},
	})
}
