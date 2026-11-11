package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flubs, the Fool — Legendary Creature — Frog Scout {G}{U}{R}, 0/5
// (Bloomburrow Commander, #5368):
//
//	"You may play an additional land on each of your turns.
//	 Whenever you play a land or cast a spell, draw a card if you have no
//	 cards in hand. Otherwise, discard a card."
//
// The extra land drop is Spec.AdditionalLandPlays. The trigger is the
// "play a card" shape (a land played or a spell cast, never a land put onto
// the battlefield by an effect). Whether the hand is empty is read as the
// trigger resolves, so a spell that just left the hand counts as gone.
func init() {
	Register(Spec{
		OracleID:            "29356966-652d-4456-84c8-7aeab5fc8d05",
		Name:                "Flubs, the Fool",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			WheneverYouPlayACard("Flubs, the Fool — draw a card if you have no cards in hand, otherwise discard a card",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil {
						return nil
					}
					if p.Hand.Size() == 0 {
						return DrawCards{N: 1}.Apply(NewContext(g, item))
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Flubs, the Fool — discard a card",
					})
					return nil
				}),
		},
	})
}
