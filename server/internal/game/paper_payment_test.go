package game

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

// #3855: the shortfall a waived payment declares, and what counts as a
// legal declaration of it — exactly as many mana as are short, each
// coloured requirement paid by a symbol it admits (a hybrid by either
// half and never a third colour), the rest generic.
func TestPaperShortfallAndItsDeclaration(t *testing.T) {
	cost, err := ParseCost("{2}{W/U}{R}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{{Color: "R"}}
	short := pool.shortfallFor(cost, 0, ManaSpendContext{}, SpendPreserveColors)
	var symbols []string
	for _, s := range short.slots() {
		symbols = append(symbols, s.Symbol)
	}
	if !reflect.DeepEqual(symbols, []string{"{W/U}", "{1}", "{1}"}) || short.String() != "{2}{W/U}" {
		t.Fatalf("shortfall = %v (%s), want {W/U} {1} {1}", symbols, short)
	}
	if got := short.slots()[0].Options; !reflect.DeepEqual(got, []string{"W", "U"}) {
		t.Errorf("hybrid options = %v, want [W U]", got)
	}
	for _, tc := range []struct {
		name string
		decl PaperPayment
		ok   bool
	}{
		{"hybrid paid with its blue half", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "C"}, {Color: "G"}}}, true},
		{"in any order", PaperPayment{Mana: []PaperMana{{Color: "C"}, {Color: "W"}, {Color: "C"}}}, true},
		{"a third colour for the hybrid", PaperPayment{Mana: []PaperMana{{Color: "R"}, {Color: "C"}, {Color: "C"}}}, false},
		{"one short", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "C"}}}, false},
		{"one over", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "C"}, {Color: "C"}, {Color: "C"}}}, false},
		{"not a mana symbol", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "C"}, {Color: "S"}}}, false},
		{"free", PaperPayment{Free: true}, true},
		{"free with mana", PaperPayment{Free: true, Mana: []PaperMana{{Color: "U"}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := short.tokensFor(tc.decl)
			if (err == nil) != tc.ok {
				t.Fatalf("tokensFor = %v, want ok %v", err, tc.ok)
			}
			if err != nil && !errors.Is(err, ErrInvalidPaperPayment) {
				t.Errorf("err = %v, want ErrInvalidPaperPayment", err)
			}
		})
	}
}

// A declaration completes the record: the paper mana joins the pool's
// tokens with no source, OnPaper is false, and the readers are exact.
func TestDeclaredPaperPaymentMakesTheRecordExact(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	p.ManaPool = ManaPool{{Color: "G"}}
	item := castForTest(t, g, p, "{2}{G}", CastSpellParams{PaperPayment: &PaperPayment{
		Mana: []PaperMana{{Color: "W"}, {Color: "U", Artifact: true}},
	}})
	spent := item.Paid.Spent()
	if !spent.Known() || item.Paid.PaperMana != 0 || spent.Total() != 3 {
		t.Fatalf("record = %+v, want known and three mana", item.Paid)
	}
	if got := spent.Colors(); !reflect.DeepEqual(got, []string{"W", "U", "G"}) {
		t.Errorf("Colors = %v, want [W U G]", got)
	}
	if got := spent.CountFrom(ManaSourceArtifact); got != 1 {
		t.Errorf("artifact mana = %d, want the one declared", got)
	}
}

// #3904: a shortfall run longer than maxListedRun is one counted slot,
// and is answered by one counted entry.
func TestALongPaperShortfallIsOneCountedSlot(t *testing.T) {
	short := paperShortfall{
		reqs: []reqRun{
			{req: ColorRequirement{Options: []string{"W", "U"}}, n: 2},
			{req: ColorRequirement{Options: []string{"B"}}, n: MaxAnnouncedX},
		},
		generic: MaxAnnouncedX,
	}
	want := []PaperSlot{
		{Symbol: "{W/U}", Options: []string{"W", "U"}},
		{Symbol: "{W/U}", Options: []string{"W", "U"}},
		{Symbol: "{B}", Options: []string{"B"}, Count: MaxAnnouncedX},
		{Symbol: "{1}", Options: paperSymbols, Count: MaxAnnouncedX},
	}
	if got := short.slots(); !reflect.DeepEqual(got, want) {
		t.Fatalf("slots = %+v, want %+v", got, want)
	}
	if got, want := short.String(), "{1000000}{W/U}{W/U}{B}×1000000"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
	for _, tc := range []struct {
		name string
		decl PaperPayment
		ok   bool
	}{
		{"counted", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "W"}, {Color: "B", Count: MaxAnnouncedX}, {Color: "G", Count: MaxAnnouncedX}}}, true},
		{"split across entries", PaperPayment{Mana: []PaperMana{{Color: "B", Count: MaxAnnouncedX / 2}, {Color: "U", Count: 2}, {Color: "C", Count: MaxAnnouncedX}, {Color: "B", Count: MaxAnnouncedX / 2}}}, true},
		{"the coloured run paid in the wrong colour", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "W"}, {Color: "R", Count: MaxAnnouncedX}, {Color: "G", Count: MaxAnnouncedX}}}, false},
		{"the coloured run one black short", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "W"}, {Color: "B", Count: MaxAnnouncedX - 1}, {Color: "G", Count: MaxAnnouncedX + 1}}}, false},
		{"one short", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "W"}, {Color: "B", Count: MaxAnnouncedX}, {Color: "G", Count: MaxAnnouncedX - 1}}}, false},
		{"one over", PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "W"}, {Color: "B", Count: MaxAnnouncedX}, {Color: "G", Count: MaxAnnouncedX + 1}}}, false},
		{"a negative count", PaperPayment{Mana: []PaperMana{{Color: "U", Count: -1}}}, false},
		{"the largest count", PaperPayment{Mana: []PaperMana{{Color: "B", Count: math.MaxInt}}}, false},
		{"counts that would wrap", PaperPayment{Mana: []PaperMana{{Color: "B", Count: math.MaxInt/2 + 1}, {Color: "B", Count: math.MaxInt/2 + 1}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := short.tokensFor(tc.decl)
			if (err == nil) != tc.ok {
				t.Fatalf("tokensFor = %v, want ok %v", err, tc.ok)
			}
			if err != nil {
				if !errors.Is(err, ErrInvalidPaperPayment) {
					t.Errorf("err = %v, want ErrInvalidPaperPayment", err)
				}
				requireCheapRefusal(t, func() { _, _ = short.tokensFor(tc.decl) })
				return
			}
			if got := TotalMana(tokens); got != short.amount() || len(tokens) != len(tc.decl.Mana) {
				t.Errorf("%d tokens for %d mana, want %d runs for %d", len(tokens), got, len(tc.decl.Mana), short.amount())
			}
		})
	}
	if got, want := paperSymbolsOf(PaperPayment{Mana: []PaperMana{{Color: "C", Count: 2}, {Color: "B", Count: MaxAnnouncedX}}}), "{C}{C}{B}×1000000"; got != want {
		t.Errorf("paperSymbolsOf = %q, want %q", got, want)
	}
}

// End to end: a "spend only black mana on X" spell cast anyway at the X
// bound, which something reads, is refused with one counted slot, and
// the counted answer casts it with an exact record.
func TestCastAnywayAtTheBoundAsksForOneCountedSlot(t *testing.T) {
	withXBoundCatalog(t)
	const oracle = xBoundSpendOnlyOracle
	prev := CatalogReadsManaSpent
	CatalogReadsManaSpent = func(key string) ManaSpentReads {
		if key == oracle {
			return ReadsOwnPayment
		}
		if prev != nil {
			return prev(key)
		}
		return 0
	}
	t.Cleanup(func() { CatalogReadsManaSpent = prev })
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	id := xBoundSpell(me, "Test Drain", "{X}{B}", oracle)
	me.ManaPool.AddMana(ManaToken{Color: "B", Count: 10})
	params := CastSpellParams{XValue: MaxAnnouncedX, Strict: true, ForceCast: true}
	err := g.CastSpell(me.ID, id, params)
	var ask *PaperPaymentRequiredError
	if !errors.As(err, &ask) {
		t.Fatalf("err = %v, want *PaperPaymentRequiredError", err)
	}
	want := []PaperSlot{{Symbol: "{B}", Options: []string{"B"}, Count: MaxAnnouncedX + 1 - 10}}
	if !reflect.DeepEqual(ask.Slots, want) {
		t.Fatalf("slots = %+v, want %+v", ask.Slots, want)
	}
	params.PaperPayment = &PaperPayment{Mana: []PaperMana{{Color: "B", Count: MaxAnnouncedX + 1 - 10}}}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("the counted answer: %v", err)
	}
	item := g.StackMeta[id]
	if spent := item.Paid.Spent(); !spent.Known() || spent.Total() != MaxAnnouncedX+1 {
		t.Errorf("record %+v, want known and %d mana", item.Paid, MaxAnnouncedX+1)
	}
}
