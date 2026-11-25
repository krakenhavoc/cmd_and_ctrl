package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Swipe — Sorcery {G} (Modern Horizons):
//
//	"Target creature you control gets +2/+2 until end of turn if its power is
//	 2. Then it fights target creature you don't control."
//
// The power is read as Savage Swipe resolves, through the layers (CR 608.2c),
// so a creature that is not a 2-power one then gets nothing but still fights.
// Both targets are chosen as it is cast (CR 601.2c); the fight happens only
// while both are legal (CR 701.14b, FightTargetClauses). The pump comes before
// the fight, so the +2 power is part of the damage dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c74c87f7-3120-4fd7-8735-1d7a7a429643",
		Name:         "Savage Swipe",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.Not(effects.YouControl())),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.CurrentPower() != 2 {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Savage Swipe — +2/+2"}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
