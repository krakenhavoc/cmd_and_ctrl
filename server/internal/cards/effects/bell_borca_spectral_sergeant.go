package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bell Borca, Spectral Sergeant — Legendary Creature — Spirit Soldier
// {2}{R}{W}, */5:
//
//	"Note the mana value of each card as it's put into exile.
//	 Bell Borca's power is equal to the greatest number noted for it this
//	 turn.
//	 At the beginning of your upkeep, exile the top card of your library.
//	 You may play that card this turn."
//
// "Note … as it's put into exile" is a static ability, so it works only
// while Bell Borca is on the battlefield: the cards noted are those put
// into exile (from any zone, by any player, not tokens) after Bell Borca
// last entered this turn. Nothing noted is a power of 0. The count is the
// turn's event log read by the characteristic-defining power (CR 604.3,
// layer 7a); DependsOnEvents{EventZoneMove} drops the cached layer pass
// the moment a card lands in exile. A card that has since left exile
// (played, say) still counts at its mana value. Bell Borca's upkeep
// trigger exiles the top card with the usual "you may play it this turn"
// permission.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e29aea1-06e3-49ef-8a22-e3acff1fb679",
		Name:         "Bell Borca, Spectral Sergeant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:           "Bell Borca's power is equal to the greatest number noted for it this turn.",
			Layer:           game.Layer7PT,
			SubLayer:        game.SubLayer7A_CDA,
			AppliesTo:       selfOnly,
			DependsOnEvents: []game.EventKind{game.EventZoneMove},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = bellBorcaGreatestNoted(g, source.InstanceID)
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bell Borca, Spectral Sergeant — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := b12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}

// bellBorcaGreatestNoted is the greatest mana value among the cards put
// into exile this turn while the permanent `bell` has been on the
// battlefield (since it last entered, if that was this turn), 0 when there
// are none.
func bellBorcaGreatestNoted(g *game.Game, bell uuid.UUID) int {
	events := g.EventsThisTurn()
	from := -1
	for i, ev := range events {
		if ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield && ev.CardID == bell {
			from = i
		}
	}
	greatest := 0
	for _, ev := range events[from+1:] {
		if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneExile {
			continue
		}
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || IsToken(card) {
			continue
		}
		if mv := card.ManaValue(); mv > greatest {
			greatest = mv
		}
	}
	return greatest
}
