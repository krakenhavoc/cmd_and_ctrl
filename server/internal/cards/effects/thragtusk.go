package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thragtusk — Creature — Beast {4}{G}, 5/3:
//
//	"When this creature enters, you gain 5 life.
//	 When this creature leaves the battlefield, create a 3/3 green Beast
//	 creature token."
//
// Two printed triggers, two registered rows. The leaves trigger watches
// every exit, not only dying (bounce, exile, sacrifice all count), and
// its controller is the player who controlled Thragtusk as it left,
// read from last-known information (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dd0e91a-d16b-4718-8d11-1a3fcf8e0753",
		Name:         "Thragtusk",
		Purpose:      game.Purpose{LifeGain: 5},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thragtusk — you gain 5 life", Do(GainLife{Amount: 5})),
			WhenThisLeaves("Thragtusk — create a 3/3 green Beast creature token",
				Do(CreateToken{Template: TokenCard("3/3 green Beast"), N: 1})),
		},
	})
}
