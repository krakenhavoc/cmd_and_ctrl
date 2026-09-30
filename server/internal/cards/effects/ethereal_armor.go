package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethereal Armor — Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each enchantment you control
//	 and has first strike."
//
// The count is re-read on every layer recompute, so it tracks the
// board. "You" is the Aura's controller (CR 109.5), and the Aura is
// itself an enchantment you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbba75f5-2404-4bd5-982b-f6c4effa5316",
		Name:         "Ethereal Armor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, enchantmentsControlledBy),
			GrantToAttached("first strike"),
		},
	})
}
