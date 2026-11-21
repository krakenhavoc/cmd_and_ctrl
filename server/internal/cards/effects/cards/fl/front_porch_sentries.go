package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Front Porch Sentries — Creature — Goblin Soldier {1}{B}, 2/2 (The Hobbit):
//
//	"When this creature dies, target creature an opponent controls gets
//	 -1/-1 until end of turn."
//
// A dies trigger with a target chosen as it goes on the stack (CR 603.3d),
// limited to creatures an opponent controls. The shrink resolves on
// whatever it is by then; a target that has left or become illegal makes
// the whole ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b39f03cf-3d1a-42ad-903f-e3b3542878b0",
		Name:         "Front Porch Sentries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Front Porch Sentries — target creature an opponent controls gets -1/-1 until end of turn",
				effects.BoostTargetUntilEOT("Front Porch Sentries — -1/-1", -1, -1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
