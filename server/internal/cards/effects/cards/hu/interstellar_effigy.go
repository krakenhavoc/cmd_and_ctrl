package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Interstellar Effigy — Artifact Creature — Construct {4}{U}, 4/3 (Star
// Trek, #4973):
//
//	"When this creature enters, return target creature an opponent
//	 controls to its owner's hand."
//
// Stingscourger's enters trigger without the echo: a mandatory target,
// so the trigger is removed when no opponent controls a creature
// (CR 603.3d), and a target gone by resolution makes it do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76f52fad-8db2-42f0-8e12-f4341e4c8680",
		Name:         "Interstellar Effigy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Interstellar Effigy — return target creature an opponent controls to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BounceTheTarget(item, effects.NewContext(g, item))
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
