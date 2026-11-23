package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortuary Mire — Land:
//
//	"This land enters tapped.
//	 When this land enters, you may put target creature card from your
//	 graveyard on top of your library.
//	 {T}: Add {B}."
//
// The reanimator's rebuy land — same trigger shape Mystic Sanctuary
// uses one instant/sorcery over: a "you may" (Optional) that TARGETS
// (Targeting), so an empty graveyard means no trigger at all
// (CR 603.3d) and a creature that leaves the graveyard in response
// makes it fizzle (CR 608.2b). The put goes through the shared
// TuckToLibraryForEffect exit primitive, so a commander card in the
// graveyard gets the CR 903.9 offer on the way to the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b3fb20a-e090-4286-9c03-6b71c27c45be",
		Name:         "Mortuary Mire",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventETB, effects.Self,
						"Mortuary Mire — put a creature card on top of your library",
						effects.PutChosenTargetOnTopOfLibrary),
					"Mortuary Mire — put target creature card from your graveyard on top of your library?"),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
		},
	})
}
