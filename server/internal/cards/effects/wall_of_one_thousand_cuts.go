package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of One Thousand Cuts — Creature — Wall {3}{W}{W}, 3/5 (Modern Horizons, #5519):
//
//	"Defender, flying
//	 {W}: This creature can attack this turn as though it didn't have
//	 defender."
//
// Hightide Hermit's defender waiver, paid with {W}: an until-end-of-turn
// waiver on the Wall; the keyword stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b97341a7-fc96-4281-b875-fb81065e0157",
		Name:            "Wall of One Thousand Cuts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{W}"),
			Effect:  CanAttackThisTurnAsThoughNoDefender("Wall of One Thousand Cuts — can attack this turn as though it didn't have defender"),
		}},
	})
}
