package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mental Misstep — Instant {U/P} (EDHREC rank 463):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Counter target spell with mana value 1."
//
// The free counter for Sol Ring, Swords to Plowshares and the
// one-mana dorks. The target clause reads the spell's mana value on
// the stack (b03ManaValueIs), so a Sol Ring is a legal target and a
// Counterspell is not. {X} counts as the value chosen for it (CR
// 202.3e): a {X}{U} spell cast with X=1 is mana value 2 and can't be
// targeted, and a {X} spell cast with X=1 is mana value 1 and can.
//
// Sandbox simplification: the PHYREXIAN mana is paid with {U} only.
// game.ParseCost recognises {U/P} and flags the requirement as
// Phyrexian, but no spend path offers the 2-life alternative, so the
// spell costs one blue mana — strictly fewer ways to pay than
// printed, never more. When Phyrexian payment lands in the cost
// engine this card needs no change.
func init() {
	Register(Spec{
		OracleID:     "1a0770e6-b093-4439-baff-6889a50ba12e",
		Name:         "Mental Misstep",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Phyrexian mana isn't supported — you must pay {U}, you can't pay 2 life instead."},
		Targets:      TargetSpell("target spell with mana value 1", b03ManaValueIs(1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
