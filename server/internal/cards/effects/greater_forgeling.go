package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Forgeling — Creature — Elemental {3}{R}{R}, 3/4 (Ravnica: City of
// Guilds, #5600):
//
//	"{1}{R}: This creature gets +3/-3 until end of turn."
//
// Hermitic Nautilus's +3/-3 row: no tap in the cost, so it can be
// activated any number of times, and each boost ends at cleanup (CR
// 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0d51f49-383b-40f1-9192-5925db1cdd1b",
		Name:         "Greater Forgeling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +3/-3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: -3}},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(3, -3, "Greater Forgeling — +3/-3 until end of turn"),
		}},
	})
}
