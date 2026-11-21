package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b97341a7-fc96-4281-b875-fb81065e0157",
		Name:            "Wall of One Thousand Cuts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.CanAttackThisTurnAsThoughNoDefender("Wall of One Thousand Cuts — can attack this turn as though it didn't have defender"),
		}},
	})
}
