package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illusory Ambusher — Creature — Cat Illusion {4}{U}, 4/1 (Lost Caverns of
// Ixalan Commander):
//
//	"Flash
//	 Whenever this creature is dealt damage, draw that many cards."
//
// The trigger waits for the damage to finish and draws the TOTAL the
// Ambusher was dealt (CR 120.3), as Body of Knowledge does.
func init() {
	Register(Spec{
		OracleID:        "ad3add32-cb87-4cd4-8c36-a737a897a8c8",
		Name:            "Illusory Ambusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Illusory Ambusher — draw that many cards", cmdr2B08DrawThatMany),
		},
	})
}
