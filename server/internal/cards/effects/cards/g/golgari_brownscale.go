package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Brownscale — Creature — Lizard {1}{G}{G}, 2/3 (Ravnica: City of Guilds):
//
//	"When this card is put into your hand from your graveyard, you gain 2 life.
//	 Dredge 2 (If you would draw a card, you may mill two cards instead. If you
//	 do, return this card from your graveyard to your hand.)"
//
// Dredge is dredge.go. The life trigger is an ability of a card in the HAND
// (CR 113.6, Zones {hand}): it watches a card's move from the graveyard to
// the hand and applies only to itself, so a Brownscale that is drawn, or
// returned by another card from any other zone, gains nothing. Whatever puts
// the card into the hand from a graveyard (dredge, Regrowth) triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34242c44-a987-412c-9652-5242b9b5200c",
		Name:         "Golgari Brownscale",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Dredge(2)},
		Triggered: []game.TriggeredAbility{func() game.TriggeredAbility {
			t := effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID &&
					ev.OldZone == game.ZoneGraveyard && ev.NewZone == game.ZoneHand
			}, "Golgari Brownscale — gain 2 life", effects.Do(effects.GainLife{Amount: 2}))
			t.Zones = []game.ZoneKind{game.ZoneHand}
			return t
		}()},
	})
}
