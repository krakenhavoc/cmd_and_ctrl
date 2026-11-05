package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "76f52fad-8db2-42f0-8e12-f4341e4c8680",
		Name:         "Interstellar Effigy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Interstellar Effigy — return target creature an opponent controls to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return bounceTheTarget(item, NewContext(g, item))
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
