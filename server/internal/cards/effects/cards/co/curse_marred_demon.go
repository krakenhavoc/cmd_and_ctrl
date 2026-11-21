package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse-Marred Demon — Creature — Demon {2}{R}{R}, 4/4:
//
//	"Flying, trample
//	 When this creature enters, search your library for a card, put it
//	 into your hand, shuffle, then discard a card at random."
//
// Gamble's shape on a body. The random discard is chained in the
// search's continuation so it takes the card from the hand the
// tutored card is already in — which may be the tutored card itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f30a084c-5705-40f5-8e0d-2083d33973c7",
		Name:            "Curse-Marred Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Curse-Marred Demon — search for a card, then discard a card at random",
				func(g *game.Game, item *game.StackItem) error {
					controller := item.Controller
					return effects.SearchLibrary{
						Player:   controller,
						Dest:     game.ZoneHand,
						Limit:    1,
						Shuffle:  true,
						Reason:   "Curse-Marred Demon — search your library for a card",
						Source:   item.SourceCardID,
						ThenTail: game.KeyedTail{Ref: effects.ThenDiscardAtRandom, Player: controller, N: 1},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
