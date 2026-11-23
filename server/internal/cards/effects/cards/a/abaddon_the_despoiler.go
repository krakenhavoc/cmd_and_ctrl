package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abaddon the Despoiler — Legendary Creature — Astartes Warrior {2}{U}{B}{R}, 5/5:
//
//	"Trample
//	 Mark of Chaos Ascendant — During your turn, spells you cast from your hand
//	 with mana value X or less have cascade, where X is the total amount of life
//	 your opponents have lost this turn."
//
// A battlefield trigger on your cast (GrantsCascade's shape): the spell's mana
// value (an X spell counts its X, CR 202.3e) is compared with the opponents'
// total life lost this turn as the spell is cast, and cascade's own limit is
// that spell's mana value.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8dedb9a7-0afa-4dff-b124-2344c170f11a",
		Name:            "Abaddon the Despoiler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.S82AfrB16GrantsCascadeWhen("Abaddon the Despoiler — cascade",
				func(ev game.Event, source *game.Card, g *game.Game) bool {
					if !effects.IsYourTurn(g, source.Controller) || ev.OldZone != game.ZoneHand {
						return false
					}
					spell, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return false
					}
					mv, known := g.ManaValueForEffect(spell)
					return known && mv <= effects.OpponentLifeLostThisTurn(g, source.Controller)
				}),
		},
	})
}
