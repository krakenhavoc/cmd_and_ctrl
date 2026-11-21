package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_creatures_test.go — S58 PR 5: the Bringers, Damia, Conflux,
// Sheoldred, Ulamog and the mana doublers. See jodi_big_spells_test.go
// for the oracle ID constants.

// putInHand pushes a card into `p`'s hand and returns its ID.
func PutInHand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}
