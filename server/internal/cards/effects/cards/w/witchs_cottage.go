package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Cottage — Land — Swamp:
//
//	"({T}: Add {B}.)
//	 This land enters tapped unless you control three or more other
//	 Swamps.
//	 When this land enters untapped, you may put target creature card
//	 from your graveyard on top of your library."
//
// Mystic Sanctuary's shape with Swamps and a creature card. The {B}
// ability is derived from the Swamp land type (CR 305.6), so none is
// declared. The trigger is optional and targets, so an empty graveyard
// means no trigger (CR 603.3d); the put goes through the shared exit
// primitive, so a commander card gets the CR 903.9 offer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c8f276e-4e7b-4974-ab02-9356cc0ffb2b",
		Name:         "Witch's Cottage",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnless(effects.OtherLandsWithSubtypeAtLeast("swamp", 3)),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventETB, effects.SelfEnteredUntapped,
						"Witch's Cottage — put a creature card on top of your library",
						effects.PutChosenTargetOnTopOfLibrary),
					"Witch's Cottage — put target creature card from your graveyard on top of your library?"),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
		},
	})
}
