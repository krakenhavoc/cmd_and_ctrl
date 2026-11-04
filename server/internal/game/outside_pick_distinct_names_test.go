package game

import "testing"

// outside_pick_distinct_names_test.go — #4947: a pick registered "with
// different names" refuses a repeated name through its set rule, keeps
// any rule of its own, and is the only kind that says so.

var testDistinctPick = RegisterOutsidePickDistinctNames("test/4947-distinct", nil, nil,
	func(*Game, CardsChosen) error { return nil })

var testDistinctPairPick = RegisterOutsidePickDistinctNames("test/4947-distinct-pair", nil,
	func(picked []Card) bool { return len(picked) == 2 },
	func(*Game, CardsChosen) error { return nil })

var testRepeatPick = RegisterOutsidePickRule("test/4947-repeat", nil,
	func(picked []Card) bool { return len(picked) <= 4 },
	func(*Game, CardsChosen) error { return nil })

func TestADistinctNamesPickRefusesARepeatedName(t *testing.T) {
	named := func(names ...string) []Card {
		out := make([]Card, len(names))
		for i, n := range names {
			out[i].Name = n
		}
		return out
	}
	rule := lookupChooseCardsRule(testDistinctPick.Key())
	if rule == nil {
		t.Fatal("a distinct-names pick has no set rule")
	}
	if !rule(named("Leonardo", "Raphael")) {
		t.Error("two different names are refused")
	}
	if rule(named("Leonardo", " leonardo ")) {
		t.Error("the same name twice, in another case and spacing, is allowed")
	}

	pair := lookupChooseCardsRule(testDistinctPairPick.Key())
	if !pair(named("A", "B")) || pair(named("A", "A")) || pair(named("A", "B", "C")) {
		t.Error("the pick's own rule is not kept alongside the names check")
	}

	if !OutsidePickNamesDiffer(testDistinctPick.Key()) || !OutsidePickNamesDiffer(testDistinctPairPick.Key()) {
		t.Error("a distinct-names pick does not say so")
	}
	if OutsidePickNamesDiffer(testRepeatPick.Key()) || OutsidePickNamesDiffer(outsidePickPlain.Key()) ||
		OutsidePickNamesDiffer("no/such-pick") {
		t.Error("a pick whose names may repeat says they differ")
	}
}
