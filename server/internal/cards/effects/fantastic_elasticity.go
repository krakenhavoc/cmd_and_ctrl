package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fantastic Elasticity — Sorcery {2}{U}:
//
//	"Choose one —
//	 • Return target nonland permanent to its owner's hand.
//	 • Return target instant or sorcery card from your graveyard to
//	   your hand.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The rebound cast chooses its mode afresh (CR 700.2a). Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67af9477-82b2-459e-8c09-b9a93a1d1c94",
		Name:            "Fantastic Elasticity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Modes: ChooseOne(
			ModeDoing("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland()),
				BounceTheModesTarget),
			ModeDoing("Return target instant or sorcery card from your graveyard to your hand.",
				TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}),
		),
	})
}
