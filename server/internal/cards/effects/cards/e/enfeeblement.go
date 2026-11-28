package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enfeeblement — Enchantment — Aura {B}{B} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature (Target a creature as you cast this. This card enters
//	 attached to that creature.)
//	 Enchanted creature gets -2/-2."
//
// A layer 7c modification (CR 613.4c): a creature whose toughness this takes to
// zero or less dies as a state-based action (CR 704.5f), not by destruction, so
// indestructible does not save it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b2db4c-4a1d-436f-9eeb-53a04db46c58",
		Name:         "Enfeeblement",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-2, -2)},
	})
}
