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
// Instant-speed Regrowth for permanents. The ungifted line is the
// whole card here: one target permanent card from the caster's
// graveyard returns to hand.
//
// DECLARED SIMPLIFICATION, weaker than printed — the Long River's
// Pull posture: gift is a cast-time promise (CR 702.174) with no seam
// in the cast path, so the gift is never offered, the opponent never
// draws, and the two-card line never happens. Never stronger: the
// card does exactly its ungifted text.
func init() {
	Register(Spec{
		OracleID:     "938c03fc-8adf-4c7a-8ae1-eca8401f7a83",
		Name:         "Peerless Recycling",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The gift can't be promised, so the spell always returns one permanent card — never two."},
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}
			return nil
		},
	})
}
