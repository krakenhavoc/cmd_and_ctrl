package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #3027: Strip Mine can legally target itself, but paying its sacrifice
// cost for no gain must rank below passing. The price depends on the row's
// printed cost, including on a different card with the same shape.
func TestTargetedActivationPaysForItsSourceSacrifice(t *testing.T) {
	for _, name := range []string{"Strip Mine", "Another land"} {
		t.Run(name, func(t *testing.T) {
			src := land(cardID(1), 0)
			src.Name = name
			src.TypeLine = "Land"
			src.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, TapCost: true, SacrificeSelf: true}}
			own := land(cardID(2), 0)
			opposing := land(cardID(3), 1)
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
				withBattlefield(src, own, opposing), withTurn(1, 0, "precombat_main"))
			pol := heuristic.New()
			move := func(target protocol.CardView) legal.Move {
				return activate(t, 0, src.InstanceID, "Target "+target.InstanceID, map[string]any{
					"targets": []protocol.TargetRefView{{Kind: "card", ID: target.InstanceID}},
				})
			}
			for _, target := range []protocol.CardView{src, own} {
				in := input(0, v, passMove(0), move(target))
				if got := rankValue(t, pol, in, in.Moves[1].Label); got >= 0 {
					t.Errorf("own land %s priced %.3f, want below passing", target.InstanceID, got)
				}
				d, err := pol.Decide(context.Background(), in)
				if err != nil || in.Moves[d.Index].Kind != legal.KindPass {
					t.Fatalf("sacrificed source targeting own land: %+v, %v", d, err)
				}
			}
			in := input(0, v, passMove(0), move(src), move(own), move(opposing))
			d, err := pol.Decide(context.Background(), in)
			if err != nil || in.Moves[d.Index].Label != move(opposing).Label {
				t.Fatalf("did not choose worthwhile opposing land removal: %+v, %v", d, err)
			}
		})
	}
}

// ADR 0142 retains the proxy for an untargeted, unpriced row, such as
// Mind Stone. Charging targeted costs must not disable that fallback.
func TestUntargetedUnpricedSourceSacrificeKeepsItsProxy(t *testing.T) {
	src := land(cardID(1), 0)
	src.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, SacrificeSelf: true}}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(src))
	m := activate(t, 0, src.InstanceID, "Unknown untargeted ability", nil)
	pol := heuristic.New()
	in := input(0, v, passMove(0), m)
	if got, want := rankValue(t, pol, in, m.Label), heuristic.DefaultConfig().ActivateBase; !nearly(got, want) {
		t.Errorf("untargeted fallback priced %.3f, want %.3f", got, want)
	}
}
