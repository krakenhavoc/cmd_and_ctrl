package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Candlelit Cavalry — Creature — Human Knight {4}{G}, 5/5 (Midnight Hunt,
// #5497):
//
//	"Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, this creature gains
//	 trample until end of turn."
//
// Candlegrove Witch's trigger with trample (midB05CovenThisGainsKeyword).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edfd2614-8875-4741-b64b-390afc0d2376",
		Name:         "Candlelit Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MidB05CovenThisGainsKeyword("Candlelit Cavalry", "trample")},
	})
}
