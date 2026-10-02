package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evil Presence — Enchantment — Aura {B}:
//
//	"Enchant land
//	 Enchanted land is a Swamp."
//
// ADR 0109 owner decision 4 (#1881): the static form of CR 305.7,
// effects.SetsBasicLandType, attached to the enchanted land. In layer 4 its
// land types are replaced by a Swamp and every other subtype stays (CR 205.1a),
// it loses the abilities its rules text gives it, keeps any another effect
// granted it (a layer-6 grant lands after the removal), and taps for {B}
// (CR 305.6). Its card types and supertypes are untouched, so an enchanted
// basic land is still basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d8ac41c-0566-48b2-a744-39db2f72272c",
		Name:         "Evil Presence",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Static: []game.StaticAbility{
			SetsBasicLandType(AttachedToSource, nil, []string{"Swamp"}),
		},
	})
}
