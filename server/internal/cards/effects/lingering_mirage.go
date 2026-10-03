package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lingering Mirage — Enchantment — Aura {1}{U}:
//
//	"Enchant land
//	 Enchanted land is an Island.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// ADR 0109 owner decision 4 (#1881): the static form of CR 305.7,
// effects.SetsBasicLandType, attached to the enchanted land. In layer 4 its
// land types are replaced by an Island and every other subtype stays (CR 205.1a),
// it loses the abilities its rules text gives it, keeps any another effect
// granted it (a layer-6 grant lands after the removal), and taps for {U}
// (CR 305.6). Its card types and supertypes are untouched, so an enchanted
// basic land is still basic.
//
// Cycling is the hand ability (CR 702.29).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cce17310-585d-4738-8d5b-bbad65008c37",
		Name:         "Lingering Mirage",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Static: []game.StaticAbility{
			SetsBasicLandType(AttachedToSource, nil, []string{"Island"}),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
