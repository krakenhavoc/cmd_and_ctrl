package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merciless Harlequin — Creature — Human Assassin {2}{B}, 2/1 (Assassin's
// Creed, #5948):
//
//	"Freerunning {1}{B} (You may cast this spell for its freerunning cost
//	 if you dealt combat damage to a player this turn with an Assassin or
//	 commander.)
//	 When this creature enters, you draw a card and you lose 1 life."
//
// Freerunning is the Brotherhood Ambushers alternative cost; the enters
// trigger is a draw then a life loss, with no choice in either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "10e0ccdf-9a31-402a-9fbf-f501ce63fa52",
		Name:             "Merciless Harlequin",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{1}{B}", "Assassin")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merciless Harlequin — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
