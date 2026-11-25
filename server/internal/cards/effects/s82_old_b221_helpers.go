package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b221_helpers.go — shared bodies for Ixalan, batch 221 (#5532):
// the "as long as you control another <type>" self statics.

// OldB221ControlsAnother reports whether the source's controller controls a
// permanent other than the source that has one of `another`'s words, or
// any permanent (the source included) that has one of `any`'s. The words are
// read as words of the source's text (CR 612.2) and the permanents through
// the layers, so a changeling counts.
func OldB221ControlsAnother(g *game.Game, source *game.Card, another, anyWords []string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != source.Controller {
			continue
		}
		if c.InstanceID != source.InstanceID {
			for _, w := range another {
				if c.HasSubtypeWord(source.TextSubtype(w)) {
					return true
				}
			}
		}
		for _, w := range anyWords {
			if c.HasSubtypeWord(source.TextSubtype(w)) {
				return true
			}
		}
	}
	return false
}

// OldB221SelfWhileAnother is the AppliesTo of "as long as you control
// another <word> [or a <word>], this creature gets/has ...": the source
// alone, while the condition holds.
func OldB221SelfWhileAnother(another, anyWords []string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && OldB221ControlsAnother(g, source, another, anyWords)
	}
}
