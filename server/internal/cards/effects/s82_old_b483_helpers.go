package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b483_helpers.go — shared pieces of Rise of the Eldrazi batch 483
// (S82, #5581): Mul Daya Channelers' reads of the top card of the library.

// OldB483TopCardOfYourLibrary is the card on top of `you`'s library, if any.
func OldB483TopCardOfYourLibrary(g *game.Game, you uuid.UUID) (game.Card, bool) {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Library == nil {
		return game.Card{}, false
	}
	top, err := p.Library.Top()
	if err != nil {
		return game.Card{}, false
	}
	return top, true
}

// OldB483TopCardIsACreature is "the top card of your library is a creature
// card" for the player `you`.
func OldB483TopCardIsACreature(g *game.Game, you uuid.UUID) bool {
	top, ok := OldB483TopCardOfYourLibrary(g, you)
	return ok && top.IsCreature()
}

// OldB483TopCardIsALand is "the top card of your library is a land card" for
// the player `you`.
func OldB483TopCardIsALand(g *game.Game, you uuid.UUID) bool {
	top, ok := OldB483TopCardOfYourLibrary(g, you)
	return ok && top.IsLand()
}
