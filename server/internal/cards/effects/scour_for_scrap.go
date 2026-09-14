package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scour for Scrap — Instant {3}{U} (EDHREC rank 2572):
//
//	"Choose one or both —
//	 • Search your library for an artifact card, reveal it, put it
//	   into your hand, then shuffle.
//	 • Return target artifact card from your graveyard to your hand."
//
// The artifact deck's two-for-one tutor. "Choose one or both" is a
// ChooseN(1, 2) mode clause; only the second option targets, which
// is the one-targeted-option limit the modal engine has. Modes
// resolve in printed order (CR 700.2c): the search first — through
// the S22 chooser, revealed — then the graveyard return, so a card
// tutored to hand and a card returned to hand are two different
// cards, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da422b3e-1092-47ad-97d6-a7a3ef060a53",
		Name:         "Scour for Scrap",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			Mode("Search your library for an artifact card, reveal it, put it into your hand, then shuffle."),
			Mode("Return target artifact card from your graveyard to your hand.",
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				if err := b06TutorToHand("Scour for Scrap — an artifact card", b03IsArtifactCard)(item, ctx); err != nil {
					return err
				}
			}
			if ctx.HasMode(1) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
