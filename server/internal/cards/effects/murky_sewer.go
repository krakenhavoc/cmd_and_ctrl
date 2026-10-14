package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murky Sewer — Land (Duskmourn Commander):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {U} or {B}."
//
// One of the "unless a player has 13 or less life" taplands. "A
// player" is anyone still in the game, you included, and the life is
// read as the land enters (SelfEntersTappedUnless keeps the condition
// in AppliesTo, so a land that will enter untapped offers no
// replacement to order).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2cdefeb-3176-4faf-be54-a62d31f777a5",
		Name:         "Murky Sewer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(aPlayerHasLifeAtMost(13))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
	})
}

// aPlayerHasLifeAtMost is "a player has N or less life": any player
// still in the game.
func aPlayerHasLifeAtMost(n int) func(g *game.Game, _ uuid.UUID) bool {
	return func(g *game.Game, _ uuid.UUID) bool {
		for _, p := range g.Seats {
			if p != nil && !p.Eliminated && p.Life <= n {
				return true
			}
		}
		return false
	}
}
