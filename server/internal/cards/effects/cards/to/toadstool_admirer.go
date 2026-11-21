package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "77b48b3c-c780-4c5b-90e6-33c5fd940c16",
		Name:         "Toadstool Admirer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Toadstool Admirer — ward {2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect:  effects.PlusOneCountersOnThis(1),
		}},
	})
}
