package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b705_helpers.go — the Urza's Destiny "Scent of" instants (S82 #5621):
// "Reveal any number of <colour> cards in your hand. Target creature gets
// +X/+X or -X/-X until end of turn, where X is the number of cards revealed
// this way."

// OldB705ScentOfX builds one Scent's OnResolve. `color` is the colour letter of
// the cards that may be revealed, `sign` is +1 or -1, and `pickKey` is the card's
// own "choose-cards/<card>" key (an on-disk identity, never renamed or reused).
//
// It is the shared reveal (OldB706RevealAnyNumber) followed by the pump or
// shrink for that many (OldB706PumpByRevealed). The reveal is made as the spell
// resolves (CR 608.2d); a target that has left or become illegal is judged
// before this runs (CR 608.2b), so a fizzled Scent reveals nothing. With no
// card of the colour in hand nothing is asked and nothing is revealed (X is 0).
func OldB705ScentOfX(name, color string, sign int, pickKey string) func(item *game.StackItem, ctx *Context) error {
	return OldB706RevealAnyNumber(name, color, pickKey, OldB706PumpByRevealed(name, sign))
}
