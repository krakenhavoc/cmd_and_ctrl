package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pests of Honor — Creature — Mouse {2}{W}, 2/2 (Wilds of Eldraine, #4471):
//
//	"Celebration — At the beginning of combat on your turn, if two or more
//	 nonland permanents entered the battlefield under your control this
//	 turn, put a +1/+1 counter on this creature."
//
// Belligerent of the Ball's trigger shape (AtBeginningOfYourCombatIfCelebration):
// an intervening "if" checked as the trigger fires and again as it
// resolves (CR 603.4), over the celebration tally (CR 207.2c), with the
// creature itself counting on the turn it entered. The counter goes on it only
// while it is still the permanent the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2a4066a-f911-4622-beb2-683e7e10822f",
		Name:         "Pests of Honor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombatIfCelebration("Pests of Honor — put a +1/+1 counter on this creature", putCounterOnSelf),
		},
	})
}
