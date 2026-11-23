package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_variants_test.go — ADR 0137's 2026-10-10 amendment (#2709):
// "one or more", rules over the chosen set, and graveyard-only
// materials, each pinned on the card that prints it, and each card's
// back face that reads what it was crafted with.

// pushGraveyardMaterial seeds a card into `p`'s graveyard with the
// fields a material test reads.
func PushGraveyardMaterial(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}
