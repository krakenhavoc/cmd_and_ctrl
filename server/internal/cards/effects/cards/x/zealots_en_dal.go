package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealots en-Dal — Creature — Human Soldier {3}{W}, 2/4 (Exodus, S82 OLD batch 743, #5626):
//
//	"At the beginning of your upkeep, if all nonland permanents you control
//	 are white, you gain 1 life."
//
// An intervening "if" (CR 603.4): checked as the upkeep begins and again as the
// ability resolves. The colors are read after the layers, so a multicolored
// permanent that is white counts as white, and a colorless one does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fa67c92-6328-4d0a-95e3-656d447ab20e",
		Name:         "Zealots en-Dal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("all nonland permanents you control are white", effects.S82OldB743AllNonlandPermanentsYouControlAreWhite,
				effects.AtYourUpkeep("Zealots en-Dal — you gain 1 life", effects.Do(effects.GainLife{Amount: 1}))),
		},
	})
}
