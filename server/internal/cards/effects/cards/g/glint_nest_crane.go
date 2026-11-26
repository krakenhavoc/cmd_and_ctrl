package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint-Nest Crane — Creature — Bird {1}{U}, 1/3 (Kaladesh, #5539):
//
//	"Flying
//	 When this creature enters, look at the top four cards of your library.
//	 You may reveal an artifact card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in any order."
//
// Rosecot Knight's look-reveal-and-take over four cards and artifact cards,
// with Impulse's "the rest on the bottom in any order" and no rider. Declining
// the reveal, or no artifact among the four, puts all four on the bottom.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1983478a-cedb-4053-adab-ac5f1a163296",
		Name:            "Glint-Nest Crane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glint-Nest Crane — look at the top four cards of your library for an artifact card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    effects.Artifact(),
						Max:      1,
						Optional: true,
						Label:    "Glint-Nest Crane — you may reveal an artifact card and put it into your hand",
						Key:      glintNestCraneTake,
					}.Apply(ctx)
				}),
		},
	})
}

// glintNestCraneTake is the take: the rest on the bottom in any order. An
// on-disk key.
var glintNestCraneTake = effects.RegisterTake("choose-cards/glint-nest-crane-take",
	effects.TakeDef{Reveal: true, Then: effects.TakeRestOnBottomInAnyOrder})
