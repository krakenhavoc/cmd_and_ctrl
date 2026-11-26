package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Arcanist — Creature — Merfolk Wizard {1}{U}, 1/3 (Core Set 2021,
// #5511):
//
//	"{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell."
//
// Ore-Rich Stalactite's restricted mana: spendable only on casting an instant
// or sorcery spell, not on abilities or other spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d07856b-d9ef-418c-bd06-3888dbcaddf4",
		Name:         "Vodalian Arcanist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}",
			Label:        "Add {C}. Spend this mana only to cast an instant or sorcery spell",
			Restrictions: []string{effects.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
		}},
	})
}
