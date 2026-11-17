package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lat-Nam Adept — Creature — Human Wizard {3}{U}, 3/3 (The Brothers' War,
// #5482):
//
//	"Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature."
//
// Atlantean Cavalry's trigger without the vigilance: b41DrewYourSecondCardThisTurn
// reads the turn's draw tally after the draw, so it fires on the second
// card and no other, on any player's turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d5a1870-6d26-47e7-a86c-f3ff15cb0707",
		Name:         "Lat-Nam Adept",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{s82BroB07SecondDrawCounter("Lat-Nam Adept")},
	})
}
