package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Convalescence — Enchantment {1}{W} (Exodus, S82 OLD batch 743, #5626):
//
//	"At the beginning of your upkeep, if you have 10 or less life, you gain 1
//	 life."
//
// An intervening "if" (CR 603.4): checked as the upkeep begins and again as the
// ability resolves, so a player at 11 or more by then gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ad86d0e-3363-4d98-b0bb-dc491be1ddd5",
		Name:         "Convalescence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have 10 or less life", effects.S82OldB743LifeAtMostTen,
				effects.AtYourUpkeep("Convalescence — you gain 1 life", effects.Do(effects.GainLife{Amount: 1}))),
		},
	})
}
