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
			if err := (BounceToHand{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if !ok || youControlled != ctx.Controller() {
				return nil
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
