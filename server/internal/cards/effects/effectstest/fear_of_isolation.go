package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// foiCastableHere is the cast gate the view's castability reads (CR
// 601.2h): whether the card's additional cost can be paid at all.
func FoiCastableHere(t *testing.T, g *game.Game, seat, card uuid.UUID) bool {
	t.Helper()
	ok, found := false, false
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for _, c := range p.Hand.Cards {
				if c.InstanceID == card {
					found = true
					ok = g.AnyAdditionalCostBranchPayableLocked(seat, c)
				}
			}
		}
	})
	if !found {
		t.Fatal("the card is not in a hand")
	}
	return ok
}

func FoiHandView(g *game.Game, seat, card uuid.UUID) *protocol.CardView {
	v := protocol.ViewOfGameFor(g, seat.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if c := &s.Hand.Cards[i]; c.InstanceID == card.String() {
				return c
			}
		}
	}
	return nil
}

func FoiCastMoves(g *game.Game, seat, card uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == card {
			out = append(out, m)
		}
	}
	return out
}
