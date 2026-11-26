package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devour in Flames — Sorcery {2}{R} (Oath of the Gatewatch, #5545):
//
//	"As an additional cost to cast this spell, return a land you control to its
//	 owner's hand.
//	 Devour in Flames deals 5 damage to target creature or planeswalker."
//
// Deprive's additional cost (a land returned as the spell is cast; with no
// land the spell can't be cast), then Bathe in Dragonfire's damage over a
// creature or planeswalker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ea1fc580-a640-4f49-8f29-405c2830eca8",
		Name:           "Devour in Flames",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.ReturnPermanentToHandCost("a land you control", effects.Land()),
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 5)),
		},
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
