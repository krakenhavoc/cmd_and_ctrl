package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consuming Aberration — Creature — Horror {3}{U}{B}, */* (EDHREC
// rank 1606):
//
//	"Consuming Aberration's power and toughness are each equal to the
//	 number of cards in your opponents' graveyards.
//	 Whenever you cast a spell, each opponent reveals cards from the
//	 top of their library until they reveal a land card, then puts
//	 those cards into their graveyard."
//
// The mill finisher that feeds itself. The size is a
// characteristic-defining ability (CR 604.3) in layer 7a, recomputed
// whenever a graveyard changes — Tarmogoyf's mechanism, counting
// cards instead of types. The trigger is an ordinary "whenever you
// cast a spell": every opponent mills from the top, one card at a
// time, until the card just milled is a land, so the land goes too;
// an opponent whose library holds no land mills the whole thing, as
// printed.
//
// "Reveals" IS a reveal as of S22: the run is measured before
// anything moves and announced to the table as one broadcast, then
// milled. It was previously left unmodelled on the grounds that the
// graveyard is public and shows the same cards a moment later —
// which is true and is not the same thing, because it never said
// which trigger turned them over.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b55fb72-237d-4935-b645-8ebc6eb4140e",
		Name:         "Consuming Aberration",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b14CardsInOpponentsGraveyards(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, ByYou, "Consuming Aberration — each opponent mills until a land", func(g *game.Game, item *game.StackItem) error {
				return b14EachOpponentMillsUntilLand(g, item)
			}),
		},
	})
}
