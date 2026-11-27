package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB521Symbols counts the `color` mana symbols in a mana cost string
// (Eventide's chroma, CR 700.5's counting): one per coloured slot whose
// options include the colour, so a hybrid symbol counts toward each colour it
// offers. An empty or unparseable cost has none.
func S82OldB521Symbols(manaCost, color string) int {
	cost, err := game.ParseCost(manaCost)
	if err != nil {
		return 0
	}
	n := 0
	for _, req := range cost.Required {
		if SymbolIsAnyColor(req.Options, []string{color}) {
			n++
		}
	}
	return n
}

// S82OldB521Chroma is chroma over one card as a permanent: its own mana
// cost, and nothing for a face-down permanent (CR 708.2a).
func S82OldB521Chroma(c *game.Card, color string) int {
	if c == nil || c.FaceDownIsPermanent() {
		return 0
	}
	return S82OldB521Symbols(c.ManaCost, color)
}

// S82OldB521ChromaOfPermanents is "the number of <colour> mana symbols in the
// mana costs of permanents you control". It walks the battlefield directly so a
// layer pass can call it.
func S82OldB521ChromaOfPermanents(g *game.Game, controller uuid.UUID, color string) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller {
			n += S82OldB521Chroma(c, color)
		}
	}
	return n
}

// S82OldB521ChromaOfGraveyard is "the number of <colour> mana symbols in the
// mana costs of cards in your graveyard".
func S82OldB521ChromaOfGraveyard(g *game.Game, player uuid.UUID, color string) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		n += S82OldB521Symbols(c.ManaCost, color)
	}
	return n
}
