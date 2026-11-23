package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8f69bd3a-244e-42d8-bfac-5a426f4b54b4",
		Name:         "Razortrap Gorge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.APlayerHasLifeAtMost(13))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
	})
}
