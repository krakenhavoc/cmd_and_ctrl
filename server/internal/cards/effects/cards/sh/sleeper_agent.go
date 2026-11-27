package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sleeper Agent — Creature — Phyrexian Minion 3/3 {B} (Urza's Saga, S82 OLD b723):
//
//	When this creature enters, target opponent gains control of it.
//	At the beginning of your upkeep, this creature deals 2 damage to you.
//
// The enter trigger targets an opponent, who gains control of it indefinitely while it is the same permanent (CR 400.7). "Your upkeep" is its current controller's, so the damage hits whoever has it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "627c9546-85b0-4b4d-93c6-e2ca8c3adf4f",
		Name:         "Sleeper Agent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sleeper Agent — target opponent gains control of it",
				effects.OldB586TargetOpponentGainsControlOfThis("Sleeper Agent — gains control")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.AtYourUpkeep("Sleeper Agent — 2 damage to you", func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
