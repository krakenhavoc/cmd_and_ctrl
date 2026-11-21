package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parasitic Impetus — Enchantment — Aura {2}{B} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and is goaded. (It attacks each combat if
//	 able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, its controller loses 2 life and
//	 you gain 2 life."
//
// Coercive Impetus's static pair (a pump and the goad requirement that ends
// when the Aura falls off) with a drain on the attack trigger. The trigger
// is the Aura's, so "you" is the Aura's controller; "its controller" is whoever
// controls the creature as the trigger resolves (CR 608.2h: its last known
// controller when it has left), which is the attacking player unless control
// changed in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff3de07d-86c3-49e1-9de0-97c85301ff93",
		Name:         "Parasitic Impetus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Parasitic Impetus — its controller loses 2 life and you gain 2 life",
				effects.Do(effects.Instruction(effects.SocB10EnchantedCreatureControllerLoses(2)), effects.GainLife{Amount: 2})),
		},
	})
}
