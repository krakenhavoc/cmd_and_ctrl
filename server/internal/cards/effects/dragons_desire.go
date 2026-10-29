package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon's Desire — Sorcery {2}{R}{R} (The Hobbit):
//
//	"Add {R} for each artifact your opponents control."
//
// Mana Geyser's shape over a different count. The artifacts are counted on
// resolution, across every opponent's side of the battlefield (tokens
// included); none adds nothing, since AddMana ignores an empty count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "caef715c-3ca6-4f56-9dd5-aebce9cd0621",
		Name:         "Dragon's Desire",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddMana{
				Player: item.Controller,
				Amount: ManaN(hobB7ArtifactsByOpponents(ctx.Game, item.Controller), "R"),
			}.Apply(ctx)
		},
	})
}
