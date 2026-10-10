package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unliving Psychopath — Creature — Zombie Assassin {2}{B}{B}, 0/4:
//
//	"{B}: This creature gets +1/-1 until end of turn.
//	 {B}, {T}: Destroy target creature with power less than this
//	 creature's power."
//
// The destroy clause is relative to the Psychopath (#2146): its power
// is compared as the ability is announced and again as it resolves, so
// pumping it in response widens nothing already chosen and a
// Psychopath that has shrunk, or died to its own pump, makes the
// target illegal (its last-known power, CR 608.2h). At its printed 0
// power nothing is legal, so the ability cannot be activated until the
// first pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "038463c8-d2e1-4e7b-829d-67a744ae2660",
		Name:         "Unliving Psychopath",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{B}"),
				Effect:  thisGetsUntilEndOfTurn(1, -1, "Unliving Psychopath — +1/-1 until end of turn"),
			},
			{
				Label: "{B}, {T}: Destroy target creature with power less than this creature's power.",
				Cost:  Plus(ManaCost("{B}"), TapCost()),
				Targets: RelativeToSource(
					TargetCreature("target creature with power less than this creature's power"),
					LesserPower()),
				Effect: destroyFirstLegalCardTarget,
			},
		},
	})
}
