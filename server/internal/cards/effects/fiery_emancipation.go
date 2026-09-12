package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Emancipation — Enchantment {3}{R}{R}{R} (EDHREC rank 855):
//
//	"If a source you control would deal damage to a permanent or
//	 player, it deals triple that damage to that permanent or player
//	 instead."
//
// Every point of damage you deal, tripled — combat, burn, pingers,
// your own painlands hurting you. Angrath's Marauders' replacement
// with a three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52159875-354c-47f9-bb1c-cd65395fcc68",
		Name:         "Fiery Emancipation",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b07TripleDamageFromYourSources()},
	})
}
