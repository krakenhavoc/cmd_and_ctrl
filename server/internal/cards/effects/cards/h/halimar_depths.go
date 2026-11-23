package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halimar Depths — Land:
//
//	"This land enters tapped.
//	 When this land enters, look at the top three cards of your library,
//	 then put them back in any order.
//	 {T}: Add {U}."
//
// Portent's look-and-reorder (LookAtLibraryThenPlace, ADR 0088) as an
// enters trigger on a tapland.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42d121a2-5266-483a-ab16-e0a8073cd6a3",
		Name:         "Halimar Depths",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Halimar Depths — look at the top three cards of your library, then put them back in any order",
				effects.Do(effects.LookAtLibraryThenPlace{
					N:         3,
					Placement: game.LibraryPlaceTop,
					Label:     "Halimar Depths — put them back in any order",
				})),
		},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("U")},
	})
}
