package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempting Licid — Creature — Licid {2}{G}, 2/2 (Stronghold, S82 OLD batch 746, #5627):
//
//	"{G}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {G} to end this effect.
//	 All creatures able to block enchanted creature do so."
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability,
// LicidForm what it turns the permanent into, and ReturnedAura the "enchant
// creature" clause. "All creatures able to block enchanted creature do so" is Lure's block requirement on the enchanted creature. Paying to end the effect is the engine's end_effect
// special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b2206ed-55e9-4735-b969-59d9e089bb10",
		Name:         "Tempting Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.Labelled(effects.BlockRequirementWhere(game.BlockRequirementLure, effects.AttachedToSource), "All creatures able to block enchanted creature do so.")),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{G}", "{G}")},
	})
}
