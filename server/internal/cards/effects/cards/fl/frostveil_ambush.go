package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frostveil Ambush — Instant {3}{U}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Tap up to two target creatures. Those creatures don't untap during
//	 their controller's next untap step.
//	 Cycling {1} ({1}, Discard this card: Draw a card.)"
//
// Frost Breath's tap-and-freeze body (b751TapFreezeTargets), and
// cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cf069cf-3a10-4427-a02b-5405429b5652",
		Name:         "Frostveil Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Frostveil Ambush")
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}")},
	})
}
