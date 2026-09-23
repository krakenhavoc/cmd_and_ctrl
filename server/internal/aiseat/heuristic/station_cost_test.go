package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// station_cost_test.go — #759: station's cost taps another creature
// (CR 702.184a), and the policy has to weigh that creature's attack
// against the charge counters. The rule it lands on: before combat on
// its own turn a bot does not tap a creature that could swing, and
// after combat — or with a creature that could not attack anyway — it
// stations.

const stationMove = "The Seriema: Station (tapping Bear)"

func stationActivate(t *testing.T, seat int, src, tapped string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: stationMove, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0, "tap_ids": []string{tapped},
		}),
	}
}

func seriemaView(id string, controller int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "The Seriema", TypeLine: "Legendary Artifact — Spacecraft",
		Owner: seatID(controller).String(), Controller: seatID(controller).String(),
		ManaCost: "{1}{W}{W}", KnownByYou: true,
	}
}

func TestStationWeighsTheTappedCreaturesAttack(t *testing.T) {
	sick := func(c *protocol.CardView) { c.SummoningSick = true }
	cases := []struct {
		name string
		step string
		opts []func(*protocol.CardView)
		want bool // station?
	}{
		{"before combat, a creature that could attack", "precombat_main", nil, false},
		{"after combat, the same creature", "postcombat_main", nil, true},
		{"before combat, a summoning-sick creature", "precombat_main", []func(*protocol.CardView){sick}, true},
	}
	for _, tc := range cases {
		bear := creature(cardID(2), 0, "Bear", 3, 3)
		for _, o := range tc.opts {
			o(&bear)
		}
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withTurn(3, 0, tc.step),
			withBattlefield(seriemaView(cardID(1), 0), bear, land(cardID(10), 0)))
		in := input(0, v, passMove(0), stationActivate(t, 0, cardID(1), cardID(2)))
		got := chose(t, in, decide(t, heuristic.New(), in))
		if stationed := got == stationMove; stationed != tc.want {
			t.Errorf("%s: the bot chose %q, want station=%v", tc.name, got, tc.want)
		}
	}
}
