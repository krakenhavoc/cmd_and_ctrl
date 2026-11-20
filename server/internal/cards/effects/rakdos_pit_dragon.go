package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakdos Pit Dragon — Creature — Dragon {2}{R}{R}, 3/3 (Dissension, #5598):
//
//	"{R}{R}: This creature gains flying until end of turn.
//	 {R}: This creature gets +1/+0 until end of turn.
//	 Hellbent — This creature has double strike as long as you have no cards
//	 in hand."
//
// Cobalt Golem's flying row, Dragon Hatchling's firebreathing
// (thisGetsRow) and a layer 6 hellbent grant (disSelfKeywordWhileHellbent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63dd611e-1c36-4e5d-ba45-87c7f629cd54",
		Name:         "Rakdos Pit Dragon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}{R}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{R}{R}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Rakdos Pit Dragon — flying"),
			},
			thisGetsRow("{R}", 1, 0, "Rakdos Pit Dragon"),
		},
		Static: []game.StaticAbility{disSelfKeywordWhileHellbent("double strike")},
	})
}
