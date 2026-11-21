package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spreading Seas — Enchantment — Aura {1}{U}:
//
//	"Enchant land
//	 When this Aura enters, draw a card.
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
		OracleID:     "af32757c-9d3c-4f6c-a6fa-d5729a440382",
		Name:         "Spreading Seas",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static: []game.StaticAbility{
			effects.SetsBasicLandType(effects.AttachedToSource, nil, []string{"Island"}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spreading Seas — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
