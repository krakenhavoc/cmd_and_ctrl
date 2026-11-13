package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fishliver Oil — Enchantment — Aura {1}{U} (Arabian Nights, #5641):
//
//	"Enchant creature (Target a creature as you cast this. This card
//	 enters attached to that creature.)
//	 Enchanted creature has islandwalk. (It can't be blocked as long as
//	 defending player controls an Island.)"
//
// Islandwalk is an engine keyword (the block check reads it), granted
// to the enchanted creature in layer 6.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75457fe5-4ab6-42c4-98e5-8ed6e8bf122c",
		Name:         "Fishliver Oil",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{GrantToAttached("islandwalk")},
	})
}
