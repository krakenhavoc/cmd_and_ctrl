package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestLandOrderAvoidsAnEmptyBoardBounce(t *testing.T) {
	forest := planForest(cardID(10))
	bounce := land(cardID(11), 0)
	bounce.Name, bounce.TypeLine = "Chamber", "Land"
	bounce.Purpose = &protocol.PurposeView{LandEntry: "tapped", ReturnsLandOnEntry: true}
	spell := creature(cardID(12), 0, "One-drop", 1, 1)
	spell.ManaCost = "{G}"
	for _, established := range []bool{false, true} {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(bounce, forest, spell)), newSeat(1)}, withTurn(1, 0, "precombat_main"))
		if established {
			v.Battlefield = protocol.ZoneView{Cards: []protocol.CardView{planForest(cardID(20))}, Count: 1}
		}
		in := input(0, v, passMove(0), landMove(t, 0, bounce.InstanceID), landMove(t, 0, forest.InstanceID))
		want := forest.InstanceID
		if established {
			// No new spell is unlocked; the bounce land is now a useful
			// land drop and can retain its enumeration-order preference.
			want = bounce.InstanceID
		}
		d := decide(t, heuristic.New(), in)
		if in.Moves[d.Index].Source.String() != want {
			t.Fatalf("established=%v: chose %s (%s)", established, in.Moves[d.Index].Source, d.Reason)
		}
		cfg := heuristic.DefaultConfig()
		cfg.OrderLandDrops = false
		old := decide(t, heuristic.NewWithConfig(cfg), in)
		if in.Moves[old.Index].Source.String() != bounce.InstanceID {
			t.Fatal("back-out must preserve the first land")
		}
	}
}

func TestLandOrderUnlocksTheRightColourAndKeepsIdleTaplands(t *testing.T) {
	for _, tc := range []struct {
		name, cost   string
		already      int
		wantUntapped bool
	}{
		{"two-drop on time", "{1}{R}", 1, true},
		{"no spell waiting", "", 1, false},
		{"wrong colour", "{1}{G}", 1, false},
		{"already payable", "{1}{R}", 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tapland := land(cardID(10), 0)
			tapland.Name, tapland.TypeLine = "Tapland", "Land"
			tapland.Purpose = &protocol.PurposeView{LandEntry: "tapped"}
			mountain := land(cardID(11), 0)
			hand := []protocol.CardView{tapland, mountain}
			if tc.cost != "" {
				spell := creature(cardID(12), 0, "Two-drop", 2, 2)
				spell.ManaCost = tc.cost
				hand = append(hand, spell)
			}
			var bf []protocol.CardView
			for i := range tc.already {
				bf = append(bf, land(cardID(20+i), 0))
			}
			v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)}, withBattlefield(bf...), withTurn(2, 0, "precombat_main"))
			moves := []legal.Move{passMove(0), landMove(t, 0, tapland.InstanceID), landMove(t, 0, mountain.InstanceID)}
			in := input(0, v, moves...)
			d := decide(t, heuristic.New(), in)
			want := tapland.InstanceID
			if tc.wantUntapped {
				want = mountain.InstanceID
			}
			if moves[d.Index].Source.String() != want {
				t.Fatalf("chose %s (%s), want %s", moves[d.Index].Source, d.Reason, want)
			}
		})
	}
	if heuristic.BaselineConfig().OrderLandDrops {
		t.Fatal("baseline must keep the old ordering")
	}
}

// Entry uncertainty is not evidence of a tempo cost. The mana bonus
// must not reorder a conditional land against a known untapped one.
func TestLandOrderDoesNotTreatUnknownEntryAsTapped(t *testing.T) {
	unknown := land(cardID(10), 0)
	unknown.Name, unknown.TypeLine = "Conditional dual", "Land"
	unknown.ManaAbilities[0].Produced = "{G|R}"
	mountain := land(cardID(11), 0)
	spell := creature(cardID(12), 0, "Two-drop", 2, 2)
	spell.ManaCost = "{1}{R}"
	v := newView([]protocol.PlayerView{newSeat(0, withHand(unknown, mountain, spell)), newSeat(1)},
		withBattlefield(land(cardID(20), 0)), withTurn(2, 0, "precombat_main"))
	in := input(0, v, passMove(0), landMove(t, 0, unknown.InstanceID), landMove(t, 0, mountain.InstanceID))
	d := decide(t, heuristic.New(), in)
	if in.Moves[d.Index].Source.String() != unknown.InstanceID {
		t.Fatalf("speculated about an unknown entry: %s (%s)", in.Moves[d.Index].Source, d.Reason)
	}
}
