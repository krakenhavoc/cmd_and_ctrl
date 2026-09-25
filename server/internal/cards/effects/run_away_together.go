package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Run Away Together — Instant {1}{U} (EDHREC rank 1680):
//
//	"Choose two target creatures controlled by different players.
//	 Return those creatures to their owners' hands."
//
// Two mana to save your own creature and bounce someone else's — or
// to reset two opponents' best threats at once.
//
// "Controlled by different players" is a rule about the PAIR, not
// about either creature, and since #1559 the clause says it:
// EachDifferentController is a set rule the announce gate enforces
// (CR 601.2c), so a same-controller pair is refused with the rule in
// the message rather than accepted and wasted, the picker greys a
// second creature of the first one's controller, and the bot is never
// offered the pair.
//
// Resolution (CR 608.2b) re-checks both halves. A creature that left
// in response is skipped and the other still returns. If a control
// change in response leaves both creatures under ONE player, the pair
// no longer meets the rule and neither is a legal target, so the
// spell does nothing — game/target_set.go's setRuleConflictLocked
// says why neither is preferred.
//
// No simplifications. Until #1559 this shipped with a caveat: the
// pair was checked only at resolution.
func init() {
	Register(Spec{
		OracleID:     "290faa28-450e-4797-9a8f-642d8af3f82a",
		Name:         "Run Away Together",
		Completeness: CompletenessFull,
		Targets: TargetCreature("two target creatures controlled by different players").
			WithCount(2, 2).EachDifferent(EachDifferentController()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
