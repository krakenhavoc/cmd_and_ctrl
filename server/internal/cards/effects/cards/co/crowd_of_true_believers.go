package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crowd of True Believers — {W} Creature — Human Citizen 1/2 (Marvel
// Super Heroes, #4455):
//
//	"{T}: Target creature you control that's attacking alone gets +1/+0
//	 until end of turn. You gain 1 life."
//
// "Attacking alone" is CR 506.5 read live (AttackingAlone: attacking and
// the only attacker). The target is checked again on resolution
// (CR 608.2b); the life is gained either way unless the whole ability
// fizzles (CR 608.2b: with its only target illegal it does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06296d4d-a1b3-4e9a-a2fd-a0183bb892a1",
		Name:         "Crowd of True Believers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature you control that's attacking alone gets +1/+0 until end of turn. You gain 1 life.",
			Cost:    game.AbilityCost{Tap: true},
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)), game.AnswerPump),
			Targets: effects.TargetCreature("target creature you control that's attacking alone", effects.YouControl(), effects.AttackingAlone()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Crowd of True Believers — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
