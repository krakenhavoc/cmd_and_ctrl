package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "938c03fc-8adf-4c7a-8ae1-eca8401f7a83",
		Name:         "Peerless Recycling",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		Gift: GiftACard().Instead(
			TargetCardInGraveyard("two target permanent cards from your graveyard", YouOwn(), Permanent()).WithCount(2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
