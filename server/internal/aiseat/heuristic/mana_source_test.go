package heuristic_test

import (
	"context"
	"math"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_source_test.go is ADR 0126 §2: a mana source priced by the mana
// it makes on the battlefield, and a ramp premium at cast time scaled
// by how short the bot is of casting what it holds.

func rock(id string, controller int, name, cost string, abs ...protocol.ManaAbilityView) protocol.CardView {
	return protocol.CardView{
		InstanceID:    id,
		Name:          name,
		Owner:         seatID(controller).String(),
		Controller:    seatID(controller).String(),
		TypeLine:      "Artifact",
		ManaCost:      cost,
		KnownByYou:    true,
		ManaAbilities: abs,
	}
}

func solRing(id string, controller int) protocol.CardView {
	return rock(id, controller, "Sol Ring", "{1}", protocol.ManaAbilityView{TapCost: true, Produced: "{C}{C}"})
}

func arcaneSignet(id string, controller int) protocol.CardView {
	return rock(id, controller, "Arcane Signet", "{2}", protocol.ManaAbilityView{TapCost: true, Produced: "{W|U|B|R|G}"})
}

func llanowarElves(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Llanowar Elves", 1, 1)
	c.ManaCost = "{G}"
	c.TypeLine = "Creature — Elf Druid"
	c.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{G}"}}
	return c
}

func rampFiveDrop(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Big Thing", 5, 5)
	c.ManaCost = "{3}{G}{G}"
	return c
}

func manaLands(n, controller int, from int) []protocol.CardView {
	out := make([]protocol.CardView, n)
	for i := range out {
		out[i] = land(cardID(from+i), controller)
	}
	return out
}

func rankValue(t *testing.T, pol *heuristic.Policy, in aiseat.Input, label string) float64 {
	t.Helper()
	for _, c := range pol.Rank(context.Background(), in) {
		if in.Moves[c.Index].Label == label {
			return c.Value
		}
	}
	t.Fatalf("no ranked move %q", label)
	return 0
}

func nearly(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// The ADR's worked table, with its starting weights.
func TestManaSourceCastPrices(t *testing.T) {
	pol := heuristic.New()
	ring, signet, elf := cardID(1), cardID(2), cardID(3)
	cases := []struct {
		name   string
		hand   []protocol.CardView
		bf     []protocol.CardView
		cast   string
		label  string
		want   float64
		reason string
	}{
		{
			name: "Sol Ring, turn one, a five-drop in hand",
			hand: []protocol.CardView{solRing(ring, 0), rampFiveDrop(cardID(9), 0)},
			bf:   manaLands(1, 0, 100),
			cast: ring, label: "Cast Sol Ring",
			want: 2.0 + 2*1.0 - 1.2,
		},
		{
			name: "Arcane Signet, turn two, deficit 3",
			hand: []protocol.CardView{arcaneSignet(signet, 0), rampFiveDrop(cardID(9), 0)},
			bf:   manaLands(2, 0, 100),
			cast: signet, label: "Cast Arcane Signet",
			want: 1.0 + 1.0 - 1.2,
		},
		{
			name: "Arcane Signet, turn nine, deficit 0",
			hand: []protocol.CardView{arcaneSignet(signet, 0), rampFiveDrop(cardID(9), 0)},
			bf:   manaLands(7, 0, 100),
			cast: signet, label: "Cast Arcane Signet",
			want: 1.0 - 1.2,
		},
		{
			name: "Llanowar Elves, turn one, deficit 3",
			hand: []protocol.CardView{llanowarElves(elf, 0), func() protocol.CardView { c := rampFiveDrop(cardID(9), 0); c.ManaCost = "{2}{G}{G}"; return c }()},
			bf:   manaLands(1, 0, 100),
			cast: elf, label: "Cast Llanowar Elves",
			want: 0.9*(1+0.45) + 1.0 - 1.2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0, withHand(tc.hand...)), newSeat(1)},
				withBattlefield(tc.bf...), withTurn(1, 0, "precombat_main"))
			in := input(0, v, passMove(0), castMove(t, 0, tc.cast, tc.label))
			if got := rankValue(t, pol, in, tc.label); !nearly(got, tc.want) {
				t.Errorf("%s: priced %.3f, want %.3f", tc.label, got, tc.want)
			}
		})
	}
}

// The card being cast is not part of what the bot wants to cast: a
// Signet in a hand of nothing else closes no gap.
func TestRampPremiumIgnoresTheCardBeingCast(t *testing.T) {
	signet := cardID(2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(arcaneSignet(signet, 0))), newSeat(1)},
		withBattlefield(manaLands(1, 0, 100)...), withTurn(1, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, signet, "Cast Arcane Signet"))
	if got := rankValue(t, heuristic.New(), in, "Cast Arcane Signet"); !nearly(got, 1.0-1.2) {
		t.Errorf("priced %.3f, want %.3f", got, 1.0-1.2)
	}
}

// The commander, with its tax, is part of what the bot wants; the cap
// keeps a ten-drop from asking for ten sources.
func TestRampPremiumCountsTheCommanderAndTheCap(t *testing.T) {
	ring := cardID(1)
	cmdr := creature(cardID(999), 0, "Commander", 4, 4, commander())
	cmdr.ManaCost = "{2}{G}{G}{G}" // 5, +2 tax after one cast = 7
	seat := newSeat(0, withHand(solRing(ring, 0)), withCommanderCasts(1))
	seat.Command = protocol.ZoneView{Kind: "command", Count: 1, Cards: []protocol.CardView{cmdr}}
	v := newView([]protocol.PlayerView{seat, newSeat(1)},
		withBattlefield(manaLands(6, 0, 100)...), withTurn(6, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, ring, "Cast Sol Ring"))
	// want 7, sources 6: one mana of the two closes the gap.
	if got := rankValue(t, heuristic.New(), in, "Cast Sol Ring"); !nearly(got, 2.0+1.0-1.2) {
		t.Errorf("priced %.3f, want %.3f", got, 2.0+1.0-1.2)
	}
}

// A one-shot source (a Treasure-like sacrifice) and a land are not
// repeatable rocks; a Signet nets what it adds less what it eats.
func TestRepeatableManaReadsTheView(t *testing.T) {
	w := heuristic.DefaultWeights()
	me := seatID(0).String()
	score := func(c protocol.CardView) float64 {
		with := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(c))
		without := newView([]protocol.PlayerView{newSeat(0), newSeat(1)})
		return w.Score(with, me) - w.Score(without, me)
	}
	// Scored against an empty opposition, so the mean and max terms
	// are constant; the difference is the permanent's own price.
	cases := []struct {
		name string
		card protocol.CardView
		want float64
	}{
		{"Sol Ring", solRing(cardID(1), 0), 2.0},
		{"tapped Sol Ring", func() protocol.CardView { c := solRing(cardID(1), 0); c.Tapped = true; return c }(), 0.55 * 2},
		{"Arcane Signet", arcaneSignet(cardID(1), 0), 1.0},
		{"Azorius Signet nets one", rock(cardID(1), 0, "Azorius Signet", "{2}", protocol.ManaAbilityView{TapCost: true, ManaCost: "{1}", Produced: "{W}{U}"}), 1.0},
		{"Hedron Archive", rock(cardID(1), 0, "Hedron Archive", "{4}", protocol.ManaAbilityView{TapCost: true, Produced: "{C}{C}"}), 2.0},
		{"Gilded Lotus", rock(cardID(1), 0, "Gilded Lotus", "{5}", protocol.ManaAbilityView{TapCost: true, Produced: "{W3|U3|B3|R3|G3}"}), 3.0},
		{"Lotus Petal is one-shot", rock(cardID(1), 0, "Lotus Petal", "{0}", protocol.ManaAbilityView{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}), 1.0},
		{"Ancient Tomb is a land", func() protocol.CardView {
			c := land(cardID(1), 0)
			c.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}{C}"}}
			return c
		}(), 1.0},
	}
	for _, tc := range cases {
		if got := score(tc.card); !nearly(got, tc.want) {
			t.Errorf("%s: worth %.3f on the battlefield, want %.3f", tc.name, got, tc.want)
		}
	}
}

// BaselineConfig prices every mana source as before ADR 0126.
func TestBaselineConfigPricesManaSourcesAsBefore(t *testing.T) {
	pol := heuristic.NewWithConfig(heuristic.BaselineConfig())
	ring := cardID(1)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(solRing(ring, 0), rampFiveDrop(cardID(9), 0))), newSeat(1)},
		withBattlefield(manaLands(1, 0, 100)...), withTurn(1, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, ring, "Cast Sol Ring"))
	if got := rankValue(t, pol, in, "Cast Sol Ring"); !nearly(got, 1.0-1.2) {
		t.Errorf("baseline priced Sol Ring at %.3f, want %.3f", got, 1.0-1.2)
	}
	if d, _ := pol.Decide(context.Background(), in); in.Moves[d.Index].Kind != legal.KindPass {
		t.Errorf("baseline cast Sol Ring: %+v", d)
	}
	if d, _ := heuristic.New().Decide(context.Background(), in); in.Moves[d.Index].Kind != legal.KindCast {
		t.Errorf("the heuristic did not cast Sol Ring: %+v", d)
	}
}
