package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Flood Maw — Instant {U} (EDHREC rank 696):
//
//	"Gift a tapped Fish (You may promise an opponent a gift as you
//	 cast this spell. If you do, they create a tapped 1/1 blue Fish
//	 creature token before its other effects.)
//	 Return target creature an opponent controls to its owner's hand.
//	 If the gift was promised, instead return target nonland permanent
//	 an opponent controls to its owner's hand."
//
// The most-played card in the batch: a one-mana bounce for an
// opponent's creature.
//
// Sandbox simplification: the GIFT is not offered. Gift is a cast-time
// promise (CR 702.174) that needs a prompt in the cast flow and a
// per-cast flag the resolution reads, neither of which exists — the
// alternative-cost seam is the closest shape and it rewrites the
// price, not the promise. So the spell is always its base mode: target
// creature an opponent controls, back to hand. That is the printed
// card with one option removed — weaker, never stronger — and the
// target picker makes the missing option visible: it offers creatures
// only. When a cast-time promise lands, the second clause is a
// TargetPermanent(Nonland(), OpponentControls()) swap plus a tapped
// Fish for the promisee.
func init() {
	Register(Spec{
		OracleID:     "8cb36a67-9206-4665-a03f-64f52ba559c4",
		Name:         "Into the Flood Maw",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The gift can't be promised, so the spell always returns a creature an opponent controls — never another nonland permanent."},
		Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
