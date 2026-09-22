package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Misdirection — Instant {3}{U}{U}:
//
//	"You may exile a blue card from your hand rather than pay this
//	 spell's mana cost.
//	 Change the target of target spell with a single target."
//
// Force of Will's cousin, and the one card in the CR 115.7 family
// that ships with no caveat at all: it prints "target SPELL", not
// "target spell or ability", so the half the engine cannot reach
// (targeting an ability item, ADR 0065's open item) is a half this
// card never had.
//
// The pitch is the Force of Will clause with the life taken off:
// exile a blue card from hand INSTEAD of paying, no life. Both halves
// of a pitch are costs and are validated before either is paid, and
// the exile happens with Misdirection already on the stack (CR 601.2a
// before 601.2h) — so countering it does not refund the pitched card,
// which is the part a resolution-time implementation gets backwards.
//
// Misdirection cannot pitch itself: CR 601.2a moved it to the stack
// before the cost was paid, so it is no longer in hand.
func init() {
	Register(Spec{
		OracleID:     "c39e5fb0-6de3-4105-ad3c-0ecb8951a1d5",
		Name:         "Misdirection",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with a single target", HasASingleTarget()),
		AlternativeCosts: []game.AlternativeCost{
			Pitch(
				"Exile a blue card from your hand",
				0,
				CardInYourHand("a blue card from your hand", OfColor("U")),
				"a blue card from your hand",
			),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ChangeTargets{
				StackID: item.Targets[0].ID,
				Policy:  game.RetargetChangeOne,
				Reason:  "Misdirection — change the target",
			}.Apply(ctx)
		},
	})
}
