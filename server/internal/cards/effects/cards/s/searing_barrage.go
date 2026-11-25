package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Searing Barrage — Instant {4}{R} (Throne of Eldraine, #5516):
//
//	"Searing Barrage deals 5 damage to target creature.
//	 Adamant — If at least three red mana was spent to cast this spell,
//	 Searing Barrage deals 3 damage to that creature's controller."
//
// Blur of Blades' controller read: the creature's controller is taken before
// the damage, so a creature the 5 damage kills still names who takes the 3
// (CR 608.2h). A target that left in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "26c17025-1123-46d9-a8bd-edbad01c4b6e",
		Name:           "Searing Barrage",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R"},
		Purpose:        effects.WithAnswers(effects.ForTargets(effects.DamageToTarget(0, 5)), game.AnswerRemove),
		Targets:        effects.TargetCreature("target creature"),
		OnResolve:      effects.OldB140SearingBarrageAdamantSpent,
	})
}
