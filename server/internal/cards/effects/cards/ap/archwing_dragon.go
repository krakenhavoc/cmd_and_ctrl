package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archwing Dragon — {2}{R}{R} Creature — Dragon 4/4: "Flying, haste. At the
// beginning of the end step, return this creature to its owner's hand."
//
// The trigger fires at every player's end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0a706b2-b237-4183-82fa-2140a23b89e3",
		Name:            "Archwing Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered:       []game.TriggeredAbility{effects.UlgB1ReturnAtEachEndStep("Archwing Dragon")},
	})
}
