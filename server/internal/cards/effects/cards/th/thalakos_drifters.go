package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1afa006b-1190-4047-a9fe-4215f20df8f2",
		Name:         "Thalakos Drifters",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains shadow until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.DiscardACard(),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("shadow", "Thalakos Drifters — shadow"),
		}},
	})
}
