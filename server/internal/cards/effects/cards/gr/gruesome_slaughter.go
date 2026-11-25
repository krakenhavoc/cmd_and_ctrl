package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruesome Slaughter — Sorcery {6} (Battle for Zendikar, #5547):
//
//	"Until end of turn, colorless creatures you control gain "{T}: This
//	 creature deals damage equal to its power to target creature.""
//
// The creatures are fixed as the spell resolves (CR 611.2c), so a colorless
// creature that arrives later in the turn does not get the ability.
//
// No simplification.
func init() {
	const grant = "gruesome-slaughter/tap-damage"
	effects.Register(effects.Spec{
		OracleID:     "5f6fbb1d-63c9-4913-8e99-9e8c3b4e8d8c",
		Name:         "Gruesome Slaughter",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to target creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: "{T}: This creature deals damage equal to its power to target creature.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Colorless()),
				Keys:     []string{grant},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Gruesome Slaughter — colorless creatures gain \"{T}: damage equal to power to target creature\"",
			}.Apply(ctx)
		},
	})
}
