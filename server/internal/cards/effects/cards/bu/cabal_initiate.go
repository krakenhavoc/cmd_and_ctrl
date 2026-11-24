package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Initiate — Creature — Human Warlock {1}{B}, 2/1 (Modern Horizons 2):
//
//	"Discard a card: This creature gains lifelink until end of turn.
//	 Threshold — This creature gets +1/+2 as long as there are seven or more
//	 cards in your graveyard."
//
// The discard is a cost, paid up front, so the discarded card counts toward
// threshold at once. Threshold is a layer 7c modification read live on every
// recompute (Billowing Shriekmass's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc19fa3-c6c2-42c0-85fa-8d13e1baa145",
		Name:         "Cabal Initiate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.DiscardACard(),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Cabal Initiate — lifelink until end of turn"),
		}},
		Static: []game.StaticAbility{effects.FraThresholdSelfPT(1, 2)},
	})
}
