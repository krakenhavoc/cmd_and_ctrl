package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b668_helpers.go — S82 OLD batch 668 (Prophecy, activated abilities).

// OldB668TuckTheTargetOnTheBottom is "Put target <permanent> on the bottom of
// its owner's library" for the first still-legal card target (Mercenary
// Informer, Rebel Informer): it goes to the bottom of its owner's library, and
// does nothing if the target is no longer legal (CR 608.2b).
func OldB668TuckTheTargetOnTheBottom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return PutIntoLibrary{Card: t.ID, ToBottom: true}.Apply(ctx)
		}
	}
	return nil
}
