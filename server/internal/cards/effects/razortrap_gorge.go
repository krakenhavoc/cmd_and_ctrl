package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razortrap Gorge — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {B} or {R}."
//
// Murky Sewer's tapland in black and red: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f69bd3a-244e-42d8-bfac-5a426f4b54b4",
		Name:         "Razortrap Gorge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
	})
}
