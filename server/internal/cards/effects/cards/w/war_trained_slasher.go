package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War-Trained Slasher — Creature — Wolverine Dinosaur {3}{R}, 4/3 (March
// of the Machine, #5649):
//
//	"Menace
//	 Whenever this creature attacks a battle, double its power until end
//	 of turn."
//
// Triggers when it is declared as an attacker attacking a battle (CR
// 508.3a). Doubling is +X/+0 where X is its power as the ability resolves
// (CR 701.10b), so a pump in response is doubled too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc89face-92b1-47d3-8cdb-360b7334d3b8",
		Name:            "War-Trained Slasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacksABattle("War-Trained Slasher — double its power until end of turn",
				effects.Cmdr2B23DoubleThisPower("War-Trained Slasher — double power")),
		},
	})
}
