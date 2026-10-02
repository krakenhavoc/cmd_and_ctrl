package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stingscourger — Creature — Goblin Warrior, {1}{R}, 2/2:
//
//	"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, return target creature an opponent controls to its owner's hand."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e8d1051-1739-4a10-9cef-78f7333b4979",
		Name:         "Stingscourger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Stingscourger", "{3}{R}"),
			Targeting(WhenThisEnters("Stingscourger — return target creature an opponent controls to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return bounceTheTarget(item, NewContext(g, item))
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
