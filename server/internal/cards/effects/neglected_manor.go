package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neglected Manor — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {W} or {B}."
//
// Murky Sewer's tapland in white and black: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8c632ab-14ec-44e1-ac00-81d48336320d",
		Name:         "Neglected Manor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
	})
}
