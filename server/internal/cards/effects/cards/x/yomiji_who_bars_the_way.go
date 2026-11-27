package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yomiji, Who Bars the Way — Legendary Creature — Spirit {5}{W}{W}, 4/4 (Betrayers of
// Kamigawa, S82 OLD b595):
//
//	"Whenever a legendary permanent other than Yomiji is put into a graveyard from the
//	 battlefield, return that card to its owner's hand."
//
// Any player's legendary permanent, read as it last existed on the battlefield (CR 603.10a);
// the card is returned from whichever graveyard it went to, and only if it is still there as
// the ability resolves. A legendary token is not a card and returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "438be2fa-5f8d-42ff-894a-671ba4221196",
		Name:         "Yomiji, Who Bars the Way",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB595LegendaryPermanentWasPutIntoAGraveyard,
				"Yomiji — return that card to its owner's hand", effects.OldB595ReturnThatCardToItsOwnersHand),
		},
	})
}
