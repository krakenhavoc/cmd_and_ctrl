package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Simulacrum — Artifact Creature — Soldier {4}, 2/4 (Ice Age,
// #5636):
//
//	"Cumulative upkeep {1}
//	 {1}: This creature gets +1/+0 until end of turn."
//
// Cumulative upkeep is the shared CumulativeUpkeep trigger (CR 702.24a):
// {1} per age counter each of your upkeeps, or it is sacrificed. The
// pump is the shared selfPumpRow at +1/+0 for {1}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "227ed0b7-7a9a-41f1-a190-1d1b85a44720",
		Name:         "Soldevi Simulacrum",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Soldevi Simulacrum — cumulative upkeep {1}", "{1}"),
		},
		Activated: []effects.ActivatedAbility{effects.SelfPumpRow("Soldevi Simulacrum", "{1}", 1, 0)},
	})
}
