package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stratus Walk — Enchantment — Aura {1}{U} (Magic Origins):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has flying.
//	 Enchanted creature can block only creatures with flying."
//
// Angelic Gift's cantrip and flying joined to Air Bladder's block rule; both
// follow the Aura off the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eda37caa-d676-460a-804b-3ba9d8a8d044",
		Name:         "Stratus Walk",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		BlockRules:   effects.NemBlocksOnlyFlyers(effects.OnAttached()),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stratus Walk — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
