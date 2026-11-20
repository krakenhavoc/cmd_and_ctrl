package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "334e2147-5944-4558-b2bc-2f2bfb96dfa5",
		Name:         "Displace",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return bragoBlinkTargets(ctx.Game, item)
		},
	})
}
