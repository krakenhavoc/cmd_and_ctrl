package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fa8ddc13-5108-4fe9-b3aa-dc28df476f7b",
		Name:         "Sticky Fingers",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: stickyFingersTreasure,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Sticky Fingers — create a Treasure token",
					Do(CreateToken{Template: TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, create a Treasure token.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("menace"),
			GrantAbilitiesToAttached(stickyFingersTreasure),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Sticky Fingers — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
