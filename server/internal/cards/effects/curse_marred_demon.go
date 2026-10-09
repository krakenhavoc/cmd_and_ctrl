package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "f30a084c-5705-40f5-8e0d-2083d33973c7",
		Name:            "Curse-Marred Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Curse-Marred Demon — search for a card, then discard a card at random",
				func(g *game.Game, item *game.StackItem) error {
					controller := item.Controller
					return SearchLibrary{
						Player:  controller,
						Dest:    game.ZoneHand,
						Limit:   1,
						Shuffle: true,
						Reason:  "Curse-Marred Demon — search your library for a card",
						Source:  item.SourceCardID,
						Then: func(g *game.Game, _ []uuid.UUID) error {
							return g.DiscardRandomForEffect(controller, 1)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
