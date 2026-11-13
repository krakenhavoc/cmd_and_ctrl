package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marker Beetles — Creature — Insect {1}{G}{G}, 2/3 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, target creature gets +1/+1 until end of
//	 turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// The shared udsB2WhenThisDiesTargetCreatureGets at +1/+1 and the
// shared udsB2SacrificeThisDrawRow. Sacrificing the Beetles to draw is
// a death, so the dies trigger fires too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b1614fd-2810-4d4d-a0d6-934f27adbe65",
		Name:         "Marker Beetles",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			udsB2WhenThisDiesTargetCreatureGets("Marker Beetles", 1, 1),
		},
		Activated: []ActivatedAbility{udsB2SacrificeThisDrawRow("creature")},
	})
}
