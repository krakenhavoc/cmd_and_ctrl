package game

import "testing"

// twobrid_payment_test.go — #4177, CR 107.4e / 601.2b: a two-mana hybrid
// symbol ({2/B}) is paid with one mana of its colour or with two mana of
// any kind. The pool solver and the auto-tapper both pay the colour when
// they can and the two generic when they cannot.

func tbPool(colors ...string) ManaPool {
	var p ManaPool
	for _, c := range colors {
		p.AddMana(ManaToken{Color: c})
	}
	return p
}

func TestTwobridIsPaidWithItsColour(t *testing.T) {
	if !tbPool("B", "B", "B").CanPay(costFor(t, "{2/B}{2/B}{2/B}"), 0) {
		t.Error("three black did not pay {2/B}{2/B}{2/B}")
	}
}

func TestTwobridIsPaidWithTwoOfAnything(t *testing.T) {
	if !tbPool("C", "C").CanPay(costFor(t, "{2/B}"), 0) {
		t.Error("two colourless did not pay {2/B}")
	}
	if !tbPool("C", "C", "C", "C", "C", "C").CanPay(costFor(t, "{2/B}{2/B}{2/B}"), 0) {
		t.Error("six colourless did not pay {2/B}{2/B}{2/B}")
	}
	if tbPool("C").CanPay(costFor(t, "{2/B}"), 0) {
		t.Error("one colourless paid a {2/B}")
	}
	if tbPool("C", "C", "C", "C", "C").CanPay(costFor(t, "{2/B}{2/B}{2/B}"), 0) {
		t.Error("five colourless paid {2/B}{2/B}{2/B}")
	}
}

// One black and four others: the colour pays one symbol, two others each
// of the rest.
func TestTwobridMixesTheColourAndGeneric(t *testing.T) {
	if !tbPool("B", "C", "C", "G", "R").CanPay(costFor(t, "{2/B}{2/B}{2/B}"), 0) {
		t.Error("B + four others did not pay {2/B}{2/B}{2/B}")
	}
	if tbPool("B", "C", "C", "G").CanPay(costFor(t, "{2/B}{2/B}{2/B}"), 0) {
		t.Error("B + three others paid {2/B}{2/B}{2/B}")
	}
}

// A plain {B} beside a {2/B} keeps the one black for itself: the hybrid is
// paid with the two others whatever order the cost lists them in.
func TestTwobridNeverTakesTheBlackAPlainSymbolNeeds(t *testing.T) {
	for _, cost := range []string{"{B}{2/B}", "{2/B}{B}"} {
		if !tbPool("B", "C", "C").CanPay(costFor(t, cost), 0) {
			t.Errorf("B + two colourless did not pay %s", cost)
		}
	}
}

func TestTwobridAlongsideGeneric(t *testing.T) {
	if !tbPool("B", "C", "C").CanPay(costFor(t, "{1}{2/B}"), 0) {
		t.Error("B + two colourless did not pay {1}{2/B}")
	}
	if tbPool("C", "C").CanPay(costFor(t, "{1}{2/B}"), 0) {
		t.Error("two colourless paid {1}{2/B}")
	}
}

func TestTwobridSpendTakesTheColourFirst(t *testing.T) {
	p := tbPool("C", "B", "C")
	if !p.SpendMana(costFor(t, "{2/B}"), 0) {
		t.Fatal("could not spend {2/B}")
	}
	if len(p) != 2 || p[0].Color != "C" || p[1].Color != "C" {
		t.Errorf("pool after paying {2/B} with a black available = %v, want the two colourless left", p)
	}
}

func TestTwobridMissingNamesWhatIsShort(t *testing.T) {
	missing := tbPool("C").MissingFor(costFor(t, "{2/B}"), 0, ManaSpendContext{})
	if len(missing) == 0 {
		t.Error("a short {2/B} reported nothing missing")
	}
}
