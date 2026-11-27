package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brutal Nightstalker — Creature — Nightstalker {3}{B}{B}, 3/2 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"When this creature enters, you may have target opponent discard a card."
//
// The opponent is the target; the "may" is asked as the trigger resolves (CR 603.5) and the opponent chooses the card they discard (CR 701.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4003cd1-e98c-4346-b685-89c7891f9920",
		Name:         "Brutal Nightstalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.WhenThisEnters("Brutal Nightstalker — you may have target opponent discard a card", effects.S82OldB734DiscardTargetPlayer),
					"Brutal Nightstalker — have target opponent discard a card?"),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
