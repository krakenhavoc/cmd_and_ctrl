package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riot Spikes — Enchantment — Aura {B/R} (Dissension, #5598):
//
//	"({B/R} can be paid with either {B} or {R}.)
//	 Enchant creature
//	 Enchanted creature gets +2/-1."
//
// PumpAttached(2, -1), a layer 7c modification of the host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3df676fe-c4f8-41c0-9743-781090bdc547",
		Name:         "Riot Spikes",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(2, -1)},
	})
}
