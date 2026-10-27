package game

import (
	"sync"

	"github.com/google/uuid"
)

// search_outside_hooks_test.go — #4361 S4 deleted the closure forms of a
// library search (SearchLibrarySpec.Then and Validate) and of an
// outside-the-game pick (OutsidePickPrompt.Match, Validate and Then), which
// only tests still set. Their tests now queue the keyed forms and hear the
// continuation through these test-only registered bodies, the way
// closure_door_hooks_test.go does for the doors #4253 deleted.

// searchHook is a search's ThenTail that runs fn, handed the found cards.
// A nil fn is the zero tail, "nothing follows".
func searchHook(fn func(g *Game, found []uuid.UUID) error) KeyedTail {
	if fn == nil {
		return KeyedTail{}
	}
	return hookTail(func(g *Game, o TailOutcome) error { return fn(g, o.IDs) })
}

var (
	testOutsideHookMu sync.Mutex
	// testOutsideHooks are the outside-pick continuations the tests
	// registered, by the ID the pick carries.
	testOutsideHooks = map[uuid.UUID]func(g *Game, picked []uuid.UUID) error{}

	// testOutsidePick filters by the word the pick carries
	// (testOutsideMatch) and runs the hook its first carried ID names.
	testOutsidePick = RegisterOutsidePick("test/4361-outside-hook", testOutsideMatch, testOutsideHookThen)
	// testOutsidePickTwoNames is testOutsidePick with Extrapolate the
	// Impossible's set rule: exactly two cards with different names.
	testOutsidePickTwoNames = RegisterOutsidePickRule("test/4361-outside-hook-two-names", testOutsideMatch,
		func(picked []Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
		testOutsideHookThen)
)

// testOutsideMatch is a test pick's filter, by word: "instant", "land",
// "sorcery", or every card.
func testOutsideMatch(c Card, word string) bool {
	switch word {
	case "instant":
		return c.IsInstant()
	case "land":
		return c.IsLand()
	case "sorcery":
		return c.IsSorcery()
	}
	return true
}

func testOutsideHookThen(g *Game, r CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	testOutsideHookMu.Lock()
	fn := testOutsideHooks[r.Carry[0]]
	testOutsideHookMu.Unlock()
	if fn == nil {
		return nil
	}
	return fn(g, r.Picked)
}

// outsideHooked is p asked through a test pick: filtered by `match` (a
// testOutsideMatch word), with `then` (nil for none) handed the picks.
func outsideHooked(p OutsidePickPrompt, match string, then func(g *Game, picked []uuid.UUID) error) OutsidePickPrompt {
	return outsideHookedOn(testOutsidePick, p, match, then)
}

// outsideHookedTwoNames is outsideHooked with the two-names set rule.
func outsideHookedTwoNames(p OutsidePickPrompt, then func(g *Game, picked []uuid.UUID) error) OutsidePickPrompt {
	return outsideHookedOn(testOutsidePickTwoNames, p, "", then)
}

func outsideHookedOn(pick OutsidePick, p OutsidePickPrompt, match string, then func(g *Game, picked []uuid.UUID) error) OutsidePickPrompt {
	id := uuid.New()
	if then != nil {
		testOutsideHookMu.Lock()
		testOutsideHooks[id] = then
		testOutsideHookMu.Unlock()
	}
	p.Keyed, p.MatchArg, p.Carry = pick, match, []uuid.UUID{id}
	return p
}
