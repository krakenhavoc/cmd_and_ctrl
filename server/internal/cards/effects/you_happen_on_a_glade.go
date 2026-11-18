package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// You Happen On a Glade — Instant {2}{G} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Journey On — Search your library for up to two basic land cards, reveal
//	   them, put them into your hand, then shuffle.
//	 • Make Camp — Return target permanent card from your graveyard to your hand."
//
// Journey of Discovery's search into the hand, or Ancestral Communion's
// graveyard return of a permanent card (a land card too).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88e5300d-8339-41cb-85bf-325909116b61",
		Name:         "You Happen On a Glade",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Journey On — Search your library for up to two basic land cards, reveal them, put them into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     2,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "You Happen On a Glade — up to two basic land cards",
					}.Apply(ctx)
				}),
			ModeDoing("Make Camp — Return target permanent card from your graveyard to your hand.",
				TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent()),
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return returnLegalGraveyardTargetsToHand(ctx)
				}),
		),
	})
}
