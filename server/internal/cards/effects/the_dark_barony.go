package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Dark Barony — Plane — Ulgrotha (Planechase Anthology, #4870):
//
//	"Whenever a nonblack card is put into a player's graveyard from
//	 anywhere, that player loses 1 life.
//	 Whenever chaos ensues, each opponent discards a card."
//
// A plane (ADR 0216 §6). The first ability reads every route into a
// graveyard (graveyard_arrival.go: a discard, a mill, any zone move) and
// judges the card's color in the graveyard, not as it last existed on the
// battlefield (the ruling, CR 603.6c). The graveyard's owner is the card's
// owner, and a token is not a card (CR 111.7). Each arrival is its own
// trigger. The chaos ability makes each opponent of the planar controller
// choose and discard a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5503c308-560d-4d5f-a86a-4b06d9947ddc",
		Name:         "The Dark Barony",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny(graveyardArrivalKinds, aNonblackCardWentToAGraveyard,
				"The Dark Barony — that player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
					if !ok {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, c.Owner, -1)
				}),
			WheneverChaosEnsues("The Dark Barony — each opponent discards a card",
				func(g *game.Game, item *game.StackItem) error {
					for _, opp := range NewContext(g, item).Opponents() {
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
					}
					return nil
				}),
		},
	})
}

var darkBaronyArrivals = graveyardArrivals{Match: func(c game.Card) bool { return !b9IsBlack(c) }}

func aNonblackCardWentToAGraveyard(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && darkBaronyArrivals.arrival(ev, g, c.Owner)
}
