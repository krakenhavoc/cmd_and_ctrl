package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helionaut — Creature — Human Soldier {2}{W}, 1/2 (Apocalypse, S82 OLD, #5614):
//
//	"Flying
//	 {1}, {T}: Add one mana of any color."
//
// A printed keyword and Rootrider Faun's filter mana ability (CR 605.1a); {T} is
// in the cost, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d4de785-20e9-430e-8df1-f39dac0ef07d",
		Name:            "Helionaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities:   []effects.ManaAbility{effects.OldB632DiscipleAnyColorMana("{1}")},
	})
}
