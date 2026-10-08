package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exile_until_another.go — "exile the top card of your library. You
// may play it until you exile another card with this <permanent>"
// (#2539, ADR 0066 amendment 2026-10-08): Unstable Amulet, Furious
// Rise, Superior Foes of Spider-Man.
//
// The body is the engine's ExileTopUntilAnotherForEffect, the one
// place such a window opens and the one place it closes. It reads the
// resolving item: its controller is "you", and its SourceObject is
// "this <permanent>" as it was when the ability went on the stack
// (CR 607.2a, CR 400.7). So the card files only say WHEN the exile
// happens.

// exileTopUntilYouExileAnother is the shared resolution: exile the top
// card of the controller's library and open its window, closing the
// one this source object opened for them before.
func exileTopUntilYouExileAnother(g *game.Game, item *game.StackItem) error {
	return g.ExileTopUntilAnotherForEffect(item, 1)
}
