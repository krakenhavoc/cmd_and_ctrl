package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bleeding Woods — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {R} or {G}."
//
// Murky Sewer's tapland in red and green: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47b6d2ae-d3d7-41eb-9172-2076eb8d028d",
		Name:         "Bleeding Woods",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
	})
}
