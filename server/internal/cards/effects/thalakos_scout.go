package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalakos Scout — Creature — Thalakos Soldier Scout {2}{U}, 2/1
// (Exodus, #5626):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 Discard a card: Return this creature to its owner's hand."
//
// Shadow is a printed keyword (CR 702.28b, enforced in block legality).
// The bounce is Bramble Familiar's return with only the discard as its
// cost: the discard is paid as the ability is activated (CR 602.2b,
// 601.2h), so it needs a card in hand, and the shared
// returnThisPermanentToOwnersHand does nothing if the Scout has already
// left or come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8877da5-7008-4b47-955e-d205220c91e6",
		Name:            "Thalakos Scout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    DiscardACard(),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
