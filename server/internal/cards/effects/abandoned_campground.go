package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abandoned Campground — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {W} or {U}."
//
// Murky Sewer's tapland in white and blue: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0eec9984-cd11-4a52-9234-469c6a5fb9aa",
		Name:         "Abandoned Campground",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
	})
}
