package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Master of Lake-town — Legendary Creature — Human Advisor {1}{B}{B},
// 3/2:
//
//	"Deathtouch
//	 Whenever a player loses life, that player mills that many cards.
//	 (Damage causes loss of life.)
//	 When The Master of Lake-town dies, draw a card for each graveyard
//	 with seven or more cards in it."
//
// "Loses life" is two event kinds in this engine (see
// b04OpponentLostLife, Exquisite Blood's reading): a drain writes an
// EventChangeLife with a negative amount, while damage to a player
// writes the life total directly and emits only EventDealDamage.
// Neither is emitted for the other's loss, so one loss is one trigger.
// Unlike Exquisite Blood this watches EVERY player, the Master's own
// controller included, and the loser is read off the event so it is
// that player who mills.
//
// The amount milled is the life lost, read from the event when the
// trigger resolves. A library with fewer cards mills what it has
// (CR 701.17b).
//
// The dies trigger counts graveyards when it resolves, and the Master is
// already in its owner's graveyard by then (it counts toward that
// player's pile, as printed). Each graveyard of seven or more is one
// card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "89c3c81b-8960-4f79-b2ec-f13560058071",
		Name:            "The Master of Lake-town",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
				Key:     "The Master of Lake-town — that player mills that many cards",
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, _, ok := masterOfLakeTownLoss(ev, g)
					return ok
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					who, amount, ok := masterOfLakeTownLoss(item.Trigger.Event, g)
					if !ok {
						return nil
					}
					return MillCards{Player: who, N: amount}.Apply(NewContext(g, item))
				},
			},
			WhenThisDies("The Master of Lake-town — draw a card for each graveyard with seven or more cards in it",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && p.Graveyard != nil && len(p.Graveyard.Cards) >= 7 {
							n++
						}
					}
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}

// masterOfLakeTownLoss reads a life loss off an event: who lost it and
// how much. Any player counts, the Master's controller included.
func masterOfLakeTownLoss(ev game.Event, g *game.Game) (who uuid.UUID, amount int, ok bool) {
	if ev.Target == uuid.Nil || g.PlayerByIDForEffect(ev.Target) == nil {
		return uuid.Nil, 0, false
	}
	switch ev.Kind {
	case game.EventChangeLife:
		if ev.Amount < 0 {
			return ev.Target, -ev.Amount, true
		}
	case game.EventDealDamage:
		// #2105: the life the damage cost, not the damage.
		if lost := ev.DamageLifeLoss(); lost > 0 {
			return ev.Target, lost, true
		}
	}
	return uuid.Nil, 0, false
}
