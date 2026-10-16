package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hightide Hermit — Creature — Crab {4}{U}, 4/4:
//
//	"Defender
//	 When this creature enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 Pay {E}{E}: This creature can attack this turn as though it didn't
//	 have defender."
//
// The energy is ADR 0129 PR 1 (#1995). The activation is ADR 0106's
// 2026-10-16 amendment (#1936): an until-end-of-turn ModWaiveDefender
// that lets it attack every target while it keeps the keyword (CR 609.4),
// rather than losing defender, which is what a card counting creatures
// with defender would see.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3689ef94-2573-4424-b52e-379656340345",
		Name:            "Hightide Hermit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Purpose:         game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Hightide Hermit", 4),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay {E}{E}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    PayEnergy(2),
			Effect:  CanAttackThisTurnAsThoughNoDefender("Hightide Hermit — can attack this turn as though it didn't have defender"),
		}},
	})
}
