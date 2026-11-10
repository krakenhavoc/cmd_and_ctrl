package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erdwal Illuminator — Creature — Spirit {1}{U}, 1/3 (Murders at Karlov
// Manor Commander, #5425):
//
//	"Flying
//	 Whenever you investigate for the first time each turn, investigate an
//	 additional time."
//
// The trigger watches game.EventInvestigate, which only the keyword action
// emits (CR 701.16a): a card that says "create a Clue token" without the
// word does not investigate. "For the first time each turn" is the
// player's turn tally, so the investigate this ability makes is the
// turn's second and does not trigger it again, two Illuminators each
// trigger once on the first, and one that enters after your first
// investigate of the turn waits for the next turn. "Investigate twice" is
// two investigates and one trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ff4ef34-acb9-46c3-8530-f4b093807066",
		Name:            "Erdwal Illuminator",
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouInvestigateForTheFirstTimeEachTurn("Erdwal Illuminator — investigate an additional time",
				Do(Investigate{N: 1})),
		},
	})
}
