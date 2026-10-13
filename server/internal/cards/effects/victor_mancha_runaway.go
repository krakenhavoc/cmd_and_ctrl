package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Victor Mancha, Runaway — Legendary Artifact Creature — Human Hero {5},
// 4/4:
//
//	"When Victor Mancha enters, exile target card from your graveyard.
//	 You may play it for as long as you control Victor Mancha."
//
// "For as long as you control Victor Mancha" is the source-bound
// duration the permission already carries (game.ForAsLongAs +
// WhileYouControlSource, CR 611.2b): it ends for good if he leaves or
// changes hands, and a Victor that was already gone as the trigger
// resolved leaves the card exiled and unplayable. The target is chosen
// when the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "606a83ae-c774-4abc-ba30-c113cbd1c05b",
		Name:         "Victor Mancha, Runaway",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Victor Mancha, Runaway — exile target card from your graveyard", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := legalTargetIDs(ctx)
					if len(ids) == 0 {
						return nil
					}
					window := forAsLongAsYouControlWindow(g, item.SourceCardID, item.Controller)
					return g.ExileCardWithPermissionForEffect(ids[0], game.CastPermission{
						Player: item.Controller, Duration: window,
					})
				}),
				TargetCardInGraveyard("target card in your graveyard", YouOwn())),
		},
	})
}
