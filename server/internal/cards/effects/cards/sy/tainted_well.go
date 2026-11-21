package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Well — Enchantment — Aura {2}{B}:
//
//	"Enchant land
//	 When this Aura enters, draw a card.
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
	effects.Register(effects.Spec{
		OracleID:     "b35b3ea6-ae9c-4128-a47f-069d88ee49c2",
		Name:         "Tainted Well",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static: []game.StaticAbility{
			effects.SetsBasicLandType(effects.AttachedToSource, nil, []string{"Swamp"}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tainted Well — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
