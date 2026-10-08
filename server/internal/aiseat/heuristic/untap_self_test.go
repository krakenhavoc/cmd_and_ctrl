package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// untap_self_test.go — #2500. A heuristic seat repeated "{3}: Untap
// Basalt Monolith": tap for {C}{C}{C}, pay {3}, again, for nothing,
// until the CR 732 breaker and the runner's #810 hold parked it.

// monolith is a tapped Monolith as the wire sends it: a {T}: Add
// {C}{C}{C} mana ability and a self-untap row at the given cost.
func monolith(id string, controller int, name, untapCost string) protocol.CardView {
	c := rock(id, controller, name, "{3}", protocol.ManaAbilityView{TapCost: true, Produced: "{C}{C}{C}"})
	c.Tapped = true
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: untapCost + ": Untap " + name, ManaCost: untapCost, UntapSelf: true,
	}}
	return c
}

func untapMove(t *testing.T, seat int, src, label string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0}),
	}
}

func TestTheBotDoesNotUntapAMonolithForNoNetMana(t *testing.T) {
	for _, tc := range []struct{ name, cost string }{
		{"Basalt Monolith", "{3}"},
		{"Grim Monolith", "{4}"},
	} {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(monolith(cardID(1), 0, tc.name, tc.cost), land(cardID(10), 0), land(cardID(11), 0), land(cardID(12), 0), land(cardID(13), 0)))
		in := input(0, v, passMove(0), untapMove(t, 0, cardID(1), tc.cost+": Untap "+tc.name))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
			t.Errorf("%s: the bot chose %q: paying %s to untap a source that makes {C}{C}{C} nets nothing (#2500)", tc.name, got, tc.cost)
		}
	}
}

// The control: the rule is about the row's net, not about untaps. A
// self-untap that costs less than the source makes still pays.
func TestTheBotStillUntapsASourceThatNetsMana(t *testing.T) {
	c := monolith(cardID(1), 0, "Cheap Monolith", "{1}")
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(c, land(cardID(10), 0)))
	move := untapMove(t, 0, cardID(1), "{1}: Untap Cheap Monolith")
	in := input(0, v, passMove(0), move)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != move.Label {
		t.Errorf("the bot chose %q over an untap that nets two mana", got)
	}
}
