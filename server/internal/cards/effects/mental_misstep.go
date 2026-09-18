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
// The PHYREXIAN mana's 2-life half landed in the cost engine with
// #787 (CastSpellParams.PhyrexianLife, CR 107.4c) and this card
// needed no change for it, exactly as predicted. The board is the
// half still missing: no button asks, so a cast from hand pays {U} —
// strictly fewer ways to pay than printed, never more.
func init() {
	Register(Spec{
		OracleID:     "1a0770e6-b093-4439-baff-6889a50ba12e",
		Name:         "Mental Misstep",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The board has no button for the Phyrexian symbol yet — casting it from hand pays {U}, not 2 life. The engine accepts the life payment."},
		Targets:      TargetSpell("target spell with mana value 1", b03ManaValueIs(1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
