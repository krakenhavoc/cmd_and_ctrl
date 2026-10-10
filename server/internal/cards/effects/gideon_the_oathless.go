package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon the Oathless — Legendary Creature — Human Mercenary {2}{B},
// 3/3 (Reality Fracture):
//
//	"Ward—Discard a card.
//	 Whenever a creature an opponent controls enters, Gideon deals 1
//	 damage to that player.
//	 Whenever an opponent activates a loyalty ability, Gideon deals 1
//	 damage to that player."
//
// DECLARED SIMPLIFICATION, weaker than printed: Ward—Discard a card is
// not implemented. A ward cost is one of mana, life or a sacrifice
// (ward.go) and there is no discard component, the same gap Graveyard
// Trespasser ships with. Leaving the ward out makes Gideon easier to
// target than printed, never harder.
//
// Both damage triggers read the player from the triggering event, not
// from a target: "that player" is not a target, so hexproof and shroud
// do not matter.
func init() {
	Register(Spec{
		OracleID:     "7afa530f-ea51-4bee-af3b-59b83ea5cb25",
		Name:         "Gideon the Oathless",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Ward—Discard a card isn't implemented — opponents can target Gideon without discarding."},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anOpponentsCreatureEntered,
				"Gideon the Oathless — 1 damage to that player",
				gideonOathlessPing(func(item *game.StackItem, g *game.Game) uuid.UUID {
					if item.Trigger == nil {
						return uuid.Nil
					}
					c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
					if !ok {
						return uuid.Nil
					}
					return c.Controller
				})),
			On(game.EventActivateAbility, anOpponentActivatedALoyaltyAbility,
				"Gideon the Oathless — 1 damage to that player",
				gideonOathlessPing(func(item *game.StackItem, _ *game.Game) uuid.UUID {
					return triggeringActor(item)
				})),
		},
	})
}

// gideonOathlessPing is "Gideon deals 1 damage to that player", with
// the player read off the triggering event when the trigger resolves.
func gideonOathlessPing(who func(*game.StackItem, *game.Game) uuid.UUID) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		player := who(item, g)
		if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: player, Amount: 1}.Apply(NewContext(g, item))
	}
}

// anOpponentsCreatureEntered is "a creature an opponent controls
// enters", read from the creature as it stands on the battlefield.
func anOpponentsCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller != source.Controller
}

// anOpponentActivatedALoyaltyAbility is "an opponent activates a
// loyalty ability": the activation event's Loyalty bit (CR 606.2), by a
// player other than the source's controller.
func anOpponentActivatedALoyaltyAbility(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventActivateAbility && ev.Loyalty && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}
