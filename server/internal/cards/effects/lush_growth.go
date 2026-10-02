package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lush Growth — Enchantment — Aura {G}:
//
//	"Enchant land
//	 Enchanted land is a Mountain, Forest, and Plains."
//
// ADR 0109 owner decision 4 (#1881): the static form of CR 305.7,
// effects.SetsBasicLandType, attached to the enchanted land. In layer 4 its
// land types are replaced by Mountain, Forest and Plains and every other subtype stays (CR 205.1a),
// it loses the abilities its rules text gives it, keeps any another effect
// granted it (a layer-6 grant lands after the removal), and taps for {R}, {G} and {W}
// (CR 305.6). Its card types and supertypes are untouched, so an enchanted
// basic land is still basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8f8fbf0-4f51-4a5c-82f9-99917513df6d",
		Name:         "Lush Growth",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Static: []game.StaticAbility{
			SetsBasicLandType(AttachedToSource, nil, []string{"Mountain", "Forest", "Plains"}),
		},
	})
}
