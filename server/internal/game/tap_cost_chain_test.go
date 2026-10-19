package game

import "testing"

// tap_cost_chain_test.go — #4177: the chained tap costs of a cast that
// owes a waterbend and a convoke (Hama's cast of a convoke spell).

// chainHooks installs the constructors a chained link is rebuilt from.
func chainHooks(t *testing.T) {
	t.Helper()
	pc, pw := ConvokeTapCost, WaterbendTapCost
	ConvokeTapCost = convokeForTest()
	WaterbendTapCost = waterbendForTest
	t.Cleanup(func() { ConvokeTapCost, WaterbendTapCost = pc, pw })
}

func TestWithAlsoChainsWithoutEditingTheSharedCost(t *testing.T) {
	chainHooks(t)
	wb, cv := waterbendForTest("{3}"), convokeForTest()
	wb.Label, cv.Label = "Waterbend {3}", "Convoke"
	got := wb.WithAlso(cv)
	if got == wb || wb.HasChain() {
		t.Fatal("WithAlso edited the catalog's cost")
	}
	if n := len(got.links()); n != 2 {
		t.Fatalf("%d links, want 2", n)
	}
	if len(got.WithAlso(convokeForTest()).Also) != 1 {
		t.Error("a second convoke was chained: two are redundant (CR 702.51b)")
	}
	if two := got.WithAlso(waterbendForTest("{X}")); len(two.links()) != 3 {
		t.Errorf("a second waterbend should be a third link, got %d", len(two.links()))
	}
}

func TestChainLabelNamesEveryCost(t *testing.T) {
	chainHooks(t)
	wb := waterbendForTest("{3}")
	wb.Label = "Waterbend {3}"
	cv := convokeForTest()
	cv.Label = "Convoke"
	ConvokeTapCost = cv
	if label := wb.WithAlso(cv).ChainLabel(); label != "Waterbend {3} + Convoke" {
		t.Errorf("label %q", label)
	}
}

func TestWithAlsoOfAnEmptyCostIsTheOther(t *testing.T) {
	cv := convokeForTest()
	if (*TapPermanentsCost)(nil).WithAlso(cv) != cv || cv.WithAlso(nil) != cv {
		t.Error("an empty side should leave the other cost as it was")
	}
}

func TestChainDemandsXWhenAnyLinkDoes(t *testing.T) {
	chainHooks(t)
	c := waterbendForTest("{3}").WithAlso(waterbendForTest("{X}"))
	if !c.DemandsX() {
		t.Error("a chain with a waterbend {X} link did not demand X")
	}
	if waterbendForTest("{3}").WithAlso(convokeForTest()).DemandsX() {
		t.Error("waterbend {3} and convoke demand no X")
	}
}

func TestChainKeepsAnInCostWaterbendInCost(t *testing.T) {
	chainHooks(t)
	wb := waterbendForTest("{3}")
	wb.InCost = true
	in := convokeForTest().WithAlso(wb)
	ls := in.links()
	if len(ls) != 2 || !ls[1].InCost || ls[1].Extra != "{3}" {
		t.Errorf("links %+v: the rebuilt waterbend lost its InCost or its {3}", ls)
	}
}
