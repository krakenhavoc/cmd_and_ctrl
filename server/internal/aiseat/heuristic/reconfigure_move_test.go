package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestReconfigureDoesNotPayToAttachForNoGain(t *testing.T) {
	for _, tc := range []struct {
		name     string
		target   string
		attached bool
		wantPass bool
	}{
		{"current host", cardID(1), true, true},
		{"worse host", cardID(2), true, true},
		{"better host", cardID(4), true, false},
		{"first attach", cardID(1), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sash := equipment(cardID(3), 0, "Lion Sash")
			if tc.attached {
				sash.AttachedTo = &protocol.TargetRefView{Kind: "card", ID: cardID(1)}
			}
			sash.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Reconfigure: true}}
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(
				creature(cardID(1), 0, "Adeline", 4, 4),
				creature(cardID(2), 0, "Small creature", 1, 1),
				creature(cardID(4), 0, "Better creature", 8, 8), sash,
			))
			m := equipMove(t, 0, cardID(3), tc.target)
			m.Label = "Reconfigure {2} (attach) → " + tc.target
			in := input(0, v, passMove(0), m)
			on := heuristic.New()
			got := rankValue(t, on, in, m.Label)
			if tc.wantPass {
				if got != -1 || chose(t, in, decide(t, on, in)) != passMove(0).Label {
					t.Errorf("idle reconfigure scores %g; want -1 and a pass", got)
				}
			} else if got <= 0 || chose(t, in, decide(t, on, in)) != m.Label {
				t.Errorf("useful reconfigure scores %g; want a positive attach", got)
			}
			cfg := heuristic.DefaultConfig()
			cfg.PriceReconfigure = false
			if got := rankValue(t, heuristic.NewWithConfig(cfg), in, m.Label); got <= 0 {
				t.Errorf("back-out scores %g; want the old positive activation", got)
			}
		})
	}
}

func TestReconfigureDoesNotDetachWithoutAPricedBenefit(t *testing.T) {
	sash := equipment(cardID(3), 0, "Lion Sash", attachedToCard(cardID(1)))
	sash.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Reconfigure: true}}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(
		creature(cardID(1), 0, "Adeline", 4, 4), sash,
	))
	m := activateMove(t, 0, cardID(3), "Reconfigure {2} (unattach)", nil)
	in := input(0, v, passMove(0), m)
	if got := rankValue(t, heuristic.New(), in, m.Label); got != -1 {
		t.Errorf("unattach scores %g, want -1", got)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
		t.Errorf("chose %q, want a pass", got)
	}
	cfg := heuristic.DefaultConfig()
	cfg.PriceReconfigure = false
	if got := rankValue(t, heuristic.NewWithConfig(cfg), in, m.Label); got <= 0 {
		t.Errorf("back-out unattach scores %g, want the old positive activation", got)
	}
	if heuristic.BaselineConfig().PriceReconfigure {
		t.Error("baseline enables reconfigure pricing")
	}
}
