package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dryad's Favor — Enchantment — Aura {G} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature has forestwalk."
//
// Fishliver Oil's grant with forestwalk: an engine keyword the block
// check reads, granted to the enchanted creature in layer 6.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af740170-af87-487c-abab-a6ccdea82d97",
		Name:         "Dryad's Favor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{GrantToAttached("forestwalk")},
	})
}
