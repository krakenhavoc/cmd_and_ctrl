package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frontline Devastator — Creature — Zombie Minotaur Warrior {3}{R}, 3/3
// (Hour of Devastation, #5534):
//
//	"Afflict 2 (Whenever this creature becomes blocked, defending player
//	 loses 2 life.)
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// Afflict is the shared Afflict row (CR 702.130a): when this creature
// becomes blocked, the defending player loses the life, which is life
// loss and not damage.
// The pump is the shared thisGetsUntilEndOfTurn body (Fathom Fleet
// Firebrand's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ace15b8-f853-4c44-b9c7-d307367b612c",
		Name:         "Frontline Devastator",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Afflict(2, "Frontline Devastator")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Frontline Devastator — +1/+0 until end of turn"),
		}},
	})
}
