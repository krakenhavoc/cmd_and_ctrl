package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Embermage — Creature — Human Wizard {3}{R}, 2/2 (Mirage, #5633):
//
//	"{1}{R}: This creature deals 1 damage to any target and 1 damage to itself."
//
// Both are one simultaneous damage event (CR 120.2); a target that is illegal
// at resolution makes the whole ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5771296a-16f8-423e-9b83-8e6f687cbb34",
		Name:         "Reckless Embermage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to any target and 1 damage to itself.",
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: effects.OldB793DealsOneToTargetAndTo(func(ctx *effects.Context) uuid.UUID {
				return ctx.Item.SourceCardID
			}),
		}},
	})
}
