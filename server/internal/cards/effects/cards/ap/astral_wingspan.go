package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Astral Wingspan — Enchantment — Aura {4}{U} (March of the Machine):
//
//	"Convoke. Enchant creature. When this Aura enters, draw a card.
//	 Enchanted creature gets +2/+2 and has flying."
//
// Lofty Dreams' shape: PumpAttached + GrantToAttached, and the draw is an
// enters trigger of the Aura itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a054b6ed-f6ca-4436-a433-89771ffdd2d6",
		Name:         "Astral Wingspan",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Astral Wingspan — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
