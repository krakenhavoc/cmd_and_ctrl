package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_a_test.go — ADR 0107 §4 (#1855): the disturb cards
// whose faces are keywords, Auras that pump or grant, and simple
// triggers. disturb_test.go has the helpers and the seam's own tests.

// backFaceOnBattlefield imports a row and puts it onto the battlefield
// back face up, as a disturbed permanent that has been there a while
// (no summoning sickness), so a back-face attack trigger can be tested
// without waiting a turn.
func BackFaceOnBattlefield(g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	c.SetFace(1)
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// openPickTarget reports whether a pick_target prompt is waiting.
func OpenPickTarget(g *game.Game) bool {
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}
