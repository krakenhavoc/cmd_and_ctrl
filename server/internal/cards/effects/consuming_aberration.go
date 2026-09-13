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
// "Reveals" is not modelled as a reveal — the cards go straight to
// the graveyard, which is public, so every seat sees exactly the
// cards that were revealed a moment later than paper would show
// them. Nothing can respond in between on paper either.
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
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Consuming Aberration — each opponent mills until a land",
					func(g *game.Game, item *game.StackItem) error {
						return b14EachOpponentMillsUntilLand(g, item)
					})
			},
		}},
	})
}
