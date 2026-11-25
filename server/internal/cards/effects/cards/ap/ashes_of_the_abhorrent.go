package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashes of the Abhorrent — Enchantment {1}{W} (Ixalan, S82 older sets):
//
//	"Players can't cast spells from graveyards or activate abilities of cards
//	 in graveyards.
//	 Whenever a creature dies, you gain 1 life."
//
// Weathered Runestone's cast restriction over graveyards alone, and an
// activation restriction on any activation whose source is in a graveyard
// (CR 602.5; mana abilities too, since the text makes no exception). The
// trigger is one for every creature that dies, yours or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a75d5b54-5cc9-49c4-8e8c-3bef22d4c01c",
		Name:         "Ashes of the Abhorrent",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.PlayersCantCastFrom("Players can't cast spells from graveyards.", game.ZoneGraveyard),
		},
		ActivationRestrictions: []game.ActivationRestriction{{
			Label: "Players can't activate abilities of cards in graveyards.",
			Forbids: func(q game.ActivationQuery) bool {
				return q.FromZone == game.ZoneGraveyard
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Ashes of the Abhorrent — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
