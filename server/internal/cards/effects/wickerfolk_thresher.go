package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wickerfolk Thresher — Artifact Creature — Scarecrow {3}{G}, 5/4
// (Duskmourn: House of Horror, #4466):
//
//	"Delirium — Whenever this creature attacks, if there are four or more
//	 card types among cards in your graveyard, look at the top card of
//	 your library. If it's a land card, you may put it onto the
//	 battlefield. If you don't put the card onto the battlefield, put it
//	 into your hand."
//
// Delirium is an ability word (CR 207.2c). The "if" is an intervening if
// (CR 603.4): asked as the Thresher attacks (youHaveDeliriumWhen) and
// again as the trigger resolves (drcDeliriumMet), so a graveyard emptied
// in response does nothing.
//
// The look is Risen Reef's: only the controller sees the card, a land may
// be put onto the battlefield (untapped, as printed, and a put, not a
// land drop, CR 305.4), and the card that did not go to the battlefield,
// a nonland, a declined land or one a replacement kept off, goes into
// the hand through the shared rest-to-hand tail.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa6d7d68-34e3-4ff6-ae6c-36848974555a",
		Name:         "Wickerfolk Thresher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, AllOf(ThisAttacked, youHaveDeliriumWhen),
				"Wickerfolk Thresher — look at the top card; you may put a land onto the battlefield, otherwise put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !drcDeliriumMet(g, item.Controller) {
						return nil
					}
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Label:    "Wickerfolk Thresher — you may put the land onto the battlefield (otherwise it goes into your hand)",
						ThenTail: putRestIntoHandTail,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
