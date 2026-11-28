package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bullwhip — Artifact {4} (Stronghold, S82 OLD batch 747, #5627):
//
//	"{2}, {T}: This artifact deals 1 damage to target creature. That creature
//	 attacks this turn if able."
//
// The damage is dealt by the artifact, then the creature (if it is still a
// legal target, CR 608.2b) gets Imps' Taunt's attack requirement until end of
// turn (AddAttackRequirementMod, checked as attackers are declared, CR 508.1d).
// It binds only the creature's controller, and only if that creature can attack
// this turn.
//
// No simplification.
func init() {
	const label = "Bullwhip"
	attacks := effects.S82OldB363TargetAttacksThisTurnIfAble(label)
	effects.Register(effects.Spec{
		OracleID:     "d4e7a9f5-5228-4e41-8f58-6e7453fe68dc",
		Name:         label,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: This artifact deals 1 damage to target creature. That creature attacks this turn if able.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return attacks(g, item)
			},
		}},
	})
}
