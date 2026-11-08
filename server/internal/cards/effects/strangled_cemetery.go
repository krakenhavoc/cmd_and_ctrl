package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strangled Cemetery — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {B} or {G}."
//
// Murky Sewer's tapland in black and green: "a player" is anyone still
// in the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a5b3405-a1e3-4aad-ab4e-1b8db2d1f3a8",
		Name:         "Strangled Cemetery",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
	})
}
