package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const deadBeforeSunriseGrant = "dead-before-sunrise/tap-damage"

// Dead Before Sunrise — Instant {3}{R} (Outlaws of Thunder Junction
// Commander, #5368):
//
//	"Until end of turn, outlaw creatures you control get +1/+0 and gain
//	 '{T}: This creature deals damage equal to its power to target
//	 creature.' (Assassins, Mercenaries, Pirates, Rogues, and Warlocks are
//	 outlaws.)"
//
// One effect over the outlaw creatures its controller has as it resolves,
// fixed then (CR 611.2c): +1/+0 and the granted ability share one timestamp
// (GrantAbilitiesFor.Also, Fake Your Own Death's shape). The granted ability
// is Burning Anger's, aimed at a creature: it is the recipient's own, so it
// taps the recipient (a summoning-sick one can't pay the {T}) and "this
// creature" is the recipient, its power read at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "949a02ad-7530-4b2b-a1e3-0c70324e23d6",
		Name:         "Dead Before Sunrise",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: deadBeforeSunriseGrant,
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
				Match:    effects.And(effects.Creature(), effects.YouControl(), effects.B40Outlaw()),
				Keys:     []string{deadBeforeSunriseGrant},
				Also:     []game.Mod{game.ModifyPTMod(1, 0)},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Dead Before Sunrise — outlaw creatures get +1/+0 and gain \"{T}: damage equal to power to target creature\"",
			}.Apply(ctx)
		},
	})
}
