package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdant Confluence — Sorcery {4}{G}{G} (Commander 2015, #5546):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Put two +1/+1 counters on target creature.
//	 • Return target permanent card from your graveyard to your hand.
//	 • Search your library for a basic land card, put it onto the
//	   battlefield tapped, then shuffle."
//
// Mystic Confluence's ChooseNRepeating (CR 700.2d): each occurrence runs
// once in printed order with its own target group. The search is Return
// from the Wilds' fetchBasicTapped, and a search bullet's prompt parks the
// bullets after it until it is answered (ModeDoing, #2789).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f786c99-d245-4b11-bcce-6e13fb1255cc",
		Name:         "Verdant Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"),
				PlusOneCountersOnTheModesTarget(2)),
			ModeDoing("Return target permanent card from your graveyard to your hand.",
				TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
				ReturnTheModesTargetsToHand),
			ModeDoing("Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return fetchBasicTapped(ctx.Game, item)
				}),
		),
	})
}
