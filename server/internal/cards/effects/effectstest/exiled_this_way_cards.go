package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// markCommanderCard makes the named card in one of `owner`'s zones a
// commander, so a move to a hand or a library offers the command zone
// (CR 903.9b) and one into a graveyard or exile is offered it afterwards
// (CR 903.9a). Tests that need the prompt rather than a whole commander
// deck set the flag directly, the way b36Commander does on the way
// onto the battlefield.
func MarkCommanderCard(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	marked := false
	g.WithWriteLock(func() {
		for _, z := range []*game.Zone{owner.Graveyard, owner.Hand, owner.Library, g.Battlefield, g.Exile} {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				if z.Cards[i].InstanceID == id {
					z.Cards[i].IsCommander = true
					marked = true
					return
				}
			}
		}
	})
	if !marked {
		t.Fatalf("no card %s in any of %s's zones to mark as a commander", id, owner.ID)
	}
}
