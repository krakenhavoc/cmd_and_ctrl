package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curious Cadaver — Creature — Zombie Detective {2}{U}{B}, 3/1 (Murders
// at Karlov Manor, #4469):
//
//	"Flying
//	 When you sacrifice a Clue, return this card from your graveyard to
//	 your hand."
//
// The second ability works only from the graveyard (CR 113.6k), as A
// Good Day to Pie's does. "You sacrifice a Clue" is Tireless Tracker's
// (lonisYouSacrificedAClue: the sacrifice event is emitted before the
// Clue leaves, so its subtype is read on the battlefield), and any Clue
// counts. In the graveyard the card's controller is its owner, so "you"
// is the owner. The return is not optional, and it does nothing if the
// card has left the graveyard by the time the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cc5370e0-a0e6-4958-a6f2-093a3868b575",
		Name:            "Curious Cadaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Curious Cadaver — return this card from your graveyard to your hand",
				returnThisFromYourGraveyardToHand)),
		},
	})
}
