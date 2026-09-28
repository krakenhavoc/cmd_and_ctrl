package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avacyn's Judgment — Sorcery {1}{R} (#1657, #653):
//
//	"Madness {X}{R} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)
//	 Avacyn's Judgment deals 2 damage divided as you choose among any
//	 number of targets. If this spell's madness cost was paid, it deals
//	 X damage divided as you choose among those permanents and/or
//	 players instead."
//
// The amount is DivideBy(DivideTwoOrXIfMadness), read at announce
// with the claimed alternative cost: 2 for a cast that paid {1}{R},
// the announced X for one that claimed the madness grant (the key
// StackItem.AltCost records, CR 702.35b). The madness cost is {X}{R}
// and the printed cost has no {X}, so CR 107.3b does not lock X for
// the madness cast — the X is announced with it, and it is also the
// amount divided. "Those permanents and/or players" are the same
// targets the spell chose; each still gets at least 1 (CR 601.2d),
// so a madness cast for X can hit at most X of them.
//
// Madness is #657's machinery, declared once with Spec.Madness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3ae58ed-8ef7-4e0a-945f-1f622157236b",
		Name:         "Avacyn's Judgment",
		Completeness: CompletenessFull,
		Madness:      "{X}{R}",
		Targets:      TargetAny().WithCount(0, 0).Dividing(DivideBy(DivideTwoOrXIfMadness)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
