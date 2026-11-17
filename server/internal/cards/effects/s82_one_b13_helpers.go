package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82One13TuckTargetToLibraryBottom puts each still-legal card target on the
// bottom of its owner's library: "put target card from your graveyard on the
// bottom of your library" (Transplant Theorist).
func s82One13TuckTargetToLibraryBottom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.TuckToLibraryForEffect(t.ID, true); err != nil {
			return err
		}
	}
	return nil
}
