package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// EPF Point Squad — {1}{R/W}{R/W} Creature — Human Soldier 2/1 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Alliance — Whenever another creature you control enters, put a +1/+1
//	 counter on this creature."
//
// Alliance is an ability word (CR 207.2c): the trigger is
// WheneverAnotherCreatureEntersUnderYourControl, as on Slash, Reptile
// Rampager, and the counter goes on this creature while it is still on
// the battlefield (CounterOnThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2dea1935-c227-46d5-9d98-e7d2cd612f98",
		Name:         "EPF Point Squad",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("EPF Point Squad — alliance: a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
