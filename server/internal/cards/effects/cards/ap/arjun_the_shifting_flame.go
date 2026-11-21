package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arjun, the Shifting Flame — Legendary Creature — Sphinx Wizard {4}{U}{R},
// 5/5 (Commander Legends):
//
//	"Flying
//	 Whenever you cast a spell, put the cards in your hand on the bottom of
//	 your library in any order, then draw that many cards."
//
// Flying is a printed keyword. The trigger goes on the stack above the spell
// being cast (the spell has already left your hand, so it is not among the
// cards). As it resolves your whole hand is the pile: you choose the order it
// goes on the bottom of your library in (a pile of one has no order, an empty
// hand does nothing), and once it has been placed you draw as many cards as
// were put there (a keyed continuation, so the order prompt is a restore
// point). The count is how many cards really went to the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea904852-6322-4a24-9145-190acfce54ec",
		Name:            "Arjun, the Shifting Flame",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Arjun, the Shifting Flame — put your hand on the bottom of your library in any order, then draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					hand := effects.AllHandCardIDs(g, item.Controller)
					if len(hand) == 0 {
						return nil
					}
					return effects.PutInLibraryInAnyOrder{
						Cards:     hand,
						From:      game.ZoneHand,
						Placement: game.LibraryPlaceBottom,
						Label:     "Arjun, the Shifting Flame — put the cards in your hand on the bottom of your library in any order",
						ThenTail:  ctx.TailN(arjunDrawThatMany, len(hand)),
					}.Apply(ctx)
				}),
		},
	})
}

// arjunDrawThatMany is "then draw that many cards": the tail's N is how many
// cards were put on the bottom. An on-disk key: never renamed, never reused.
var arjunDrawThatMany = effects.CardTail("arjun-the-shifting-flame/draw-that-many", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return effects.DrawCards{N: o.Tail.N}.Apply(ctx)
})
