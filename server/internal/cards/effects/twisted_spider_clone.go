package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Spider-Clone — Creature — Spider Human {3}{G}, 4/4 (Marvel's
// Spider-Man, #4460):
//
//	"When this creature enters, put a +1/+1 counter on each creature you
//	 control with a +1/+1 counter on it."
//
// The set is read before the first counter goes on (CR 608.2c), so a
// creature that had none gets none, and each creature gets one placement.
func init() {
	Register(Spec{
		OracleID:     "58dee718-6820-495a-8c03-5cc130b84167",
		Name:         "Twisted Spider-Clone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twisted Spider-Clone — +1/+1 counter on each creature you control with a +1/+1 counter",
				putPlusOneCountersOnEachYouControlThat(1, Not(b13WithoutPlusOneCounter()))),
		},
	})
}
