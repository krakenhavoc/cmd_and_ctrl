package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavaron Skywarden — Creature — Kavu Soldier {4}{R}, 4/5 (Edge of
// Eternities):
//
//	"Reach
//	 Void — At the beginning of your end step, if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn, put
//	 a +1/+1 counter on this creature."
//
// Void is an intervening if (void.go). Reach is printed and read from the
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c1f1ba17-b5e8-4bb3-a767-befd52e2db72",
		Name:            "Kavaron Skywarden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfVoid("Kavaron Skywarden — put a +1/+1 counter on this creature", effects.PutACounterOnThis("+1/+1")),
		},
	})
}
