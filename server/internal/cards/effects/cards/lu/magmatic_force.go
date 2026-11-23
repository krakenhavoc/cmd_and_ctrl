package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magmatic Force — Creature — Elemental {5}{R}{R}{R}, 7/7 (Commander Starter
// Decks):
//
//	"At the beginning of each upkeep, this creature deals 3 damage to any
//	 target."
//
// Verdant Force's every-upkeep trigger with a target clause chosen as the
// ability goes on the stack and re-checked as it resolves (CR 608.2b); the
// damage is the Force's own, and nothing is dealt if the target has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea73153b-1ecb-4593-a9ec-ce4f54e69824",
		Name:         "Magmatic Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginUpkeep
			}, "Magmatic Force — 3 damage to any target", effects.B33DamageChosenTargetFromSource(3)), effects.TargetAny()),
		},
	})
}
