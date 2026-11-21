package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// finB3Has reports whether the card, read through the layers, has the
// ability.
func FinB3Has(c game.Card, ability string) bool {
	for _, a := range c.Effective().Abilities {
		if a == ability {
			return true
		}
	}
	return false
}

// --- Resentful Revelation ------------------------------------------------------

func FinB3LibraryCards(me *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	return ids
}
