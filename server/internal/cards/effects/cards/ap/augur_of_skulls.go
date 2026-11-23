package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Augur of Skulls — Creature — Skeleton Wizard {1}{B}, 1/1 (Future Sight, #5594):
//
//	"{1}{B}: Regenerate this creature.
//	 Sacrifice this creature: Target player discards two cards. Activate
//	 only during your upkeep."
//
// The shared regeneration row, and Mind Rot's body on an upkeep-only
// sacrifice (DuringYourUpkeep is the step AND your turn): the targeted
// player picks the two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "723398f2-eaf2-4615-9078-418b9f540a33",
		Name:         "Augur of Skulls",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.RegenerateThisRow("{1}{B}"),
			{
				Label:     "Sacrifice this creature: Target player discards two cards. Activate only during your upkeep.",
				Cost:      effects.SacrificeThis(),
				Condition: effects.DuringYourUpkeep(),
				Targets:   effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MindRotOnResolve(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
