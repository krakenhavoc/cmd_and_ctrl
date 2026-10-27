package effects

import (
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_door_hooks_test.go — #4253's last part deleted the engine closure
// doors only tests still called (DestroyPermanentsThenForEffect,
// PutOntoBattlefieldTogetherThenForEffect, SurveilThenForEffect, ...). The
// tests here that used them now ask through the keyed siblings, and hear the
// continuation through this test-only registered tail: it looks up, by the
// word its data carries, a Go function the test handed it. A production
// tail never captures anything; a test's hook may, because a test process
// keeps the map for its whole run, restore or no restore.

var (
	testHookMu sync.Mutex
	// testHooks are the hooks the tests registered, by key.
	testHooks = map[string]func(g *game.Game, o game.TailOutcome) error{}

	// testHookTail runs the hook its tail's Word names.
	testHookTail = game.RegisterTail("test/4253-effects-hook-tail", func(g *game.Game, o game.TailOutcome) error {
		testHookMu.Lock()
		fn := testHooks[o.Tail.Word]
		testHookMu.Unlock()
		if fn == nil {
			return nil
		}
		return fn(g, o)
	})
)

// hookTail is a keyed tail that runs fn, told what the tail is told.
func hookTail(fn func(g *game.Game, o game.TailOutcome) error) game.KeyedTail {
	key := uuid.NewString()
	testHookMu.Lock()
	testHooks[key] = fn
	testHookMu.Unlock()
	return game.KeyedTail{Ref: testHookTail, Word: key}
}

// tailFirstID is the first ID a tail was told, or uuid.Nil.
func tailFirstID(o game.TailOutcome) uuid.UUID {
	if len(o.IDs) == 0 {
		return uuid.Nil
	}
	return o.IDs[0]
}
