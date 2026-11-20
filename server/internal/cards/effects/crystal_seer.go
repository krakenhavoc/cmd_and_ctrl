package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Seer — Creature — Vedalken Wizard {4}{U}, 2/2 (Guildpact,
// #5599):
//
//	"When this creature enters, look at the top four cards of your
//	 library, then put them back in any order.
//	 {4}{U}: Return this creature to its owner's hand."
//
// Sage Owl's enters trigger (LookAtTop) and Aethertide Whale's bounce
// of itself: the ability returns this permanent while it is still the
// same object, so a Seer that left and came back is not returned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58838588-13d5-4948-babf-d09cc59d76a1",
		Name:         "Crystal Seer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crystal Seer — look at the top four cards of your library, then put them back in any order", Do(LookAtTop{N: 4})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{4}{U}"),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
