package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loporrit Scout — Creature — Rabbit Scout {2}{G}, 3/2 (Final Fantasy,
// #4462):
//
//	"Whenever another creature you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// A creature entering under your control other than the Scout triggers
// it once per creature; the pump is a layer 7c effect on the Scout that
// ends at cleanup, so two creatures entering in one turn make it +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24630190-a630-4dd3-a8f7-82f129c881e3",
		Name:         "Loporrit Scout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Loporrit Scout — gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Loporrit Scout — +1/+1 until end of turn")),
		},
	})
}
