package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intrepid Tenderfoot — Creature — Insect Citizen {1}{G}, 2/2 (Edge of
// Eternities, #4461):
//
//	"{3}: Put a +1/+1 counter on this creature. Activate only as a
//	 sorcery."
//
// Hangarback Walker's counter on its own source, with no tap and the
// sorcery timing gate (CR 602.5d): your main phase, with the stack
// empty. It can be activated any number of times a turn. A sorcery-speed
// row answers nothing on the stack, so it declares only the counter it
// puts on its own source (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94023b23-339a-426e-85cc-cc361ef68754",
		Name:         "Intrepid Tenderfoot",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
			Purpose:      game.Purpose{SourceCounters: 1},
			Cost:         ManaCost("{3}"),
			SorcerySpeed: true,
			Effect:       plusOneCountersOnThis(1),
		}},
	})
}
