package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secret Rendezvous — Sorcery {1}{W}{W} (EDHREC rank 1130):
//
//	"You and target opponent each draw three cards."
//
// White's group-hug draw three. The caster draws first, then the
// targeted opponent — the printed order — and a target that became
// illegal in response fizzles the whole spell, which the engine's
// CR 608.2b all-targets-illegal check does before OnResolve runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34b90d3e-f48c-41ff-b3e4-5cab7a9cd597",
		Name:         "Secret Rendezvous",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
				return err
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: 3}.Apply(ctx)
		},
	})
}
