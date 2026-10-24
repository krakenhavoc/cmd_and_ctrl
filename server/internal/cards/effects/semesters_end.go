package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Semester's End — Instant {3}{W} (Guffster's Party, #4184):
//
//	"Exile any number of target creatures and/or planeswalkers you
//	 control. At the beginning of the next end step, return each of
//	 them to the battlefield under its owner's control. Each of them
//	 enters with an additional +1/+1 counter on it if it's a creature
//	 and an additional loyalty counter on it if it's a planeswalker."
//
// Eerie Interlude's shape: every still-legal target is exiled as one
// event, and the cards that reached exile come back together at the
// next end step as new objects (CR 400.7), so each sees the others
// enter (CR 603.6a). A target that became illegal in response stays
// where it is (CR 608.2b), and a card that left exile before the end
// step is not followed (CR 603.7c).
//
// The counters are the return's own "enters with" clause (CR 614.1c),
// given card by card by what each would be on the battlefield
// (CR 614.12, #4193): a creature gets a +1/+1 counter, a planeswalker a
// loyalty counter on top of its printed loyalty (CR 306.5b), a creature
// planeswalker both. They ride the entry, so an enters trigger sees
// them and Doubling Season or Hardened Scales applies (CR 122.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3aeba3c2-5517-4870-8802-29b0fa70ffae",
		Name:         "Semester's End",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("any number of target creatures and/or planeswalkers you control",
			Or(Creature(), Planeswalker()), YouControl()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetsThenScheduleReturnBody(ctx,
				"Semester's End — return the exiled permanents with an additional counter",
				returnExiledWithAdditionalCountersBody)
		},
	})
}
