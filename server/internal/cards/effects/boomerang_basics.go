package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boomerang Basics — Sorcery — Lesson {U} (EDHREC rank 4403):
//
//	"Return target nonland permanent to its owner's hand. If you
//	 controlled that permanent, draw a card."
//
// Strixhaven's one-mana Lesson bounce. In Commander it is almost
// always pointed at your own permanent: a one-mana cantrip that
// rebuys an enters-the-battlefield trigger, with the option to answer
// somebody else's problem permanent at the cost of the card. Roadmap
// batch 42 (#449) lists it under "no new machinery".
//
// The "if you controlled that permanent" clause is read BEFORE the
// bounce, not after. Once the card is in its owner's hand it has no
// controller at all (CR 108.4 — only a permanent, spell or ability
// has one), so a check written after the move would never fire for
// anybody, and a check that read the OWNER instead would draw off an
// opponent's permanent you had stolen and fail to draw off your own
// permanent that an opponent owns.
//
// # The draw waits for the return, and is not gated on it (#993)
//
// A hand is a CR 903.9 destination, so the return can PAUSE: a bounced
// commander stops to ask its owner about the command zone. The draw is
// therefore the bounce's continuation (ADR 0013 §5m item 2, Path to
// Exile's shape) — otherwise the card is drawn while that question is
// still on the table, which is two things happening in the wrong
// order.
//
// The answer is deliberately IGNORED. "If you controlled that
// permanent" is a condition about the PLAYER's past relationship to the
// spell's target, not about the card that arrived: it is true or false
// before anything moves, and no replacement can rewrite it — a
// commander of yours that took the command zone was still yours. That
// is the line ADR 0013 §5t draws and stays on the ungated side of it:
// §5t gates "if it WAS a creature card", whose subject is the object
// the first sentence moved, and leaves alone a clause that merely reads
// a pre-move fact. So a return the CR 614 window cancelled still draws,
// exactly as Path to Exile still searches.
//
// A Lesson is an ordinary sorcery here: learn (CR 701.47) fetches
// Lessons from the sideboard, and Commander has no sideboard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acd881eb-f3b4-4174-9fc2-e3c7c213886b",
		Name:         "Boomerang Basics",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			youControlled, ok := controllerOfTarget(ctx, target)
			drawer := ctx.Controller()
			return BounceToHand{
				Target: target,
				Then: func(ctx *Context, _ bool) error {
					if !ok || youControlled != drawer {
						return nil
					}
					return DrawCards{Player: drawer, N: 1}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
