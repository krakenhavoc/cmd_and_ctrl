package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Thug — Creature — Human Warrior {1}{B}, 1/1:
//
//	"When this creature dies, put target creature card from your
//	 graveyard on top of your library.
//	 Dredge 4 (If you would draw a card, you may mill four cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// The dies trigger is Mortuary Mire's put (the shared tuck, so a
// commander card gets the CR 903.9 offer on the way). It goes on the
// stack with the Thug already in the graveyard, so it may target
// itself, which is the classic loop: put the Thug on top, draw it.
// Dredge 4 is dredge.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a426a258-fd8b-489c-8642-9868ee47de85",
		Name:         "Golgari Thug",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Dredge(4)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Golgari Thug — put target creature card from your graveyard on top of your library",
					effects.PutChosenTargetOnTopOfLibrary),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
		},
	})
}
