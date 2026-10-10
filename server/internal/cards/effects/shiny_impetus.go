package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shiny Impetus — Enchantment — Aura {2}{R}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, you create a Treasure
//	 token."
//
// The goad is a static requirement pair (GoadAttached): it binds the
// declaration exactly as a goad does (CR 701.15b, counted under
// CR 508.1d) and ends the moment the Aura falls off. A card that asks
// whether a creature is goaded sees it (Card.Goaded, #2733).
//
// The Treasure is the Aura's trigger, so its controller gets it.
func init() {
	Register(Spec{
		OracleID:     "aae76e5c-f5e0-4d18-b465-e6a829be908a",
		Name:         "Shiny Impetus",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Shiny Impetus — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
