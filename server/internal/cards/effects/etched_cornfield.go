package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etched Cornfield — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {G} or {W}."
//
// Murky Sewer's tapland in green and white: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ccca5c2-66c3-495a-8d9e-1a9805569e52",
		Name:         "Etched Cornfield",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
	})
}
