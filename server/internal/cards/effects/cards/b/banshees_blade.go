package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banshee's Blade — Artifact — Equipment {2} (Mirrodin, #5606):
//
//	"Equipped creature gets +1/+1 for each charge counter on this
//	 Equipment.
//	 Whenever equipped creature deals combat damage, put a charge counter
//	 on this Equipment.
//	 Equip {2}"
//
// Excalibur II's counter-scaled pump (PumpAttachedPer) and Lost Jitte's
// trigger on combat damage to anything, a blocker included
// (WheneverEquippedCreatureDealsCombatDamage). The engine opens one
// damage event per recipient, and a trampler that deals damage to its
// blocker and the player at once has dealt combat damage once, so the
// trigger is one per batch: one counter per combat damage step, two for
// a creature with double strike, which deals damage in both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81a770e9-adea-4799-9fed-f10e97e7e901",
		Name:         "Banshee's Blade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters[game.CounterCharge] }),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEquippedCreatureDealsCombatDamage("Banshee's Blade — put a charge counter on this Equipment", effects.B33PutChargeCounterOnSelf),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
