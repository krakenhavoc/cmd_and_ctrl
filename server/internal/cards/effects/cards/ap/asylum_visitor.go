package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Asylum Visitor — Creature — Vampire Wizard {1}{B}, 3/1 (Innistrad Remastered,
// #5948):
//
//	"At the beginning of each player's upkeep, if that player has no cards in
//	 hand, you draw a card and you lose 1 life.
//	 Madness {1}{B}"
//
// Ghirapur Orrery's trigger: each upkeep is a trigger, with an intervening "if"
// on the hand of the player whose upkeep it is (CR 603.4), read as it triggers
// and as it resolves. The beneficiary is the Visitor's controller, whoever's
// upkeep it is, and the life is lost with the draw. Madness is the printed
// keyword's alternative cost (CR 702.35).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "870deab4-8570-40e9-aeda-ac3726cebb84",
		Name:         "Asylum Visitor",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{B}",
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.Cmdr2B01CondActorHandEmpty{},
				effects.AtEachUpkeep("Asylum Visitor — you draw a card and you lose 1 life",
					effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1}))),
		},
	})
}
