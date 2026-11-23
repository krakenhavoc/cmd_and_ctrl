package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peerless Recycling — Instant {1}{G} (EDHREC rank 2776):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Return target permanent card from your graveyard to your hand.
//	 If the gift was promised, instead return two target permanent
//	 cards from your graveyard to your hand."
//
// Instant-speed Regrowth for permanents.
//
// Gift (CR 702.174, ADR 0089): Wear Down's shape one zone over — the
// promise swaps the clause for its two-target twin (CR 702.174m) and
// the resolution returns every target that is still legal.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "938c03fc-8adf-4c7a-8ae1-eca8401f7a83",
		Name:         "Peerless Recycling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		Gift: effects.GiftACard().Instead(
			effects.TargetCardInGraveyard("two target permanent cards from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
