package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narcomoeba — 1/1 Creature — Illusion for {1}{U}:
//
//	"Flying
//	 When this card is put into your graveyard from your library, you
//	 may put it onto the battlefield."
//
// The plainest triggered ability in Magic that does not work from
// the battlefield. The card is in the graveyard when the ability
// triggers — that is the only place it can be, because being put
// there is the trigger condition — so it is the smallest honest
// proof of #925's Zones dimension.
//
// "You" is the card's OWNER (CR 108.4): a Narcomoeba has no
// controller while it sits in a graveyard, so the "you may" is
// offered to the player whose library it came off, which is also
// whose graveyard it is in.
//
// CR 603.5: the "you may" is a prompt at trigger time, not a choice
// at resolution, and declining drops the trigger. It goes on the
// stack either way, so the table can respond to the return before it
// happens.
func init() {
	Register(Spec{
		OracleID:        "dc65bb62-4ba2-4ec1-b3b9-9c51e64cbfc8",
		Name:            "Narcomoeba",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisIsPutIntoYourGraveyardFromYourLibrary(
				"Narcomoeba — put it onto the battlefield",
				returnThisCardFromYourGraveyard), "Narcomoeba — put it onto the battlefield?"),
		},
	})
}
