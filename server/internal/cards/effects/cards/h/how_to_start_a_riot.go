package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// How to Start a Riot — Instant — Lesson {2}{R}:
//
//	"Target creature gains menace until end of turn. (It can't be
//	 blocked except by two or more creatures.)
//	 Creatures target player controls get +2/+0 until end of turn."
//
// Two target clauses, a creature and a player, read separately: when
// one is illegal as the spell resolves the other still happens
// (CR 608.2b). The +2/+0 locks its set as it resolves (CR 611.2c), so a
// creature that player gets afterwards is not pumped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "268d0056-0d3e-438a-ab17-625a6f5ec68a",
		Name:         "How to Start a Riot",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPlayer("target player"),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "menace")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"menace"},
					Label:    "How to Start a Riot — menace",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			p, ok := ctx.ClauseTarget(1)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.ControlledByPlayer(p.ID)),
				Power: 2,
				Label: "How to Start a Riot — +2/+0",
			}.Apply(ctx)
		},
	})
}
