package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapExileTargetSpellLinked is "exile target spell" said by an enter
// trigger whose card is linked to this object (Card.ExiledWith, CR 400.7):
// the shared body of Spell Queller and Smirking Spelljacker. A spell that has
// left the stack by resolution links nothing.
func GapExileTargetSpellLinked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return g.ExileSpellLinkedKeyedForEffect(ref, t.ID, game.KeyedTail{})
		}
	}
	return nil
}
