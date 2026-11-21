package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Interdisciplinary Mascot — Creature — Elemental Fractal {6}{U}{U} 5/5
// (March of the Machine):
//
//	"Convoke. Ward {3}. When this creature enters, look at the top four
//	 cards of your library. Put one of them into your hand and the rest on
//	 the bottom of your library in a random order."
//
// The look is Farsight Ritual's (TakeFromLibraryToHand with
// takeRestBottomRandom), one card taken, the rest bottomed at random.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74d9331b-832b-420b-9249-f2a58253a361",
		Name:         "Interdisciplinary Mascot",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Interdisciplinary Mascot — ward {3}"),
			effects.WhenThisEnters("Interdisciplinary Mascot — look at the top four cards, put one into your hand and the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Controller
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 4),
						Max:    1,
						Label:  "Interdisciplinary Mascot — put one of them into your hand and the rest on the bottom of your library in a random order",
						Key:    effects.TakeRestBottomRandom,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
