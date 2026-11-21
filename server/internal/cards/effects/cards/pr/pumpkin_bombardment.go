package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pumpkin Bombardment — Sorcery {B/R}:
//
//	"As an additional cost to cast this spell, discard a card or pay
//	 {2}. Pumpkin Bombardment deals 3 damage to target creature."
//
// The either/or cost (ADR 0100 §2); the {2} branch joins the total at
// CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37827e84-9f5f-49ed-b939-0cf80dac82e7",
		Name:         "Pumpkin Bombardment",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.DiscardCost(1).Keyed("discard"),
			effects.ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Targets:   effects.TargetCreature("target creature"),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: effects.DamageToFirstTargetStep(3),
	})
}
