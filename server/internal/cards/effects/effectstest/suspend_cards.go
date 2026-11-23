package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stampLayout marks a hand fixture as Scryfall-imported, which is
// what game.HasNoManaCost requires before it will read an empty mana
// cost as CR 118.6's unpayable one.
func StampLayout(g *game.Game, p *game.Player, id uuid.UUID, layout string) {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].Layout = layout
			return
		}
	}
	_ = g
}

// suspendOf reads the card's declared suspend action, or nil.
func SuspendOf(oracle string) *game.SpecialAction {
	for _, sa := range game.SpecialActionsFor(oracle) {
		if sa.Kind == game.SpecialActionSuspend {
			out := sa
			return &out
		}
	}
	return nil
}
