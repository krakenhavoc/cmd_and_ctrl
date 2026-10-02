package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contaminated Ground — Enchantment — Aura {1}{B}:
//
//	"Enchant land
//	 Enchanted land is a Swamp.
//	 Whenever enchanted land becomes tapped, its controller loses 2 life."
//
// ADR 0109 owner decision 4 (#1881): the static form of CR 305.7,
// effects.SetsBasicLandType, attached to the enchanted land. In layer 4 its
// land types are replaced by a Swamp and every other subtype stays (CR 205.1a),
// it loses the abilities its rules text gives it, keeps any another effect
// granted it (a layer-6 grant lands after the removal), and taps for {B}
// (CR 305.6). Its card types and supertypes are untouched, so an enchanted
// basic land is still basic.
//
// The trigger is "becomes tapped" (CR 701.26a), for mana or any other
// reason, read off the permanent the Aura is attached to; "its
// controller" is the land's controller as the trigger resolves, and the
// loss is a loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55e04860-f4f5-445b-81f2-b500fa9b456a",
		Name:         "Contaminated Ground",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Static: []game.StaticAbility{
			SetsBasicLandType(AttachedToSource, nil, []string{"Swamp"}),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: EnchantedPermanentBecameTapped,
			Key:       "Contaminated Ground — its controller loses 2 life",
			Effect:    EnchantedPermanentsControllerLosesLife(2),
		}},
	})
}
