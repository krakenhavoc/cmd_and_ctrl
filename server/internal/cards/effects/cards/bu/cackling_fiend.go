package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cackling Fiend — Creature — Zombie {2}{B}{B}, 2/1 (Urza's Saga, #5624):
//
//	"When this creature enters, each opponent discards a card."
//
// Burglar Rat's enters trigger (eachOpponentDiscardsOne): one discard
// prompt per opponent, each addressed to that opponent and offering only
// their own hand (CR 701.9a). An opponent with no cards in hand is
// skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2029954b-6fa7-40d7-bb19-d7534c62be5d",
		Name:         "Cackling Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cackling Fiend — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
