package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grotag Siege-Runner — Creature — Goblin Rogue {1}{R}, 2/1 (Rise of the
// Eldrazi, S82 OLD b485):
//
//	"{R}, Sacrifice this creature: Destroy target creature with defender.
//	 This creature deals 2 damage to that creature's controller."
//
// The sacrifice is a cost (CR 118.3). The target's controller is read before
// the destroy, while the creature is still on the battlefield (CR 608.2h), so
// an indestructible or regenerated creature's controller is hit too; the damage
// is dealt by the Goblin as it last existed (Crater Elemental's source). A
// target that is no longer a creature with defender fails the resolution
// recheck (CR 608.2b) and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f76969e7-b075-417c-bc15-1b94a17fef54",
		Name:         "Grotag Siege-Runner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target creature with defender. This creature deals 2 damage to that creature's controller.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature with defender", effects.HasKeyword("defender")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					victim, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
