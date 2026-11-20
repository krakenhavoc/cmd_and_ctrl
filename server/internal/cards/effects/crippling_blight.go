package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crippling Blight — Enchantment — Aura {B} (Magic 2013, #5568):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1 and can't block."
//
// Failed Conversion's shrink at -1/-1 with the host's can't-block
// restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e365699a-bd77-473f-86b0-4e1ba0371cdc",
		Name:         "Crippling Blight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{PumpAttached(-1, -1), RestrictAttached(game.CantBlock)},
	})
}
