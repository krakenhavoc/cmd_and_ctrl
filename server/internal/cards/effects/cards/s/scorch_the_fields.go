package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorch the Fields — Sorcery {4}{R}:
//
//	"Destroy target land. Scorch the Fields deals 1 damage to each Human
//	 creature."
//
// Every Human creature on the table takes the damage, yours included (one
// simultaneous damage event, CR 120.2); the land is destroyed first.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f9bce77-f7f0-453b-86cd-1e6bd71980a3",
		Name:         "Scorch the Fields",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		Targets:      effects.Harmful(effects.TargetPermanent("target land", effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.HasSubtype("Human"))) {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
