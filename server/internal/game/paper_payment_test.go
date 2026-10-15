package game

import (
	"errors"
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
