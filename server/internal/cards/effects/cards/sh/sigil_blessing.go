package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigil Blessing — Instant {G}{W} (Shards of Alara, S82 OLD batch 511, #5588):
//
//	"Until end of turn, target creature you control gets +3/+3 and other
//	 creatures you control get +1/+1."
//
// The other creatures are the ones you control as it resolves (CR 611.2c);
// a target that left is not boosted and the rest still get +1/+1 only if
// the spell resolves at all (an illegal sole target counters it, CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87b5d197-4400-47ff-89a7-9aeeb7c9a0b4",
		Name:         "Sigil Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.BoostUntilEOT{Target: target, Power: 3, Toughness: 3, Label: "Sigil Blessing — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(target)), Power: 1, Toughness: 1,
				Label: "Sigil Blessing — other creatures you control get +1/+1",
			}.Apply(ctx)
		},
	})
}
