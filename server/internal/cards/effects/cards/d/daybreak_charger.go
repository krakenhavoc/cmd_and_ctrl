package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daybreak Charger — Creature — Unicorn {1}{W}, 3/1 (Jumpstart):
//
//	"When this creature enters, target creature gets +2/+0 until end of
//	 turn."
//
// Soldier of the Grey Host's entry pump: the target is any creature,
// the Charger itself included, picked as the trigger goes on the stack
// and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3de6b36-c631-4b52-b16b-0749a347d73d",
		Name:         "Daybreak Charger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Daybreak Charger — target creature gets +2/+0 until end of turn",
				effects.BoostTargetUntilEOT("Daybreak Charger — +2/+0", 2, 0)),
				effects.TargetCreature("target creature")),
		},
	})
}
