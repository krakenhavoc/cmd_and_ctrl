package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea's Presence — Enchantment — Aura {1}{G} (Theros, S82 OLD b390, #5561):
//
//	"Enchant land
//	 When this Aura enters, draw a card.
//	 Enchanted land is every basic land type in addition to its other types."
//
// Spreading Seas' enters trigger beside Dryad of the Ilysian Grove's layer 4
// grant of all five basic land types, which appends the types the land lacks
// and leaves the others in place (CR 305.7 does not apply: the land keeps its
// own abilities); the land gains each type's intrinsic mana ability
// (CR 305.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93fe5eed-520f-467a-8e27-6f40eaa727e8",
		Name:         "Nylea's Presence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static: []game.StaticAbility{{
			Label:     "Enchanted land is every basic land type in addition to its other types.",
			Layer:     game.Layer4Type,
			AppliesTo: effects.AttachedToSource,
			Apply:     effects.DryadOfTheIlysianGroveGrantAllBasicTypes,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nylea's Presence — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
