package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One with the Machine — Sorcery {3}{U} (EDHREC rank 2874):
//
//	"Draw cards equal to the greatest mana value among artifacts you
//	 control."
//
// The artifact deck's Overflowing Insight. The number is read as the
// spell resolves — the highest mana value among the caster's
// artifacts, post-layer types so an animated artifact counts — and
// drawn in one go. No artifacts, or only mana-value-zero ones, draws
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "757eadac-dd29-4a7f-8683-5bc823168326",
		Name:         "One with the Machine",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{
				Player: item.Controller,
				N:      b27GreatestManaValueAmongArtifactsControlled(ctx.Game, item.Controller),
			}.Apply(ctx)
		},
	})
}
