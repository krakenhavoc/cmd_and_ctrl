package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// open_the_way_test.go — #2581's catalog card: "X can't be greater than
// the number of players in the game. Reveal cards from the top of your
// library until you reveal X land cards. Put those land cards onto the
// battlefield tapped and the rest on the bottom of your library in a
// random order."

const OpenTheWayOracle = "e07297c7-83bc-4162-bbb1-362bc737efe0"

func CastOpenTheWay(t *testing.T, g *game.Game, x int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Open the Way", TypeLine: "Sorcery", ManaCost: "{X}{G}{G}",
		OracleID: OpenTheWayOracle, Owner: me.ID, Controller: me.ID})
	AdvanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x})
}
