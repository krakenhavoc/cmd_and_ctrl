package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incriminating Impetus — Enchantment — Aura {2}{B/R}:
//
//	"Enchant creature
//	 When this Aura enters, suspect enchanted creature. (It has menace
//	 and can't block.)
//	 Enchanted creature gets +2/+2 and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)"
//
// Shiny Impetus's static pair (a +2/+2 and the goad as a requirement
// pair that ends when the Aura falls off) with Convenient Target's
// enters trigger. Suspect gives the host menace and stops it blocking,
// so a goaded, suspected creature is one that must attack and can't
// stay home. A card that asks whether a creature is goaded sees the
// static goad (Card.Goaded, #2733).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75a7cad2-782c-4a5b-9193-5ba789bf5326",
		Name:         "Incriminating Impetus",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Incriminating Impetus — suspect enchanted creature", suspectEnchantedCreature),
		},
	})
}
