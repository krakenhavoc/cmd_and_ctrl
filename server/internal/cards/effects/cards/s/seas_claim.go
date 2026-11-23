package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea's Claim — Enchantment — Aura {U}:
//
//	"Enchant land (Target a land as you cast this. This card enters attached to that land.)
//	 Enchanted land is an Island."
//
// ADR 0109 owner decision 4 (#1881): the static form of CR 305.7,
// effects.SetsBasicLandType, attached to the enchanted land. In layer 4 its
// land types are replaced by an Island and every other subtype stays (CR 205.1a),
// it loses the abilities its rules text gives it, keeps any another effect
// granted it (a layer-6 grant lands after the removal), and taps for {U}
// (CR 305.6). Its card types and supertypes are untouched, so an enchanted
// basic land is still basic.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed11efb6-1423-4ce0-be1d-9e0320124758",
		Name:         "Sea's Claim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static: []game.StaticAbility{
			effects.SetsBasicLandType(effects.AttachedToSource, nil, []string{"Island"}),
		},
	})
}
