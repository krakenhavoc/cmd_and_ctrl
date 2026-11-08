package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krotiq Nestguard — Creature — Insect {2}{G}, 4/4 (Tarkir: Dragonstorm,
// #4463):
//
//	"Defender
//	 {2}{G}: This creature can attack this turn as though it didn't have
//	 defender."
//
// Hightide Hermit's activation for mana: an until-end-of-turn
// ModWaiveDefender on the Nestguard, so it can attack while it keeps the
// keyword (CR 702.3b is waived, not removed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33117d34-ea70-4566-92a3-5dec9bfe529a",
		Name:            "Krotiq Nestguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{G}"),
			Effect:  CanAttackThisTurnAsThoughNoDefender("Krotiq Nestguard — can attack this turn as though it didn't have defender"),
		}},
	})
}
