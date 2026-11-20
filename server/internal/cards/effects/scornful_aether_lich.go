package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scornful Aether-Lich — Artifact Creature — Zombie Wizard {3}{U}, 2/4
// (Conflux, #5587):
//
//	"{W}{B}: This creature gains fear and vigilance until end of turn.
//	 (Attacking doesn't cause it to tap, and it can't be blocked except
//	 by artifact creatures and/or black creatures.)"
//
// thisCreatureUntilEOT with two keywords and no pump, ending in the
// cleanup step (CR 514.2). With no {T} it can be activated any number
// of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8834831-5571-466d-9227-bb91ca2dce3f",
		Name:         "Scornful Aether-Lich",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}{B}: This creature gains fear and vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"fear", "vigilance"}}},
			Cost:    ManaCost("{W}{B}"),
			Effect:  thisCreatureUntilEOT("Scornful Aether-Lich — fear and vigilance until end of turn", 0, 0, "fear", "vigilance"),
		}},
	})
}
