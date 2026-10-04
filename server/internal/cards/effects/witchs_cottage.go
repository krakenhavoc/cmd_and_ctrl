package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6c8f276e-4e7b-4974-ab02-9356cc0ffb2b",
		Name:         "Witch's Cottage",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnless(otherLandsWithSubtypeAtLeast("swamp", 3)),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventETB, selfEnteredUntapped,
						"Witch's Cottage — put a creature card on top of your library",
						PutChosenTargetOnTopOfLibrary),
					"Witch's Cottage — put target creature card from your graveyard on top of your library?"),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			),
		},
	})
}
