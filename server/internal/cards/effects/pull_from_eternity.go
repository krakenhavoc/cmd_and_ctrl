package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pull from Eternity -- Instant {W}:
//
//	"Put target face-up exiled card into its owner's graveyard."
//
// The first card on the exile target clause (#3259): TargetCardInExile
// narrowed by FaceUp, answered in the zone browser's exile tab and
// re-checked as the spell resolves (CR 608.2b), so a card that left
// exile in response is an illegal target and the spell is countered on
// resolution. The card goes through the graveyard route, so a commander
// card still gets its CR 903.9 choice and "if a card would be put into a
// graveyard" replacements see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2903bd66-6948-445b-a18a-12e2ce1a86c3",
		Name:         "Pull from Eternity",
		Completeness: CompletenessFull,
		Targets:      TargetCardInExile("target face-up exiled card", FaceUp()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := ctx.Game.PutIntoGraveyardThenForEffect(t.ID, nil); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
