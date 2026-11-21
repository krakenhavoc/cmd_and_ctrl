package effectstest

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// odyB1Flashback casts the card out of the active seat's graveyard for
// its flashback price, resolves it, and checks it was exiled.
func OdyB1Flashback(t *testing.T, g *game.Game, id uuid.UUID, cost string, targets ...game.TargetRef) {
	t.Helper()
	OdyB1CastFlashback(t, g, id, cost, targets...)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}

// odyB1CastFlashback casts the card out of the active seat's graveyard
// for its flashback price, paid from a pool holding exactly that price,
// and passes priority around the table once.
func OdyB1CastFlashback(t *testing.T, g *game.Game, id uuid.UUID, cost string, targets ...game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	DcMana(t, g, me.ID, OdyB1Pool(cost))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true, Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left after paying %s", n, cost)
	}
	PassPriorityAroundTable(t, g)
}

// odyB1Pool is the mana that pays exactly `cost`, its generic part as
// {C}: "{4}{R}" is "{C}{C}{C}{C}{R}".
func OdyB1Pool(cost string) string {
	var out strings.Builder
	for _, sym := range strings.Split(strings.Trim(cost, "{}"), "}{") {
		if n, err := strconv.Atoi(sym); err == nil {
			out.WriteString(strings.Repeat("{C}", n))
			continue
		}
		out.WriteString("{" + sym + "}")
	}
	return out.String()
}
