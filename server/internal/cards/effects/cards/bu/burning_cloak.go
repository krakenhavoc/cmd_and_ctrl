package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Cloak — Sorcery {R} (Portal, S82 OLD batch 782, #5630):
//
//	"Target creature gets +2/+0 until end of turn. Burning Cloak deals 2
//	 damage to that creature."
//
// The pump, then the damage as the next instruction (CR 608.2c): the creature
// is pumped even if the damage is prevented, and the damage is dealt by the
// spell to the creature it pumped. Nothing happens when the target is illegal
// on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4527c622-cebe-41eb-9178-4895cccffe99",
		Name:         "Burning Cloak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(burningCloakEntry()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(2, 0, "Burning Cloak — +2/+0 until end of turn")(item, ctx); err != nil {
				return err
			}
			return effects.DamageTheClauseTarget(0, 2)(ctx)
		},
	})
}

// burningCloakEntry is the clause as printed: the pick gets +2/+0 until end of
// turn (a benefit) and takes 2 damage from the spell.
func burningCloakEntry() game.TargetPurpose {
	e := effects.BenefitUntilEOT(0, 2, 0)
	e.Damage = 2
	return e
}
