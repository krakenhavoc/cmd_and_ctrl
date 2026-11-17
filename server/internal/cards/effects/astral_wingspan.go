package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astral Wingspan — Enchantment — Aura {4}{U} (March of the Machine):
//
//	"Convoke. Enchant creature. When this Aura enters, draw a card.
//	 Enchanted creature gets +2/+2 and has flying."
//
// Lofty Dreams' shape: PumpAttached + GrantToAttached, and the draw is an
// enters trigger of the Aura itself.
func init() {
	Register(Spec{
		OracleID:     "a054b6ed-f6ca-4436-a433-89771ffdd2d6",
		Name:         "Astral Wingspan",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Astral Wingspan — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
