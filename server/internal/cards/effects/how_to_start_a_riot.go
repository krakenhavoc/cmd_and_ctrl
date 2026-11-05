package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "268d0056-0d3e-438a-ab17-625a6f5ec68a",
		Name:         "How to Start a Riot",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetPlayer("target player"),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (GrantKeywordUntilEOT{
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
			return BoostUntilEOT{
				Match: And(Creature(), controlledBy(p.ID)),
				Power: 2,
				Label: "How to Start a Riot — +2/+0",
			}.Apply(ctx)
		},
	})
}
