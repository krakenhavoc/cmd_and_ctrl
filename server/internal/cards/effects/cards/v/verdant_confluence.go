package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0f786c99-d245-4b11-bcce-6e13fb1255cc",
		Name:         "Verdant Confluence",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeDoing("Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"),
				effects.PlusOneCountersOnTheModesTarget(2)),
			effects.ModeDoing("Return target permanent card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
				effects.ReturnTheModesTargetsToHand),
			effects.ModeDoing("Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.FetchBasicTapped(ctx.Game, item)
				}),
		),
	})
}
