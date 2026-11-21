package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aqueous Form — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 Enchanted creature can't be blocked.
//	 Whenever enchanted creature attacks, scry 1."
//
// "Can't be blocked" is a Restriction bit (CR 509.1b), not a keyword,
// so a creature that loses all abilities is still unblockable. The
// scry is the Aura's own trigger: its controller scries, whoever
// controls the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3378fef3-4d8c-4d9e-a338-4abb6f70d410",
		Name:         "Aqueous Form",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantBeBlocked),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Aqueous Form — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
