package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toadstool Admirer — Creature — Ouphe {G}, 1/1 (Wilds of Eldraine,
// #4471):
//
//	"Ward {2} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {2}.)
//	 {3}{G}: Put a +1/+1 counter on this creature."
//
// Ward is the shared trigger. The counter row is Fire Sages's at {3}{G}:
// any time you have priority, as often as you can pay, and nothing if the
// Admirer has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77b48b3c-c780-4c5b-90e6-33c5fd940c16",
		Name:         "Toadstool Admirer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Toadstool Admirer — ward {2}"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{G}"),
			Effect:  plusOneCountersOnThis(1),
		}},
	})
}
