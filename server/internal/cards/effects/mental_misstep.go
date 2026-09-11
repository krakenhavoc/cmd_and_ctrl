package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mental Misstep — Instant {U/P} (EDHREC rank 463):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Counter target spell with mana value 1."
//
// The free counter for Sol Ring, Swords to Plowshares and the
// one-mana dorks. The target clause reads the printed mana value
// (CR 202.3), so a Sol Ring is a legal target and a Counterspell is
// not; an {X} spell is mana value 1 only when its printed cost says
// so, since the sandbox does not read the announced X for this
// clause (Mana Drain's refund is the one place that does).
//
// Sandbox simplification: the PHYREXIAN mana is paid with {U} only.
// game.ParseCost recognises {U/P} and flags the requirement as
// Phyrexian, but no spend path offers the 2-life alternative, so the
// spell costs one blue mana — strictly fewer ways to pay than
// printed, never more. When Phyrexian payment lands in the cost
// engine this card needs no change.
func init() {
	Register(Spec{
		OracleID: "1a0770e6-b093-4439-baff-6889a50ba12e",
		Name:     "Mental Misstep",
		Targets:  TargetSpell("target spell with mana value 1", b03ManaValueIs(1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
