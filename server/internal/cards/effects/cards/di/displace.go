package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Displace — Instant {2}{U} (Eldritch Moon, #5541):
//
//	"Exile up to two target creatures you control, then return those cards
//	 to the battlefield under their owner's control."
//
// Brago, King Eternal's blink as a spell (bragoBlinkTargets): every
// still-legal target is exiled at once, and the cards that reached exile
// come back together under their owners' control, so each sees the other
// enter (CR 603.6a). A token does not come back (CR 111.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "334e2147-5944-4558-b2bc-2f2bfb96dfa5",
		Name:         "Displace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BragoBlinkTargets(ctx.Game, item)
		},
	})
}
