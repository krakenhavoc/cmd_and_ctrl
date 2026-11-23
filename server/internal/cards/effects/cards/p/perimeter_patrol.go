package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perimeter Patrol — Creature — Human Soldier {2}{G}, 3/3 (The Brothers'
// War, #5482):
//
//	"Whenever an artifact you control enters, this creature gets +1/+0
//	 until end of turn."
//
// Phalanx Vanguard's trigger (s82BroB07PumpSelfOnArtifact).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef625caf-c5d3-47f4-b1fb-bd040c5d864c",
		Name:         "Perimeter Patrol",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82BroB07PumpSelfOnArtifact("Perimeter Patrol")},
	})
}
