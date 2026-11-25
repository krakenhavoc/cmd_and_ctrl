package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grasping Thrull — Creature — Thrull {3}{W}{B}, 3/3 (Ravnica Allegiance):
//
//	"Flying
//	 When this creature enters, it deals 2 damage to each opponent and you gain 2
//	 life."
//
// The damage to every opponent is one instance; the life gain follows in the same
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec8a14d4-6ca0-43b1-abf6-0c129daf8350",
		Name:            "Grasping Thrull",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grasping Thrull — 2 damage to each opponent and you gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.DamageToEachOpponent(g, item, 2); err != nil {
						return err
					}
					return effects.GainLife{Amount: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
