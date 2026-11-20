package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impractical Joke — Sorcery {R}:
//
//	"Damage can't be prevented this turn. Impractical Joke deals 3
//	 damage to up to one target creature or planeswalker."
//
// The turn grant (ADR 0107 §5) begins first. "Up to one": with no
// target chosen the spell still resolves and the grant still happens.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8da73ed-2834-4094-8723-a4ad75660290",
		Name:         "Impractical Joke",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to one target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageCantBePreventedThen(effects.DamageToFirstTargetStep(3)),
	})
}
