package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trapped in the Tower — Enchantment — Aura {1}{W} (Throne of Eldraine, S82
// old sets):
//
//	"Enchant creature without flying
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated."
//
// Arrest's restriction (CR 509.1b, 602.5): can't attack or block, and no
// activated abilities, mana abilities included. The enchant clause names a
// creature without flying, which is also re-judged as state-based actions
// look at the Aura (CR 704.5m), so an enchanted creature that gains flying
// makes the Aura fall off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55746bf4-86e0-489b-b9dc-eba5f5b401dd",
		Name:         "Trapped in the Tower",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.WithoutKeyword("flying")),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
