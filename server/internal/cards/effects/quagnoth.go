package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quagnoth — Creature — Beast {5}{G}, 4/5:
//
//	"Split second
//	 Shroud
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, return it to your hand."
//
// Split second and shroud are printed keywords. The discard line is the
// family's "this card" trigger (ADR 0153 §4), watching from
// wherever the card landed (WhereverDiscarded, #3859) on a discard an opponent's spell or ability caused (the stack item's
// controller, CR 113.8), returning the card at once when the ability
// resolves. It is the same object it was when it triggered, or nothing
// returns (CR 400.7).
//
// No simplification: the discard ability is watched from wherever the
// discard left the card (WhereverDiscarded, #3859), so madness or Rest in
// Peace sending it to exile instead of the graveyard does not stop it.
func init() {
	Register(Spec{
		OracleID:          "9250eb9e-690f-4563-8132-d737f6d22883",
		Name:              "Quagnoth",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"split second", "shroud"},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			WhereverDiscarded(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Quagnoth — discarded: return it to your hand",
				returnDiscardedCardToHand)),
		},
	})
}
