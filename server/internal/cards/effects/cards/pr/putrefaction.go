package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Putrefaction — Enchantment {4}{B} (Mercadian Masques, #5619):
//
//	"Whenever a player casts a green or white spell, that player discards a
//	 card."
//
// Any player's cast, the spell's colours read as it is cast (green, white or
// both); the caster is the one who chooses a card from their own hand to
// discard, and a caster with an empty hand discards nothing. Putrefaction's
// controller is not special: their own green or white spells trigger it too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c4545f8-10b6-4a62-9bce-8c9ffa8b8471",
		Name:         "Putrefaction",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.HasColor(source.TextColor("G")) || spell.HasColor(source.TextColor("W")))
			}, "Putrefaction — that player discards a card", effects.TheCastersDiscardsACard),
		},
	})
}
