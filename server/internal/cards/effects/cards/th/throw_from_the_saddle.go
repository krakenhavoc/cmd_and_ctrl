package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throw from the Saddle — Sorcery {1}{G} (Outlaws of Thunder Junction,
// #4468):
//
//	"Target creature you control gets +1/+1 until end of turn. Put a
//	 +1/+1 counter on it instead if it's a Mount. Then it deals damage
//	 equal to its power to target creature you don't control."
//
// Wolf Strike's shape: the first target grows, the layers are recomputed,
// and Bite Down's one-sided bite (oneSidedBite) reads its power as it now
// stands. Whether it is a Mount is read as the spell resolves, through the
// card's own text (CR 612.2), and picks the counter over the pump, never
// both. A biter gone by resolution grows and deals nothing; a victim gone
// leaves the growth in place (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6780d29-0dd7-4f04-95b8-4ca6a55308f2",
		Name:         "Throw from the Saddle",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(biter.ID)
				if !ok {
					return nil
				}
				var err error
				if c.HasSubtype(ctx.TextSubtype("Mount")) {
					err = effects.AddCounter{Target: biter.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				} else {
					err = effects.BoostUntilEOT{Target: biter.ID, Power: 1, Toughness: 1, Label: "Throw from the Saddle — +1/+1"}.Apply(ctx)
				}
				if err != nil {
					return err
				}
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
			}
			return effects.OneSidedBite(item, ctx)
		},
	})
}
