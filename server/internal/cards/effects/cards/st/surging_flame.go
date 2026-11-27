package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surging Flame — Instant {1}{R} (Coldsnap, #5597):
//
//	"Ripple 4 (When you cast this spell, you may reveal the top four cards of
//	 your library. You may cast spells with the same name as this spell from
//	 among those cards without paying their mana costs. Put the rest on the
//	 bottom of your library.)
//	 Surging Flame deals 2 damage to any target."
//
// Bee Sting's damage with ripple (CR 702.60a) as a trigger of the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c51b991-f6fa-44d2-81a3-0894939f3c22",
		Name:         "Surging Flame",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		Triggered:    []game.TriggeredAbility{effects.S82OldB566Ripple("Surging Flame", 4)},
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
