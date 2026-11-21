package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcanist's Owl — Artifact Creature — Bird {W/U}{W/U}{W/U}{W/U}, 3/3
// (S82 reprint pass):
//
//	"Flying
//	 When this creature enters, look at the top four cards of your library.
//	 You may reveal an artifact or enchantment card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Astor's look-and-take, with the same registered take (rest on the bottom
// in a random order, the pick revealed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4bd38337-fb79-4809-848a-308c682825d1",
		Name:            "Arcanist's Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arcanist's Owl — look at the top four cards; you may reveal an artifact or enchantment card and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    effects.Or(effects.Artifact(), effects.Enchantment()),
						Max:      1,
						Optional: true,
						Label:    "Arcanist's Owl — you may reveal an artifact or enchantment card and put it into your hand",
						Key:      effects.AstorTake,
					}.Apply(ctx)
				}),
		},
	})
}
