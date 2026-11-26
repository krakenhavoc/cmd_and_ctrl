package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destructive Revelry — Instant {R}{G} (Theros, S82 OLD b392, #5561):
//
//	"Destroy target artifact or enchantment. Destructive Revelry deals 2
//	 damage to that permanent's controller."
//
// Root Out's target, then Kami's Flare's controller read (DestroyTargetThen
// DamageItsController): the controller is taken before the destruction, so
// the damage lands whether or not the permanent dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7165fade-ea96-4026-bc80-620729b8f70c",
		Name:         "Destructive Revelry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTargetThenDamageItsController(2),
	})
}
