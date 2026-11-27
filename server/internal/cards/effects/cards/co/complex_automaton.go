package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Complex Automaton — Artifact Creature — Golem {4}, 4/4 (Nemesis, S82 OLD
// batch 680, #5618):
//
//	"At the beginning of your upkeep, if you control seven or more
//	 permanents, return this creature to its owner's hand."
//
// An intervening "if" (CR 603.4): it triggers only with seven or more
// permanents under your control, tokens and the Automaton itself included,
// and does nothing if you have fewer as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4951b4e-ba68-495c-8062-e89250d02cf3",
		Name:         "Complex Automaton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control seven or more permanents",
				func(g *game.Game, you uuid.UUID) bool {
					return effects.CountControlled(g, you, func(game.Card) bool { return true }) >= 7
				},
				effects.AtYourUpkeep("Complex Automaton — return this creature to its owner's hand",
					effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)))),
		},
	})
}
