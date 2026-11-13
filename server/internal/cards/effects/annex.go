package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Annex — Enchantment — Aura {2}{U}{U} (Onslaught):
//
//	"Enchant land
//	 You control enchanted land."
//
// Confiscate's layer-2 control change (ControlAttachedBySource, CR
// 613.1b) on the land enchant clause. Taking control of the land does
// not take the Aura with it, or the other way round (CR 303.4e). When
// the Aura leaves the battlefield the effect ends and the land returns
// to the player it would otherwise have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f1f49dd-b7ce-494a-9b4b-ac702d3664c5",
		Name:         "Annex",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Static: []game.StaticAbility{
			ControlAttachedBySource(),
		},
	})
}
