package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithless Looting — Sorcery for {R}:
//
//	"Draw two cards, then discard two cards."
//
// The archetypal loot. Printed order matters and the engine gets it
// right for free: DrawCards resolves synchronously, then the discard
// choice is queued against the post-draw hand, so the cards just
// drawn are legal discards exactly as in paper.
//
// Flashback {2}{R} is an alternative cast path from the graveyard
// (S29) and isn't modelled.
func init() {
	Register(Spec{
		OracleID:     "3d6fa57a-aa53-4b5c-b8af-a7612c823117",
		Name:         "Faithless Looting",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Flashback isn't implemented — the spell can only be cast from hand, never recast from your graveyard."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return lootOne(ctx.Game, item, 2)
		},
	})
}
