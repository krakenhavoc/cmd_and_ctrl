package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lakeside Shack — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {G} or {U}."
//
// Murky Sewer's tapland in green and blue: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c56cd2ec-5907-4282-9162-d93b7dfd63b5",
		Name:         "Lakeside Shack",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
	})
}
