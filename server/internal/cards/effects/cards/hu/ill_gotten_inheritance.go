package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ill-Gotten Inheritance — Enchantment {3}{B} (Ravnica Remastered):
//
//	"At the beginning of your upkeep, this enchantment deals 1 damage to
//	 each opponent and you gain 1 life.
//	 {5}{B}, Sacrifice this enchantment: It deals 4 damage to target
//	 opponent and you gain 4 life."
//
// No simplification.
func illGottenInheritanceDrain() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(0, 4))
	p.LifeGain = 4
	return p
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "84e8ca5e-730e-468b-943c-7c6358ade95c",
		Name:         "Ill-Gotten Inheritance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ill-Gotten Inheritance — 1 damage to each opponent and you gain 1 life",
				effects.Do(effects.DamageEachOpponent{N: 1}, effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}, Sacrifice this enchantment: It deals 4 damage to target opponent and you gain 4 life.",
			Purpose: illGottenInheritanceDrain(),
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{5}{B}"},
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DealDamageToTheTarget(4)(g, item); err != nil {
					return err
				}
				return effects.GainLife{Amount: 4}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
