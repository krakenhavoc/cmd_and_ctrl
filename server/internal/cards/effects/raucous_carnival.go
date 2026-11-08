package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raucous Carnival — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {R} or {W}."
//
// Murky Sewer's tapland in red and white: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24a97436-ba61-4ebc-a560-a6c027ccfdf3",
		Name:         "Raucous Carnival",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
	})
}
