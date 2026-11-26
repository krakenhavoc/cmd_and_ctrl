package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b375_helpers.go — shared by the Inspired creatures of Born of the Gods
// (S82 OLD batch 375, #5559).

// TokenCopyIsAlsoAnEnchantment is "except it's an enchantment in addition to its
// other types" on a token copy's template (CR 707.9b): Felhide Spiritbinder's
// copy. The type joins the copiable values, so a later copy of the token is an
// enchantment too.
func TokenCopyIsAlsoAnEnchantment(t *game.Card) {
	v := game.PrintedValues{TypeLine: t.TypeLine}
	v.AddCardType("Enchantment")
	t.TypeLine = v.TypeLine
}
