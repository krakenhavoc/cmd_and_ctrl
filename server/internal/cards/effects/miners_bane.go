package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miner's Bane — Creature — Elemental {4}{R}{R}, 6/3 (Magic 2015, #5556):
//
//	"{2}{R}: This creature gets +1/+0 and gains trample until end of
//	 turn."
//
// The eerie creatures' self pump (thisCreatureUntilEOT) behind a mana
// cost with no {T}: each activation stacks until the cleanup step, and a
// Bane that left the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ca75f61-60bf-487f-8cb3-b955cdd81e2a",
		Name:         "Miner's Bane",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +1/+0 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{2}{R}"),
			Effect:  thisCreatureUntilEOT("Miner's Bane — +1/+0 and trample until end of turn", 1, 0, "trample"),
		}},
	})
}
