package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leaping Lizard — Creature — Lizard {1}{G}{G}, 2/3 (Homelands, #5635):
//
//	"{1}{G}: This creature gets -0/-1 and gains flying until end of turn."
//
// Both halves land on the Lizard as the ability resolves and end at
// cleanup (CR 514.2): the -0/-1 in layer 7c (CR 613.4c) and flying in
// layer 6 (CR 613.1f). Each activation takes another point of
// toughness, so a third one in a turn kills it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23d42beb-5293-4f12-a732-3531a0fa0ca0",
		Name:         "Leaping Lizard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: This creature gets -0/-1 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{G}"),
			Effect:  thisCreatureUntilEOT("Leaping Lizard — -0/-1 and flying until end of turn", 0, -1, "flying"),
		}},
	})
}
