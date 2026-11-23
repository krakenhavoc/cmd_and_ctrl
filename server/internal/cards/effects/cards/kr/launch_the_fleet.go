package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const launchTheFleetGrant = "launch-the-fleet/attack-soldier"

// Launch the Fleet — Sorcery {W} (Warhammer 40,000 Commander):
//
//	"Strive — This spell costs {1} more to cast for each target beyond the
//	 first.
//	 Until end of turn, any number of target creatures each gain "Whenever
//	 this creature attacks, create a 1/1 white Soldier creature token that's
//	 tapped and attacking.""
//
// Strive is the per-target surcharge (CR 207.2c). Zero targets is legal; each
// target still legal as the spell resolves gets the ability, and one that left
// does not (CR 608.2b). The token attacks the player the creature attacked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3362dec2-b157-4a00-8abf-0d7b5a14c6f8",
		Name:         "Launch the Fleet",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{1}", "Strive — This spell costs {1} more to cast for each target beyond the first."),
		},
		Grants: []effects.AbilityGrant{{
			Key: launchTheFleetGrant,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisAttacksTokensTappedAndAttacking("Launch the Fleet — create a 1/1 white Soldier tapped and attacking", "1/1 white Soldier", 1),
			},
			Text: "Whenever this creature attacks, create a 1/1 white Soldier creature token that's tapped and attacking.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || !effects.OnBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (effects.GrantAbilitiesFor{Target: t.ID, Keys: []string{launchTheFleetGrant},
					Label: "Launch the Fleet — \"whenever this creature attacks, create a 1/1 white Soldier tapped and attacking\""}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
