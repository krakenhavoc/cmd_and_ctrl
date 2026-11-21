package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const stickyFingersTreasure = "sticky-fingers/treasure"

// Sticky Fingers — Enchantment — Aura {R}:
//
//	"Enchant creature
//	 Enchanted creature has menace and 'Whenever this creature deals
//	 combat damage to a player, create a Treasure token.'
//	 When enchanted creature dies, draw a card."
//
// Two different owners, which is the whole card. The Treasure trigger
// is GRANTED to the creature (ADR 0093), so it is the creature's
// controller who creates it — a stolen creature pays its new
// controller — and a creature that loses all abilities loses it. The
// draw is the Aura's own dies trigger, so the Aura's controller draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa8ddc13-5108-4fe9-b3aa-dc28df476f7b",
		Name:         "Sticky Fingers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: stickyFingersTreasure,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Sticky Fingers — create a Treasure token",
					effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, create a Treasure token.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("menace"),
			effects.GrantAbilitiesToAttached(stickyFingersTreasure),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Sticky Fingers — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
