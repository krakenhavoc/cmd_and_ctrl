package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peculiar Lighthouse — Land (Duskmourn: House of Horror, #4466):
//
//	"This land enters tapped unless a player has 13 or less life.
//	 {T}: Add {U} or {R}."
//
// Murky Sewer's tapland in blue and red: "a player" is anyone still in
// the game, you included, and the life is read as the land enters
// (aPlayerHasLifeAtMost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d55f7e20-11c6-44e2-8a21-dca67d3dbc68",
		Name:         "Peculiar Lighthouse",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.APlayerHasLifeAtMost(13))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
	})
}
