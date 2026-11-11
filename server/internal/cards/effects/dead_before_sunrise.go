package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "949a02ad-7530-4b2b-a1e3-0c70324e23d6",
		Name:         "Dead Before Sunrise",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: deadBeforeSunriseGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to target creature.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect:  sourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: "{T}: This creature deals damage equal to its power to target creature.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantAbilitiesFor{
				Match:    And(Creature(), YouControl(), b40Outlaw()),
				Keys:     []string{deadBeforeSunriseGrant},
				Also:     []game.Mod{game.ModifyPTMod(1, 0)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Dead Before Sunrise — outlaw creatures get +1/+0 and gain \"{T}: damage equal to power to target creature\"",
			}.Apply(ctx)
		},
	})
}
