package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower Geist — Creature — Spirit {3}{U}, 2/2 (Innistrad Remastered):
//
//	"Flying
//	 When this creature enters, look at the top two cards of your library.
//	 Put one of them into your hand and the other into your graveyard."
//
// Looked at, not revealed: the pair is shown to its controller alone. A
// library of one card gives that card to the hand; an empty one does
// nothing. The take is Pattern Buffer Preservation's (takeRestGraveyard).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10e19453-84fb-4845-8db0-c9da52847e09",
		Name:            "Tower Geist",
		Purpose:         game.Purpose{Draws: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tower Geist — look at the top two cards of your library, one to hand and the other to your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  "Tower Geist — put one of them into your hand and the other into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}),
		},
	})
}
