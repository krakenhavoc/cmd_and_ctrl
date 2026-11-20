package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locthwain Gargoyle — Artifact Creature — Gargoyle {1}, 0/3 (Throne of Eldraine, #5516):
//
//	"{4}: This creature gets +2/+0 and gains flying until end of turn."
//
// The pump and the keyword are one until-end-of-turn effect on itself
// (thisCreatureUntilEOT), which does nothing if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8a04142-1d28-4c2e-b2d5-da88b2199678",
		Name:         "Locthwain Gargoyle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}: This creature gets +2/+0 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:    ManaCost("{4}"),
			Effect:  thisCreatureUntilEOT("Locthwain Gargoyle — +2/+0 and flying", 2, 0, "flying"),
		}},
	})
}
