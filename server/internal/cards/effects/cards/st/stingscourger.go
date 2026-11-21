package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingscourger — Creature — Goblin Warrior, {1}{R}, 2/2:
//
//	"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, return target creature an opponent controls to its owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e8d1051-1739-4a10-9cef-78f7333b4979",
		Name:         "Stingscourger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Stingscourger", "{3}{R}"),
			effects.Targeting(effects.WhenThisEnters("Stingscourger — return target creature an opponent controls to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BounceTheTarget(item, effects.NewContext(g, item))
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
