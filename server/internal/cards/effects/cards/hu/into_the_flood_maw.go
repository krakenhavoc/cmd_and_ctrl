package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// opponent's creature, or for any nonland permanent at the price of a
// Fish.
//
// Gift (CR 702.174, ADR 0089): the promise is made at CR 601.2b and
// the target at 601.2c, so the promised text is a whole different
// target clause — `.Instead(…)` — and the engine announces, re-checks
// and offers the clause the cast was actually made under (CR
// 702.174m). The resolution is the same bounce either way; only what
// it may be aimed at changes. The Fish is the token catalog's 1/1 blue
// Fish, entering tapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cb36a67-9206-4665-a03f-64f52ba559c4",
		Name:         "Into the Flood Maw",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		Gift: effects.GiftATappedFish().Instead(
			effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
