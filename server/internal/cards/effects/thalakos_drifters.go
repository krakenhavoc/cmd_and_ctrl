package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalakos Drifters — Creature — Thalakos {2}{U}{U}, 3/3
// (Exodus, #5626):
//
//	"Discard a card: This creature gains shadow until end of turn."
//
// Thalakos Scout's discard cost (CR 602.2b) on the shared
// thisGainsKeywordUntilEndOfTurn grant, which ends in the cleanup step
// (CR 514.2). Shadow is enforced in block legality (CR 702.28b), so
// gained after blockers are declared it doesn't undo a block already
// made (CR 509.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1afa006b-1190-4047-a9fe-4215f20df8f2",
		Name:         "Thalakos Drifters",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gains shadow until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    DiscardACard(),
			Effect:  thisGainsKeywordUntilEndOfTurn("shadow", "Thalakos Drifters — shadow"),
		}},
	})
}
