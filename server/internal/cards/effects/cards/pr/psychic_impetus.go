package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Impetus — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, you scry 2."
//
// Coercive Impetus's statics (a pump and the goad pair that ends when
// the Aura falls off) with a scry on the attack trigger. "You" is the
// Aura's controller, who scries even when the enchanted creature is
// someone else's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47214acc-08c4-4dcb-9a8b-09196ce41045",
		Name:         "Psychic Impetus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Scry: 2},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Psychic Impetus — you scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
