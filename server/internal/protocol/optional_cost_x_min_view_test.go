package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Thieving Skydiver, Verdeloth
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestOptionalCostViewCarriesTheKickersXFloor is the #3232 view half:
// a kicker whose cost is {X} ships its printed floor ("X can't be 0")
// as `x_min` on the optional-cost offer, so the client's X picker opens
// at 1 once the kicker is chosen. A kicker with no floor (Verdeloth) and
// the printed cost itself carry none.
func TestOptionalCostViewCarriesTheKickersXFloor(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	skydiver, verdeloth := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{InstanceID: skydiver, Name: "Thieving Skydiver", TypeLine: "Creature — Merfolk Rogue",
				OracleID: "f509ac02-f22e-4d1a-9fe1-afa96dd7806d", ManaCost: "{1}{U}"},
			{InstanceID: verdeloth, Name: "Verdeloth the Ancient", TypeLine: "Legendary Creature — Treefolk",
				OracleID: "8222321f-b74d-4f74-9ee7-79455433345a", ManaCost: "{4}{G}{G}"},
		} {
			c.Owner, c.Controller = me.ID, me.ID
			c.KnownBy = map[uuid.UUID]bool{me.ID: true}
			me.Hand.PushTop(c)
		}
	})

	v := ViewOfGameFor(g, me.ID.String())
	sky := handCardIn(t, v, me.ID.String(), skydiver.String())
	if len(sky.OptionalCosts) != 1 {
		t.Fatalf("Thieving Skydiver offers %d optional costs, want 1", len(sky.OptionalCosts))
	}
	if oc := sky.OptionalCosts[0]; oc.XMin != 1 || oc.ManaCost != "{X}" {
		t.Errorf("Skydiver's kicker = %+v, want mana {X} with x_min 1", oc)
	}
	ver := handCardIn(t, v, me.ID.String(), verdeloth.String())
	if len(ver.OptionalCosts) != 1 || ver.OptionalCosts[0].XMin != 0 || ver.OptionalCosts[0].ManaCost != "{X}" {
		t.Errorf("Verdeloth's kicker = %+v, want mana {X} and no floor", ver.OptionalCosts)
	}
}
