package effectstest

import (
	"testing"

	"github.com/google/uuid"

	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Bilbo, Thief in the Night --------------------------------------

func BilboTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bilbo := PushDiesCreatureForTest(g, me.ID, cardsb.BilboThiefName, cardsb.BilboThiefOracleID,
		"Legendary Creature — Halfling Rogue", 2, 2)
	return g, me, opp, bilbo
}

// bilboOffer attacks with Bilbo and leaves the cast offer open.
func BilboOffer(t *testing.T, g *game.Game, me, opp *game.Player, bilbo uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, opp.ID, bilbo)
	PassPriorityAroundTable(t, g)
	MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, want...)
}
