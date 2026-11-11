package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Printlifter Ooze — Creature — Ooze {1}{G}, 2/2 (Murders at Karlov Manor
// Commander, #5368):
//
//	"Deathtouch
//	 Whenever this creature or another creature you control is turned face
//	 up, create a 0/0 green Ooze creature token with trample. The token
//	 enters with X +1/+1 counters on it, where X is the number of other
//	 creatures you control.
//	 Disguise {3}{G}"
//
// X is read as the ability resolves, before the token exists, so every
// creature counted is another one; the token enters with the counters
// (CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3cbed112-9f2b-46b4-a9c3-bfd8a6856f79",
		Name:             "Printlifter Ooze",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				On(game.EventTurnedFaceUp, ACreatureYouControlWasTurnedFaceUp,
					"Printlifter Ooze — create a 0/0 green Ooze token with trample and X +1/+1 counters",
					cmdr2S38OozeFromFaceUp),
				game.Purpose{Tokens: 1}),
		},
	})
}
