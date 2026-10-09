package heuristic

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// plan_internal_test.go is ADR 0136 §2's mana model, tested on its own:
// what the bot can make from the view, and whether a set of costs can be
// paid from it.

func planSeatView(me uuid.UUID, bf []protocol.CardView, pool ...string) protocol.GameView {
	var v protocol.GameView
	v.Seats = []protocol.PlayerView{{ID: me.String(), Life: 40, ManaPool: pool}}
	v.Turn.ActiveSeat = 0
	v.Turn.Step = "precombat_main"
	for i := range bf {
		bf[i].Controller = me.String()
		bf[i].Owner = me.String()
		if bf[i].InstanceID == "" {
			bf[i].InstanceID = uuid.NewString()
		}
	}
	v.Battlefield.Cards = bf
	return v
}

func planLand(produced string, opts ...func(*protocol.CardView)) protocol.CardView {
	c := protocol.CardView{Name: "Land", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: produced}}}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func sortedUnits(u []uint8) []uint8 {
	out := append([]uint8(nil), u...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func TestManaModelReadsWhatTheSeatCanMakeNow(t *testing.T) {
	me := uuid.New()
	orchard := planLand("", func(c *protocol.CardView) {
		c.ManaAbilities[0].ColorOptions = [][]string{{"G", "U"}}
	})
	blankOrchard := planLand("", func(c *protocol.CardView) {
		c.ManaAbilities[0].AddsNoMana = false
	})
	signet := protocol.CardView{Name: "Simic Signet", TypeLine: "Artifact",
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, ManaCost: "{1}", Produced: "{G}{U}"}}}
	sickDork := protocol.CardView{Name: "Llanowar Elves", TypeLine: "Creature — Elf", SummoningSick: true,
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{G}"}}}
	readyDork := protocol.CardView{Name: "Birds of Paradise", TypeLine: "Creature — Bird",
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{W|U|B|R|G}", ColorOptions: [][]string{{"G", "U"}}}}}
	tappedLand := planLand("{G}", func(c *protocol.CardView) { c.Tapped = true })
	greyed := planLand("{C}", func(c *protocol.CardView) { c.ManaAbilities[0].ConditionUnmet = true })
	treasure := protocol.CardView{Name: "Treasure", TypeLine: "Artifact — Treasure",
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}}
	solRing := protocol.CardView{Name: "Sol Ring", TypeLine: "Artifact",
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}{C}"}}}
	painland := protocol.CardView{Name: "Yavimaya Coast", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
		{TapCost: true, Produced: "{C}"}, {TapCost: true, Produced: "{G|U}", LifeCost: 1}}}

	cases := []struct {
		name string
		bf   []protocol.CardView
		pool []string
		want []uint8
	}{
		{"an Island", []protocol.CardView{planLand("{U}")}, nil, []uint8{manaU}},
		{"an Orchard names its colours", []protocol.CardView{orchard}, nil, []uint8{manaG | manaU}},
		{"an output the view cannot size is one of no colour", []protocol.CardView{blankOrchard}, nil, []uint8{manaC}},
		{"a filter eats a land and nets one", []protocol.CardView{planLand("{C}"), signet}, nil, []uint8{manaU, manaG}},
		{"a filter with nothing to eat adds nothing", []protocol.CardView{signet}, nil, nil},
		{"a sick dork makes nothing, a ready one its colours", []protocol.CardView{sickDork, readyDork}, nil, []uint8{manaG | manaU}},
		{"tapped, greyed and one-shot sources make nothing", []protocol.CardView{tappedLand, greyed, treasure}, nil, nil},
		{"Sol Ring makes two colourless", []protocol.CardView{solRing}, nil, []uint8{manaC, manaC}},
		{"a painland is one mana of any of its colours", []protocol.CardView{painland}, nil, []uint8{manaG | manaU | manaC}},
		{"Mana Confluence is painful", []protocol.CardView{{Name: "Mana Confluence", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
			{TapCost: true, Produced: "{W|U|B|R|G}", LifeCost: 1}}}}, nil, []uint8{manaAnyColor | manaPain}},
		{"Ancient Tomb's damage is painful", []protocol.CardView{{Name: "Ancient Tomb", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
			{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}. Ancient Tomb deals 2 damage to you."}}}}, nil, []uint8{manaC | manaPain, manaC | manaPain}},
		{"the floating pool counts", nil, []string{"R", "C"}, []uint8{manaR, manaC}},
		{"restricted mana is left out", []protocol.CardView{{Name: "Delighted Halfling", TypeLine: "Creature — Halfling", ManaAbilities: []protocol.ManaAbilityView{
			{TapCost: true, Produced: "{C}"}, {TapCost: true, Produced: "{W|U|B|R|G}", Restrictions: []string{"legendary"}}}}}, nil, []uint8{manaC}},
		{"a fetchland makes no mana", []protocol.CardView{{Name: "Fabled Passage", TypeLine: "Land"}}, nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := New()
			st := p.newState(aiseat.Input{Seat: me, View: planSeatView(me, c.bf, c.pool...)})
			got := sortedUnits(st.manaAvailable(false))
			want := sortedUnits(c.want)
			if len(got) != len(want) {
				t.Fatalf("units = %v, want %v", got, want)
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("units = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestManaModelPaysColouredSymbolsFirst(t *testing.T) {
	cases := []struct {
		name  string
		units []uint8
		costs []string
		ok    bool
	}{
		{"{1}{G} from an Island and a Forest", []uint8{manaU, manaG}, []string{"{1}{G}"}, true},
		{"{G}{G} from an Island and a Forest", []uint8{manaU, manaG}, []string{"{G}{G}"}, false},
		{"the flexible source is kept for the colour that needs it", []uint8{manaG | manaU, manaG}, []string{"{G}", "{U}"}, true},
		{"a hybrid symbol takes either half", []uint8{manaU}, []string{"{G/U}"}, true},
		{"a Phyrexian symbol is paid with its colour", []uint8{manaR}, []string{"{R/P}"}, true},
		{"{C} wants colourless", []uint8{manaG}, []string{"{C}"}, false},
		{"generic takes anything", []uint8{manaC, manaW}, []string{"{2}"}, true},
		{"two casts, five of five", []uint8{manaG, manaG, manaU, manaC, manaC}, []string{"{2}", "{1}{G}{U}"}, true},
		{"two casts, six of five", []uint8{manaG, manaG, manaU, manaC, manaC}, []string{"{3}", "{1}{G}{U}"}, false},
		// The auto-tapper spends painless sources first, so Hedron
		// Archive is paid with the Swamps and leaves Ancient Tomb.
		{"a painful source is spent last", []uint8{manaC | manaPain, manaC | manaPain, manaB, manaB, manaB, manaB}, []string{"{4}", "{B}{B}"}, false},
		{"pain is spent when it is all there is", []uint8{manaC | manaPain, manaC | manaPain, manaB, manaB}, []string{"{2}", "{B}{B}"}, false},
		{"colourless before a colour", []uint8{manaC | manaPain, manaC | manaPain, manaB, manaB, manaB, manaB}, []string{"{B}{B}", "{2}"}, true},
		// The auto-tapper does not know what comes next: a tie between a
		// Mountain and an Island may spend the Island the next spell needs.
		{"a tie is spent the worst way", []uint8{manaR, manaU}, []string{"{1}", "{U}"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			units := append([]uint8(nil), c.units...)
			ok := true
			for i, cost := range c.costs {
				var later []uint8
				for _, next := range c.costs[i+1:] {
					later = append(later, parseManaCost(next, 0).colored...)
				}
				if units, ok = payMana(units, parseManaCost(cost, 0), later); !ok {
					break
				}
			}
			if ok != c.ok {
				t.Errorf("paid = %v, want %v", ok, c.ok)
			}
		})
	}
}

func TestPlanCastCostReadsTheMoveThenThePrintedCost(t *testing.T) {
	me := uuid.New()
	cmdr := protocol.CardView{InstanceID: uuid.NewString(), Name: "Tatyova", ManaCost: "{3}{G}{U}", TypeLine: "Legendary Creature"}
	v := planSeatView(me, nil)
	v.Seats[0].Command.Cards = []protocol.CardView{cmdr}
	v.Seats[0].CommanderCasts = map[string]int{cmdr.InstanceID: 1}
	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: v})

	stamped := legal.Move{Kind: legal.KindCast, Cost: &legal.MoveCost{Mana: "{7}{G}{U}"}}
	if c, ok := st.planCastCost(stamped, &cmdr, castParams{FromZone: "command"}); !ok || c.total() != 9 {
		t.Errorf("stamped cost = %+v (%v), want the move's 9", c, ok)
	}
	// A view from before legal.MoveCost.Mana: the printed cost and the
	// tax (CR 903.8).
	if c, ok := st.planCastCost(legal.Move{Kind: legal.KindCast}, &cmdr, castParams{FromZone: "command"}); !ok || c.total() != 7 {
		t.Errorf("unstamped commander cost = %+v (%v), want 5 + 2 tax", c, ok)
	}
	x := protocol.CardView{Name: "Blaze", ManaCost: "{X}{R}"}
	if c, ok := st.planCastCost(legal.Move{Kind: legal.KindCast}, &x, castParams{XValue: 3}); !ok || c.total() != 4 {
		t.Errorf("X cost = %+v (%v), want 4", c, ok)
	}
	if _, ok := st.planCastCost(legal.Move{Kind: legal.KindCast}, &x, castParams{AlternativeCost: "flashback"}); ok {
		t.Error("an unstamped alternative cost was planned")
	}
}

func TestPlanMembersAreCastsWithOnlyManaCosts(t *testing.T) {
	raw := func(m map[string]any) json.RawMessage { b, _ := json.Marshal(m); return b }
	cases := []struct {
		name string
		m    legal.Move
		ok   bool
	}{
		{"a plain cast", legal.Move{Kind: legal.KindCast, Params: raw(map[string]any{"instance_id": "x"})}, true},
		{"a cast with a discard", legal.Move{Kind: legal.KindCast, Params: raw(map[string]any{"discard_ids": []string{"y"}})}, false},
		{"a cast with a sacrifice", legal.Move{Kind: legal.KindCast, Params: raw(map[string]any{"sacrifice_ids": []string{"y"}})}, false},
		{"a cast that pays life", legal.Move{Kind: legal.KindCast, Cost: &legal.MoveCost{Life: 2}}, false},
		{"an activation", legal.Move{Kind: legal.KindActivate}, false},
	}
	for _, c := range cases {
		if got := planEligible(c.m); got != c.ok {
			t.Errorf("%s: eligible = %v, want %v", c.name, got, c.ok)
		}
	}
}

// TestPlanCastsTheSweepBeforeItsPermanents: a sweep is priced against
// the board as it stands, so in a plan it goes before the permanents
// the plan adds, whichever is worth more.
func TestPlanCastsTheSweepBeforeItsPermanents(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	var bf []protocol.CardView
	for i := 0; i < 6; i++ {
		bf = append(bf, planLand("{W}"))
	}
	v := planSeatView(me, bf)
	v.Seats = append(v.Seats, protocol.PlayerView{ID: opp.String(), Life: 40})
	wrath := protocol.CardView{InstanceID: uuid.NewString(), Name: "Wrath", ManaCost: "{2}{W}{W}", TypeLine: "Sorcery",
		Purpose: &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "destroy"}}}
	angel := protocol.CardView{InstanceID: uuid.NewString(), Name: "Angel", ManaCost: "{1}{W}", TypeLine: "Creature — Angel", Power: 4, Toughness: 4}
	v.Seats[0].Hand.Cards = []protocol.CardView{angel, wrath}
	v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: uuid.NewString(), Name: "Ogre",
		TypeLine: "Creature — Ogre", Power: 1, Toughness: 1, Controller: opp.String(), Owner: opp.String()})
	cast := func(c protocol.CardView) legal.Move {
		raw, _ := json.Marshal(map[string]any{"instance_id": c.InstanceID, "from_zone": "hand"})
		return legal.Move{Kind: legal.KindCast, Label: "Cast " + c.Name, Source: uuid.MustParse(c.InstanceID), Params: raw,
			Cost: &legal.MoveCost{Mana: c.ManaCost}}
	}
	moves := []legal.Move{cast(angel), cast(wrath)}
	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: v, Moves: moves})
	vals := make([]float64, len(moves))
	for i := range moves {
		vals[i], _ = p.valueOf(st, moves[i])
	}
	if vals[0] <= vals[1] {
		t.Fatalf("the Angel (%v) is not worth more than the sweep (%v); the order would not be tested", vals[0], vals[1])
	}
	cands := p.planCandidates(st, moves, vals)
	if len(cands) != 2 || cands[0].card.Name != "Wrath" {
		t.Errorf("order = %v, want the sweep first", []string{cands[0].card.Name, cands[1].card.Name})
	}
}

// TestPlanSharesTheRampPremium is §3: two rocks in one turn do not both
// claim the whole deficit, and the set's premium is measured against
// the cards not in it.
func TestPlanSharesTheRampPremium(t *testing.T) {
	me := uuid.New()
	rock := func(name string) protocol.CardView {
		return protocol.CardView{InstanceID: uuid.NewString(), Name: name, ManaCost: "{1}", TypeLine: "Artifact",
			ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}"}}}
	}
	a, b := rock("Rock A"), rock("Rock B")
	four := protocol.CardView{InstanceID: uuid.NewString(), Name: "Four-drop", ManaCost: "{4}", TypeLine: "Sorcery"}
	v := planSeatView(me, []protocol.CardView{planLand("{C}"), planLand("{C}"), planLand("{C}")})
	v.Seats[0].Hand.Cards = []protocol.CardView{a, b, four}
	cast := func(c protocol.CardView) legal.Move {
		raw, _ := json.Marshal(map[string]any{"instance_id": c.InstanceID, "from_zone": "hand"})
		return legal.Move{Kind: legal.KindCast, Label: "Cast " + c.Name, Source: uuid.MustParse(c.InstanceID), Params: raw,
			Cost: &legal.MoveCost{Mana: c.ManaCost}}
	}
	moves := []legal.Move{cast(a), cast(b)}
	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: v, Moves: moves})
	vals := make([]float64, len(moves))
	for i := range moves {
		vals[i], _ = p.valueOf(st, moves[i])
	}
	// The deficit is one (want 4, three sources), so each rock alone
	// claims all of it.
	noRamp := &Policy{cfg: p.cfg}
	noRamp.cfg.RampPerMana = 0
	base, _ := noRamp.valueOf(st, moves[0])
	if prem := vals[0] - base; math.Abs(prem-p.cfg.RampPerMana) > 1e-9 {
		t.Fatalf("one rock's premium = %v, want RampPerMana × 1", prem)
	}
	pl, ok := p.planTurn(context.Background(), st, moves, vals)
	if !ok {
		t.Fatal("no plan of the two rocks")
	}
	if want := vals[0] + vals[1] - p.cfg.RampPerMana; math.Abs(pl.value-want) > 1e-9 {
		t.Errorf("two rocks are worth %v, want %v: the one mana of deficit is claimed once", pl.value, want)
	}
}

// TestManaModelRunsAFilterLand is Flooded Grove under
// Config.PlanFilterLands (ADR 0136 §2): "{G/U}, {T}: Add {G}{G}, {G}{U},
// or {U}{U}" is a filter, fed by a green or blue mana, rather than its
// plain {C} row. Off, the land is its {C} row, as PR 4 shipped. With
// nothing to feed it, it is still its {C}.
func TestManaModelRunsAFilterLand(t *testing.T) {
	me := uuid.New()
	grove := func() protocol.CardView {
		return protocol.CardView{Name: "Flooded Grove", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
			{TapCost: true, Produced: "{C}"},
			{TapCost: true, ManaCost: "{G/U}", Produced: "{G|U}{G|U}", ColorOptions: [][]string{{"G", "U"}, {"G", "U"}}}}}
	}
	gu := manaG | manaU
	cases := []struct {
		name string
		on   bool
		bf   []protocol.CardView
		want []uint8
	}{
		{"off: its {C} row", false, []protocol.CardView{planLand("{U}"), grove()}, []uint8{manaU, manaC}},
		{"on: an Island feeds it", true, []protocol.CardView{planLand("{U}"), grove()}, []uint8{gu, gu}},
		{"on: a Swamp cannot feed it", true, []protocol.CardView{planLand("{B}"), grove()}, []uint8{manaB, manaC}},
		{"on: alone it is its {C}", true, []protocol.CardView{grove()}, []uint8{manaC}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := New()
			st := p.newState(aiseat.Input{Seat: me, View: planSeatView(me, c.bf)})
			got := sortedUnits(st.manaAvailable(c.on))
			want := sortedUnits(c.want)
			if len(got) != len(want) {
				t.Fatalf("units = %v, want %v", got, want)
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("units = %v, want %v", got, want)
				}
			}
		})
	}
}
