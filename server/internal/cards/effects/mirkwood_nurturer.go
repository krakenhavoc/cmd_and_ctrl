package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirkwood Nurturer — Creature — Elf Ranger {2}{G/U} (The Hobbit), 3/2:
//
//	"When this creature enters, return up to one other target permanent you
//	 control to its owner's hand. If you do, put a +1/+1 counter on this
//	 creature."
//
// #4454, S76. Caustic Bronco's bounce with a continuation: the counter is
// the move's registered tail, run only when the permanent reached a hand
// (oneWentThisWay), so a commander that went to the command zone instead, or
// a bounce a replacement redirected, adds nothing. The hybrid cost is
// printed card data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66630640-9605-44da-afdf-eb9d458d056f",
		Name:         "Mirkwood Nurturer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mirkwood Nurturer — return up to one other permanent you control to its owner's hand; if you do, +1/+1 counter on this creature", mirkwoodNurturerBounce),
				Another(TargetPermanent("up to one other target permanent you control", YouControl())).WithCount(0, 1),
			),
		},
	})
}

func mirkwoodNurturerBounce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: id, ThenTail: ctx.ItemTail(mirkwoodNurturerCounter)}.Apply(ctx)
}

// mirkwoodNurturerCounter is "if you do, put a +1/+1 counter on this
// creature", run once the permanent has reached a hand. An on-disk key.
var mirkwoodNurturerCounter = CardTail("mirkwood-nurturer/counter", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
