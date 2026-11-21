package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Queza, Augur of Agonies — Legendary Creature — Octopus Advisor
// {1}{W}{U}{B}, 3/4:
//
//	"Whenever you draw a card, target opponent loses 1 life and you
//	 gain 1 life."
//
// EventDrawCard fires once per card with the drawer in Actor (Sphinx's
// Tutelage's shape), so a draw-three is three triggers, each choosing
// its own opponent as it goes on the stack. The drain is the shared
// drainTargetOpponentOne: loss, not damage, then the gain; a target that
// has left the game makes the whole ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c983338d-ae6b-4a15-9931-daa20fca8269",
		Name:         "Queza, Augur of Agonies",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDrawCard, effects.ByYou,
					"Queza, Augur of Agonies — target opponent loses 1 life and you gain 1 life",
					effects.DrainTargetOpponentOne),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
