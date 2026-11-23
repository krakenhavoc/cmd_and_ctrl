package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Betrayer's Bargain — Instant {1}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 enchantment or pay {2}.
//	 Betrayer's Bargain deals 5 damage to target creature. If that
//	 creature would die this turn, exile it instead."
//
// The either/or cost (ADR 0100 §2), branches in printed order. The
// replacement is the spell's (ADR 0108 §1): the target is marked whether
// or not the damage is dealt.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfb3d862-d92d-4a4f-9a22-671b55d954fd",
		Name:         "Betrayer's Bargain",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Targets:   effects.TargetCreature("target creature"),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(5)),
	})
}
