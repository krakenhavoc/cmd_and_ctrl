package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Wall — Creature — Wall {3}{U}, 3/4 (Judgment, #5611):
//
//	"Defender
//	 {W}: This creature can attack this turn as though it didn't have
//	 defender."
//
// Wall of One Thousand Cuts's defender waiver.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4ab446a-97e4-46a5-89a8-180a0bf3447c",
		Name:            "Mirror Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.CanAttackThisTurnAsThoughNoDefender("Mirror Wall — can attack this turn as though it didn't have defender"),
		}},
	})
}
