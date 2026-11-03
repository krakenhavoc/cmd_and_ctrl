package heuristic

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4694: restricted mana that FLOATS (Ashling, Rimebound's "add two mana
// of any one color. Spend this mana only to cast spells with mana value 4
// or greater", added by a trigger) is in the seat's pool, its restriction
// on the run (ManaPoolRun.Restrictions). The mana model counts it as it
// counts a restricted row: for the casts it admits, and only with
// PlanRestrictedMana.

// mv4Only is game.ManaRestrictManaValueAtLeast(4) as the wire carries
// it; a policy test may not import internal/game.
var mv4Only = []string{"mv:>=4"}

func restrictedPoolState(t *testing.T, bf []protocol.CardView) (*Policy, *state) {
	t.Helper()
	me := uuid.New()
	v := planSeatView(me, bf)
	v.Seats[0].ManaPool = []protocol.ManaPoolRun{{Color: "R", Count: 2, Restrictions: mv4Only}}
	p := New()
	return p, p.newState(aiseat.Input{Seat: me, View: v})
}

func TestManaModelCountsRestrictedPoolManaForTheCastsItAdmits(t *testing.T) {
	p, st := restrictedPoolState(t, []protocol.CardView{planLand("{R}"), planLand("{R}")})
	if off := st.manaAvailable(manaModel{}); len(off) != 2 {
		t.Fatalf("off: units = %#x, want the two Mountains alone, the restricted mana left out", off)
	}
	units := st.manaAvailable(p.manaModel(st))
	if len(units) != 4 {
		t.Fatalf("on: units = %#x, want two restricted {R} and two Mountains", units)
	}

	three := protocol.CardView{Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}"}
	four := protocol.CardView{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}"}
	fireball := protocol.CardView{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}"}
	xCost := func(x int) manaCost {
		c := parseManaCost(fireball.ManaCost, x)
		c.admits = st.castAdmits(&fireball, "", 0, x)
		return c
	}
	cases := []struct {
		name string
		cost manaCost
		ok   bool
	}{
		{"a three-drop can't spend it, and two Mountains don't pay {2}{R}", costFor(st, three.ManaCost, three), false},
		{"a four-drop spends it beside the Mountains", costFor(st, four.ManaCost, four), true},
		{"{X}{R} at X = 2 is mana value 3: the Mountains alone can't pay it", xCost(2), false},
		{"{X}{R} at X = 3 is mana value 4: everything pays it", xCost(3), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, ok := payMana(append([]manaUnit(nil), units...), c.cost, nil); ok != c.ok {
				t.Errorf("paid = %v, want %v", ok, c.ok)
			}
		})
	}
}
