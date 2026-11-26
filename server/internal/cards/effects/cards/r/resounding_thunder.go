package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resounding Thunder — Instant {2}{R} (Shards of Alara, S82 OLD batch 509,
// #5588):
//
//	"Resounding Thunder deals 3 damage to any target.
//	 Cycling {5}{B}{R}{G}
//	 When you cycle this card, it deals 6 damage to any target."
//
// The cycle trigger fires from the graveyard (CR 702.29c), like Deem Worthy's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0e1d26e-2fa1-4826-bfe9-01b0bd2c138b",
		Name:         "Resounding Thunder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{5}{B}{R}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Resounding Thunder — cycled: 6 damage to any target",
					effects.DamageToFirstTarget(6))),
				effects.TargetAny()),
		},
	})
}
