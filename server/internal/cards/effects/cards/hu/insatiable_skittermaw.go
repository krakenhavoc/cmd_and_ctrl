package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insatiable Skittermaw — Creature — Insect Horror {2}{B}, 2/2 (Edge of
// Eternities):
//
//	"Menace
//	 Void — At the beginning of your end step, if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn, put
//	 a +1/+1 counter on this creature."
//
// Void is an intervening if (void.go). Menace is printed and read from the
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69a68dec-4b01-4528-81fc-a3c5f5097f72",
		Name:            "Insatiable Skittermaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfVoid("Insatiable Skittermaw — put a +1/+1 counter on this creature", effects.PutACounterOnThis("+1/+1")),
		},
	})
}
