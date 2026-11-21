package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shielding Plax — {2}{G/U} Enchantment — Aura:
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature can't be the target of spells or abilities
//	 your opponents control."
//
// NOT hexproof: "your opponents" are the AURA controller's opponents,
// not the creature's controller's, the same distinction Vines of
// Vastwood makes (game/targeting_bar.go). The bar is a
// TargetingRestriction contributed by the Aura and read live at the
// targeting walk, so it lifts the moment the Aura leaves, a hexproof
// waiver can't lift it, and it binds a spell or ability anyone but the
// Aura's controller controls, the enchanted creature's controller
// included when that is an opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9bdb194-745d-4abb-9312-a65a68c40a18",
		Name:         "Shielding Plax",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shielding Plax — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		TargetingRestrictions: []game.TargetingRestriction{{
			Label: "Enchanted creature can't be the target of spells or abilities your opponents control.",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			Forbids: func(q game.TargetingQuery) bool {
				if q.Source.AttachedTo.ID != q.Card.InstanceID {
					return false
				}
				return q.Controller != q.Source.Controller
			},
		}},
	})
}
