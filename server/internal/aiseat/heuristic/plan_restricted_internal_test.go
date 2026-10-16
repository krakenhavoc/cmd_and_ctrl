package heuristic

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// plan_restricted_internal_test.go is ADR 0136's amendment of 2026-10-16
// (#3976): a mana row whose mana may be spent only on some spells counts
// toward the casts its restriction admits, and never toward the rest.
// The restriction tags are the wire's (ManaAbilityView.Restrictions); a
// policy test may not import internal/game for its constants.

var (
	legendaryOnly = []string{"purpose:cast", "supertype:Legendary"}
	elfOnly       = []string{"purpose:cast", "type:Creature", "subtype:Elf"}
	creatureOnly  = []string{"purpose:cast", "type:Creature"}
)

// halfling is Delighted Halfling: {C} for anything, or one mana of any
// colour for a legendary spell.
func halfling() protocol.CardView {
	return protocol.CardView{Name: "Delighted Halfling", TypeLine: "Creature — Halfling Citizen", ManaAbilities: []protocol.ManaAbilityView{
		{TapCost: true, Produced: "{C}"},
		{TapCost: true, Produced: "{W|U|B|R|G}", Restrictions: legendaryOnly}}}
}

// cavern is Cavern of Souls named for Elf, its restriction as the view
// stamps it.
func cavern() protocol.CardView {
	return protocol.CardView{Name: "Cavern of Souls", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
		{TapCost: true, Produced: "{C}"},
		{TapCost: true, Produced: "{W|U|B|R|G}", Restrictions: elfOnly}}}
}

func restrictedState(t *testing.T, bf []protocol.CardView, hand ...protocol.CardView) (*Policy, *state, uuid.UUID) {
	t.Helper()
	me := uuid.New()
	v := planSeatView(me, bf)
	v.Seats[0].Hand.Cards = hand
	p := New()
	return p, p.newState(aiseat.Input{Seat: me, View: v}), me
}

func costFor(st *state, cost string, card protocol.CardView) manaCost {
	c := parseManaCost(cost, 0)
	c.admits = st.castAdmits(&card, "")
	return c
}

func TestManaModelCountsARestrictedRowForTheCastsItAdmits(t *testing.T) {
	p, st, _ := restrictedState(t, []protocol.CardView{halfling()})
	if !p.cfg.PlanRestrictedMana {
		t.Fatal("PlanRestrictedMana is off by default")
	}
	if BaselineConfig().PlanRestrictedMana {
		t.Error("BaselineConfig counts restricted mana; the frozen heuristic never did")
	}

	off := st.manaAvailable(manaModel{})
	if len(off) != 1 || off[0] != manaC {
		t.Fatalf("off: units = %v, want the Halfling's {C} alone", off)
	}
	units := st.manaAvailable(p.manaModel(st))
	if len(units) != 1 || units[0]&manaAny != manaC || units[0].restricted() != manaAnyColor || units[0].class() == 0 {
		t.Fatalf("on: units = %#x, want {C} for anything and any colour for its class", units)
	}

	legend := protocol.CardView{Name: "Legend", TypeLine: "Legendary Creature — Elf", ManaCost: "{G}"}
	bear := protocol.CardView{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	cases := []struct {
		name string
		cost manaCost
		ok   bool
	}{
		{"a legendary spell's {G} is paid by the coloured row", costFor(st, "{G}", legend), true},
		{"a nonlegendary spell's {G} is not", costFor(st, "{G}", bear), false},
		{"a nonlegendary spell's {1} is paid by the {C} row", costFor(st, "{1}", bear), true},
		{"a cost with no cast admits nothing coloured", parseManaCost("{G}", 0), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, ok := payMana(append([]manaUnit(nil), units...), c.cost, nil); ok != c.ok {
				t.Errorf("paid = %v, want %v", ok, c.ok)
			}
		})
	}
}

// TestRestrictedOnlyManaWaitsForTheCastItAdmits is Ancient Ziggurat's
// shape: a source whose only row is restricted pays nothing for a spell
// it refuses, so that spell is paid around it and it stays for the
// creature after; and for a creature it is spent first, so the Forest
// stays up for the sorcery after.
func TestRestrictedOnlyManaWaitsForTheCastItAdmits(t *testing.T) {
	ziggurat := protocol.CardView{Name: "Ancient Ziggurat", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
		{TapCost: true, Produced: "{W|U|B|R|G}", Restrictions: creatureOnly}}}
	p, st, _ := restrictedState(t, []protocol.CardView{ziggurat, planLand("{G}")})
	units := st.manaAvailable(p.manaModel(st))
	if len(units) != 2 {
		t.Fatalf("units = %#x, want the Ziggurat's and the Forest's", units)
	}
	creature := protocol.CardView{Name: "Bear", TypeLine: "Creature — Bear"}
	sorcery := protocol.CardView{Name: "Rites", TypeLine: "Sorcery"}
	cases := []struct {
		name  string
		costs []manaCost
		ok    bool
	}{
		{"the sorcery, then the creature", []manaCost{costFor(st, "{G}", sorcery), costFor(st, "{G}", creature)}, true},
		{"the creature, then the sorcery", []manaCost{costFor(st, "{G}", creature), costFor(st, "{G}", sorcery)}, true},
		{"two sorceries", []manaCost{costFor(st, "{G}", sorcery), costFor(st, "{G}", sorcery)}, false},
		{"a sorcery's {1} is not paid by it either", []manaCost{costFor(st, "{2}", sorcery)}, false},
		{"two creatures", []manaCost{costFor(st, "{1}", creature), costFor(st, "{G}", creature)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			left := append([]manaUnit(nil), units...)
			ok := true
			for i, cost := range c.costs {
				var later []manaUnit
				for _, next := range c.costs[i+1:] {
					later = append(later, next.colored...)
				}
				if left, ok = payMana(left, cost, later); !ok {
					break
				}
			}
			if ok != c.ok {
				t.Errorf("paid = %v, want %v", ok, c.ok)
			}
		})
	}
}

// TestPlanCountsCavernManaForTheElvesItNames is the issue's case: an Elf
// deck with Cavern of Souls named for Elf and one Forest can cast two
// one-drop Elves this turn, and the plan says so. A Goblin is not paid by
// the Cavern's colour, so a Goblin and an Elf is still not a plan; and
// with PlanRestrictedMana off (ADR 0136 §2 as it shipped) neither is.
func TestPlanCountsCavernManaForTheElvesItNames(t *testing.T) {
	card := func(name, typeLine, cost string) protocol.CardView {
		return protocol.CardView{InstanceID: uuid.NewString(), Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1}
	}
	cast := func(c protocol.CardView) legal.Move {
		raw, _ := json.Marshal(map[string]any{"instance_id": c.InstanceID, "from_zone": "hand"})
		return legal.Move{Kind: legal.KindCast, Label: "Cast " + c.Name, Source: uuid.MustParse(c.InstanceID), Params: raw,
			Cost: &legal.MoveCost{Mana: c.ManaCost}}
	}
	plan := func(t *testing.T, restricted bool, hand ...protocol.CardView) bool {
		t.Helper()
		me := uuid.New()
		v := planSeatView(me, []protocol.CardView{cavern(), planLand("{G}")})
		v.Seats[0].Hand.Cards = hand
		var moves []legal.Move
		for _, c := range hand {
			moves = append(moves, cast(c))
		}
		p := New()
		p.cfg.PlanRestrictedMana = restricted
		st := p.newState(aiseat.Input{Seat: me, View: v, Moves: moves})
		vals := make([]float64, len(moves))
		for i := range moves {
			vals[i], _ = p.valueOf(st, moves[i])
		}
		pl, ok := p.planTurn(context.Background(), st, moves, vals)
		return ok && len(pl.members) == 2
	}
	elves := func() []protocol.CardView {
		return []protocol.CardView{card("Llanowar Elves", "Creature — Elf Druid", "{G}"), card("Elvish Mystic", "Creature — Elf Druid", "{G}")}
	}
	if !plan(t, true, elves()...) {
		t.Error("two Elves are not a plan with Cavern of Souls named for Elf and a Forest")
	}
	if plan(t, false, elves()...) {
		t.Error("with PlanRestrictedMana off, the Cavern's colour was counted")
	}
	mixed := []protocol.CardView{card("Llanowar Elves", "Creature — Elf Druid", "{G}"), card("Goblin Guide", "Creature — Goblin Scout", "{R}")}
	if plan(t, true, mixed...) {
		t.Error("a Goblin and an Elf are a plan: the Cavern's colour paid for a Goblin")
	}
}

// TestRestrictedOverlayIsExactOrNothing: a source is one activation, so
// its plain and restricted rows are overlaid only where no mix of the
// two can be spent that neither row makes.
func TestRestrictedOverlayIsExactOrNothing(t *testing.T) {
	const k = 1
	r := func(m manaUnit) manaUnit { return m<<manaRestrictedShift | k<<manaClassShift }
	cases := []struct {
		name        string
		plain, rest []manaUnit
		want        []manaUnit
		ok          bool
	}{
		{"no plain row (Ancient Ziggurat)", nil, []manaUnit{manaAnyColor}, []manaUnit{r(manaAnyColor)}, true},
		{"one and one (Delighted Halfling)", []manaUnit{manaC}, []manaUnit{manaAnyColor}, []manaUnit{manaC | r(manaAnyColor)}, true},
		{"{C} or {C}{C} (Eldrazi Temple)", []manaUnit{manaC}, []manaUnit{manaC, manaC}, []manaUnit{manaC | r(manaC), r(manaC)}, true},
		{"{C}{C} or {G}: the plain row alone", []manaUnit{manaC, manaC}, []manaUnit{manaG}, nil, false},
		{"{C}{C} or {G}{G}: a mix neither row makes", []manaUnit{manaC, manaC}, []manaUnit{manaG, manaG}, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := restrictedOverlay(c.plain, c.rest, k)
			if ok != c.ok || len(got) != len(c.want) {
				t.Fatalf("overlay = %#x, %v; want %#x, %v", got, ok, c.want, c.ok)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("overlay = %#x, want %#x", got, c.want)
				}
			}
		})
	}
}
