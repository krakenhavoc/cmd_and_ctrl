package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Ambush — Sorcery {1}{R} (Portal Three Kingdoms, #5622):
//
//	"Fire Ambush deals 3 damage to any target."
//
// Volcanic Hammer's body (damageToFirstTarget). A target that became
// illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50463946-1ce3-4ff0-ad68-2fb87adbe2fd",
		Name:         "Fire Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
