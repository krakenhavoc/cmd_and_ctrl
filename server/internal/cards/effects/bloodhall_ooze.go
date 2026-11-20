package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodhall Ooze — Creature — Ooze {R}, 1/1 (Conflux, #5587):
//
//	"At the beginning of your upkeep, if you control a black permanent,
//	 you may put a +1/+1 counter on this creature.
//	 At the beginning of your upkeep, if you control a green permanent,
//	 you may put a +1/+1 counter on this creature."
//
// Two separate triggers, each with its own intervening if (CR 603.4):
// with a black and a green permanent both trigger, and each asks its
// "you may" (CR 603.5) as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79591c18-0065-49bb-a77a-734d4af87a3c",
		Name:         "Bloodhall Ooze",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(conB2AtYourUpkeepIfYouControlA("B", "black",
				"Bloodhall Ooze — put a +1/+1 counter on it (black)", putACounterOnThis(game.CounterPlusOne)),
				"Bloodhall Ooze — put a +1/+1 counter on it?"),
			Optional(conB2AtYourUpkeepIfYouControlA("G", "green",
				"Bloodhall Ooze — put a +1/+1 counter on it (green)", putACounterOnThis(game.CounterPlusOne)),
				"Bloodhall Ooze — put a +1/+1 counter on it?"),
		},
	})
}
