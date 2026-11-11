package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Theoretical Duplication — Instant {2}{U}:
//
//	"Whenever a nontoken creature an opponent controls enters this turn,
//	 create a token that's a copy of that creature."
//
// CR 603.7's repeating delayed trigger (Glimpse of Nature's, the
// per-object shape of First Day of Class): it fires once for each
// qualifying creature that enters until cleanup, whoever's turn it is, and
// is not tied to the instant. The copy is the creature's copiable values as
// it is on the battlefield (CR 707.2) and is made under the caster's
// control; the token's own entry cannot retrigger it (a token is not
// nontoken). A creature that left before the trigger resolves is not
// copied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efe7e212-5c91-42f0-9dff-591131b4f6d8",
		Name:         "Theoretical Duplication",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Theoretical Duplication — create a token that's a copy of that creature",
				On:        []game.EventKind{game.EventETB},
				Condition: cmdr2S47OpponentNontokenCreatureEntersCondition,
				Body:      cmdr2S47CopyThePayloadBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}
