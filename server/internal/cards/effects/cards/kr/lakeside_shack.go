package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c56cd2ec-5907-4282-9162-d93b7dfd63b5",
		Name:         "Lakeside Shack",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.APlayerHasLifeAtMost(13))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
	})
}
