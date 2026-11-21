package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Owner decision (2026-09-17): "any color" mana offers ALL FIVE
// colours, with the commander's colour identity listed first — no
// narrowing. Only the cards whose printed text says "any color in your
// commander's color identity" still narrow.

// monoGreenFirst is what a five-colour pick looks like under a
// mono-green commander: G first, then the rest in WUBRG order.
var MonoGreenFirst = []string{"G", "W", "U", "B", "R"}

// riderGiveImportedCommander pushes a commander stamped the way
// deck.ToGameCard stamps a real card: a Scryfall printing behind it and
// Scryfall's own color_identity, EMPTY included. An empty identity on
// such a card is Kozilek's real answer, not a missing import.
func RiderGiveImportedCommander(g *game.Game, p *game.Player, name, manaCost string, identity []string) {
	g.WithWriteLock(func() {
		p.Command.PushTop(game.Card{
			InstanceID:    uuid.New(),
			ScryfallID:    uuid.NewString(),
			OracleID:      uuid.NewString(),
			Name:          name,
			TypeLine:      "Legendary Creature — Eldrazi",
			ManaCost:      manaCost,
			ColorIdentity: identity,
			Owner:         p.ID,
			Controller:    p.ID,
			IsCommander:   true,
		})
	})
}
