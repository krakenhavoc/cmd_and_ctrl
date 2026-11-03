package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coercive Impetus — Enchantment — Aura {2}{B}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, you draw a card and lose 1
//	 life."
//
// Shiny Impetus's static pair (a pump and the goad as a requirement pair
// that ends when the Aura falls off) with Elegy Acolyte's draw-and-lose
// body on the attack trigger. The trigger is the Aura's, so its
// controller draws and loses the life, whoever controls the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbb8a3d9-6b1b-4ad1-b8b9-f92ead03d1d9",
		Name:         "Coercive Impetus",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Coercive Impetus — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
	})
}
