package aiseat_test

import (
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_door_hooks_test.go — #4253's last part deleted the engine closure
// doors only tests still called (LookAtTopThenForEffect, QueueConfirmForEffect, ManifestDreadThenForEffect, PermanentsPickedThenForEffect, RingTemptsForEffect). The tests here that used them
// now ask through the keyed siblings, and hear the continuation through
// these test-only registered bodies: each looks up, by the word or ID its
// data carries, a Go function the test handed it. A production body never
// captures anything; a test's hook may, because a test process keeps the
// map for its whole run.

// testHookMu guards the hook maps below.
var testHookMu sync.Mutex

var (
	// testHooks are the tail hooks the tests registered, by key.
	testHooks = map[string]func(g *game.Game, o game.TailOutcome) error{}

	// testHookTail runs the hook its tail's Word names.
	testHookTail = game.RegisterTail("test/4253-aiseat-hook-tail", func(g *game.Game, o game.TailOutcome) error {
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

var (
	// testRunHooks are the run hooks, by the ID the run carries first.
	testRunHooks = map[uuid.UUID]func(g *game.Game, r game.RunDone) error{}

	// testRunHookThen runs the run hook its run's first carried ID names.
	testRunHookThen = game.RegisterPromptRunThen("test/4253-aiseat-hook-run", func(g *game.Game, r game.RunDone) error {
		if len(r.Carry) == 0 {
			return nil
		}
		testHookMu.Lock()
		fn := testRunHooks[r.Carry[0]]
		testHookMu.Unlock()
		if fn == nil {
			return nil
		}
		return fn(g, r)
	})
)

// runHook is the RunData of a run on testRunHookThen that runs fn, handed
// the run's answer.
func runHook(fn func(g *game.Game, r game.RunDone) error) game.RunData {
	id := uuid.New()
	testHookMu.Lock()
	testRunHooks[id] = fn
	testHookMu.Unlock()
	return game.RunData{Carry: []uuid.UUID{id}}
}

var (
	// testRevealPickHooks are the reveal-pick hooks (#4361: the closure
	// form of RevealPickThenForEffect is gone), by the ID the pick carries.
	testRevealPickHooks = map[uuid.UUID]func(g *game.Game, r game.CardsChosen) error{}

	// testRevealPickHookThen runs the hook its pick's first carried ID names.
	testRevealPickHookThen = game.RegisterChooseCardsThen("test/4361-aiseat-hook-reveal-pick", func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) == 0 {
			return nil
		}
		testHookMu.Lock()
		fn := testRevealPickHooks[r.Carry[0]]
		testHookMu.Unlock()
		if fn == nil {
			return nil
		}
		return fn(g, r)
	})
)

// revealPickHooked is p with fn as its keyed continuation.
func revealPickHooked(p game.RevealPickPrompt, fn func(g *game.Game, r game.CardsChosen) error) game.RevealPickPrompt {
	id := uuid.New()
	testHookMu.Lock()
	testRevealPickHooks[id] = fn
	testHookMu.Unlock()
	p.ThenKey, p.Carry = testRevealPickHookThen, []uuid.UUID{id}
	return p
}

var (
	// testConfirmHooks are the confirm branches, accept then decline.
	testConfirmHooks = map[string][2]func(g *game.Game) error{}

	// testConfirmHook runs the branch its prompt's word names.
	testConfirmHook = game.RegisterConfirmThen("test/4253-aiseat-hook-confirm", func(g *game.Game, r game.ConfirmAnswer) error {
		testHookMu.Lock()
		fns := testConfirmHooks[r.Word]
		testHookMu.Unlock()
		fn := fns[1]
		if r.Accepted {
			fn = fns[0]
		}
		if fn == nil {
			return nil
		}
		return fn(g)
	})
)

// queueConfirmHooked queues p as a keyed confirm whose two branches are
// onAccept and onDecline (either may be nil), and returns its ID.
func queueConfirmHooked(g *game.Game, p game.ConfirmPrompt, onAccept, onDecline func(g *game.Game) error) uuid.UUID {
	key := uuid.NewString()
	testHookMu.Lock()
	testConfirmHooks[key] = [2]func(g *game.Game) error{onAccept, onDecline}
	testHookMu.Unlock()
	p.Word = key
	id, err := g.QueueConfirmKeyedForEffect(p, testConfirmHook, nil)
	if err != nil {
		panic(err)
	}
	return id
}

var (
	// testPayUnlessHooks are the pay-unless declines the tests registered,
	// by the ID the prompt carries first (#4361).
	testPayUnlessHooks = map[uuid.UUID]func(g *game.Game) error{}

	// testPayUnlessHook runs the decline its prompt's first carried ID
	// names: the stand-in for the OnDecline closure #4361 deleted.
	testPayUnlessHook = game.RegisterPayUnlessElse("test/4361-aiseat-hook-payunless", func(g *game.Game, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		testHookMu.Lock()
		fn := testPayUnlessHooks[r.Carry[0]]
		testHookMu.Unlock()
		if fn == nil {
			return nil
		}
		return fn(g)
	})
)

// payUnlessHookCarry files onDecline for a prompt keyed testPayUnlessHook
// and returns the Carry that names it.
func payUnlessHookCarry(onDecline func(g *game.Game) error) []uuid.UUID {
	id := uuid.New()
	testHookMu.Lock()
	testPayUnlessHooks[id] = onDecline
	testHookMu.Unlock()
	return []uuid.UUID{id}
}
