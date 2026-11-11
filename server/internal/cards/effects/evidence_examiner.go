package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evidence Examiner — Creature — Merfolk Detective {G}{U}, 2/2 (Murders at
// Karlov Manor, #5223):
//
//	"At the beginning of combat on your turn, you may collect evidence 4.
//	 (Exile cards with total mana value 4 or greater from your graveyard.)
//	 Whenever you collect evidence, investigate. (Create a Clue token. It's
//	 an artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// The combat trigger is the effect form of collect evidence 4
// (RegisterCollectEvidence) with nothing after it; a graveyard holding less
// than 4 in all asks nothing (CR 701.59b). "Whenever you collect evidence"
// is WheneverYouCollectEvidence, off game.EventCollectEvidence, so it sees
// every collection its controller makes, not only this one (the card's
// ruling). Investigate is creating a Clue (CR 701.16a, 111.10f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3468804-26ad-4919-b030-a1906a7029e8",
		Name:         "Evidence Examiner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Evidence Examiner — you may collect evidence 4",
				func(g *game.Game, item *game.StackItem) error {
					return evidenceExaminerEvidence.Ask(NewContext(g, item), "Evidence Examiner — collect evidence 4 (exile cards with total mana value 4 or more from your graveyard)? Choose none to decline.")
				}),
			WheneverYouCollectEvidence("Evidence Examiner — investigate", Do(Investigate{N: 1})),
		},
	})
}

// evidenceExaminerEvidence is "you may collect evidence 4." with nothing
// after it. An on-disk key: never renamed, never reused.
var evidenceExaminerEvidence = RegisterCollectEvidence("choose-cards/evidence-examiner-collect-evidence", 4, nil)
