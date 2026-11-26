package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Volley — Instant {3}{R} (Born of the Gods, S82 OLD b378, #5559):
//
//	"Until end of turn, creatures you control gain "{T}: This creature deals 1
//	 damage to any target.""
//
// Root Manipulation's grant: the set of creatures is fixed as the spell
// resolves (CR 611.2c), each gets the granted ability for the turn, and a
// creature that arrives later does not.
//
// No simplification.
func init() {
	const grant = "lightning-volley/ping"
	effects.Register(effects.Spec{
		OracleID:     "93b6083b-f49a-4641-b76f-8ccbec22f91d",
		Name:         "Lightning Volley",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to any target.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Keys:  []string{grant},
				Label: "Lightning Volley — creatures you control gain \"{T}: This creature deals 1 damage to any target.\"",
			}.Apply(ctx)
		},
	})
}
