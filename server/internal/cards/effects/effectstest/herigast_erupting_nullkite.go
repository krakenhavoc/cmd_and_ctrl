package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// handOffers is the wire's offer list for a hand card, by key.
func HandOffers(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID) map[string]protocol.AlternativeCostView {
	t.Helper()
	out := map[string]protocol.AlternativeCostView{}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID != card.String() {
				continue
			}
			for _, o := range s.Hand.Cards[i].AlternativeCosts {
				out[o.Key] = o
			}
		}
	}
	return out
}
