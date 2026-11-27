package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curtain of Light — Instant {1}{W} (Saviors of Kamigawa, #5601):
//
//	"Cast this spell only during combat after blockers are declared.
//	 Target unblocked attacking creature becomes blocked. (This spell works
//	 on creatures that can't be blocked.)
//	 Draw a card."
//
// The timing is a CastCondition on the steps from declare blockers on: nobody
// has priority in the declare blockers step until every blocker is declared
// (#1501), so every step after it is "after blockers are declared". The target
// is an attacking creature no creature is blocking; if it has been blocked, or
// has left combat, by resolution it is an illegal target and the spell does
// nothing, as the whole spell does when its only target is illegal (CR
// 608.2b). Otherwise it becomes blocked with no creature blocking it
// (Game.BecomeBlockedForEffect, CR 509.1h), whatever its evasion, and Curtain
// draws a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "d6efbbf9-1449-45ce-b187-30041eb36d93",
		Name:               "Curtain of Light",
		Completeness:       effects.CompletenessFull,
		Purpose:            game.Purpose{Draws: 1},
		CastCondition:      effects.DuringCombatAfterBlockers,
		CastConditionLabel: "Cast this spell only during combat after blockers are declared.",
		Targets:            effects.TargetCreature("target unblocked attacking creature", effects.OldB585UnblockedAttacker()),
		OnResolve:          effects.OldB585CurtainOfLightResolve,
	})
}
