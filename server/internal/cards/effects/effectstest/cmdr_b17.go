package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdrB17ToGraveyard moves a card from a hand to its owner's graveyard
// through the engine, so the layer cache hears about it.
func CmdrB17ToGraveyard(t *testing.T, g *game.Game, p *game.Player, c game.Card) {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, c.InstanceID); err != nil {
		t.Fatalf("to graveyard %s: %v", c.Name, err)
	}
}
