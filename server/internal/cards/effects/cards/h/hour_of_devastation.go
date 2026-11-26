package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hour of Devastation — {3}{R}{R} Sorcery (Hour of Devastation):
//
//	"All creatures lose indestructible until end of turn. Hour of Devastation
//	 deals 5 damage to each creature and each non-Bolas planeswalker."
//
// The creatures that lose indestructible are the ones on the battlefield as it
// resolves (CR 611.2c); the damage then goes to every creature and to every
// planeswalker that does not have the Bolas subtype.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9ea0d4c-cd71-4817-8213-96899e1a14bb",
		Name:         "Hour of Devastation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
				if err := effects.G2LosesIndestructibleUntilEOT(ctx, c.InstanceID); err != nil {
					return err
				}
			}
			return effects.DamageEachMatching(ctx, effects.Or(effects.Creature(),
				effects.And(effects.Planeswalker(), effects.Not(effects.HasSubtype("Bolas")))), 5)
		},
	})
}
