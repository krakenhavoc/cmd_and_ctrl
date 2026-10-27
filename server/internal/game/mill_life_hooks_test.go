package game

import "github.com/google/uuid"

// mill_life_hooks_test.go — #4361 S9 deleted the closure mill and life doors
// (MillToZoneThenForEffect, MillToZoneThenMovedForEffect,
// ChangePlayerLifeThenForEffect). Their tests ask through the keyed doors and
// hear the continuation through hookTail (closure_door_hooks_test.go); an
// "until" clause is a test-only registered clause that looks up, by the
// number its MillUntil carries, a Go predicate the test handed it.

var (
	// testMillUntils are the stop clauses the tests registered, by number.
	testMillUntils   = map[int]func([]Card) bool{}
	testMillUntilSeq int

	// testHookMillUntil asks the clause its MillUntil's N names.
	testHookMillUntil = RegisterMillUntil("test/4361-hook-mill-until", func(landed []Card, n int) bool {
		testHookMu.Lock()
		fn := testMillUntils[n]
		testHookMu.Unlock()
		return fn != nil && fn(landed)
	})
)

// hookMillUntil is a registered stop clause that asks fn.
func hookMillUntil(fn func([]Card) bool) MillUntil {
	testHookMu.Lock()
	testMillUntilSeq++
	n := testMillUntilSeq
	testMillUntils[n] = fn
	testHookMu.Unlock()
	return MillUntil{Ref: testHookMillUntil, N: n}
}

// lifeHook is a keyed life tail that runs fn, told the amount that moved.
func lifeHook(fn func(g *Game, applied int) error) KeyedTail {
	return hookTail(func(g *Game, o TailOutcome) error { return fn(g, o.Amount) })
}

// millHook is a keyed mill tail that runs fn, told the cards.
func millHook(fn func(g *Game, milled []uuid.UUID) error) KeyedTail {
	return hookTail(func(g *Game, o TailOutcome) error { return fn(g, o.IDs) })
}
