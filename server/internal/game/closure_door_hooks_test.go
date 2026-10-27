package game

import (
	"sync"

	"github.com/google/uuid"
)

// closure_door_hooks_test.go — #4253's last part deleted the closure doors
// only tests still called (AddCounterThenForEffect, AmassForEffect,
// DestroyPermanentsThenForEffect, LoseLifeEachThenForEffect, ...). Their
// tests now ask through the keyed siblings, and hear the continuation
// through these test-only registered bodies: each looks up, by the word (or
// key) its data carries, a Go function the test handed it. A production
// body never captures anything; a test's hook may, because a test process
// keeps the map for its whole run, restore or no restore.

var (
	testHookMu sync.Mutex
	// testHooks are the hooks the tests registered, by key.
	testHooks = map[string]func(g *Game, o TailOutcome) error{}

	// testHookTail runs the hook its tail's Word names.
	testHookTail = RegisterTail("test/4253-hook-tail", func(g *Game, o TailOutcome) error {
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
func hookTail(fn func(g *Game, o TailOutcome) error) KeyedTail {
	return KeyedTail{Ref: testHookTail, Word: registerTestHook(fn)}
}

// registerTestHook files fn under a fresh key and returns the key.
func registerTestHook(fn func(g *Game, o TailOutcome) error) string {
	key := uuid.NewString()
	testHookMu.Lock()
	testHooks[key] = fn
	testHookMu.Unlock()
	return key
}

// tailFirstID is the first ID a tail was told, or uuid.Nil.
func tailFirstID(o TailOutcome) uuid.UUID {
	if len(o.IDs) == 0 {
		return uuid.Nil
	}
	return o.IDs[0]
}

var (
	// testRunHooks are the run hooks the tests registered, by ID.
	testRunHooks = map[uuid.UUID]func(g *Game, r RunDone) error{}

	// testRunHookThen runs the run hook its run's first carried ID names.
	testRunHookThen = RegisterPromptRunThen("test/4253-hook-run", func(g *Game, r RunDone) error {
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
func runHook(fn func(g *Game, r RunDone) error) RunData {
	id := uuid.New()
	testHookMu.Lock()
	testRunHooks[id] = fn
	testHookMu.Unlock()
	return RunData{Carry: []uuid.UUID{id}}
}

var (
	// testConfirmHooks are the confirm branches the tests registered, by
	// key: accept, then decline.
	testConfirmHooks = map[string][2]func(g *Game) error{}

	// testConfirmHook runs the branch its prompt's word names.
	testConfirmHook = RegisterConfirmThen("test/4253-hook-confirm", func(g *Game, r ConfirmAnswer) error {
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
func queueConfirmHooked(g *Game, p ConfirmPrompt, onAccept, onDecline func(g *Game) error) uuid.UUID {
	key := uuid.NewString()
	testHookMu.Lock()
	testConfirmHooks[key] = [2]func(g *Game) error{onAccept, onDecline}
	testHookMu.Unlock()
	p.Word = key
	id, err := g.QueueConfirmKeyedForEffect(p, testConfirmHook, nil)
	if err != nil {
		panic(err)
	}
	return id
}

// anyPlayerTestOffer is an any-player offer whose consequences are Go
// functions, for a test: queueAnyPlayerHooked files them under an ID the
// offer carries and queues the keyed offer
// (QueueAnyPlayerPaysKeyedForEffect).
type anyPlayerTestOffer struct {
	Source, Controller         uuid.UUID
	Cost                       string
	Action                     *PayAction
	PerPayer                   func(g *Game, payer uuid.UUID) (PayAction, string)
	AcceptLabel, DeclineLabel  string
	Question                   string
	Opponents, HelpsController bool
	OnPaid                     func(g *Game, payer uuid.UUID) error
	OnNobody                   func(g *Game) error
	Guards                     uuid.UUID
}

var (
	// testAnyPlayerOffers are the offers the tests queued, by the ID each
	// carries.
	testAnyPlayerOffers = map[uuid.UUID]anyPlayerTestOffer{}

	testAnyPlayerHook = RegisterAnyPlayerPays("test/4253-game-hook-anyplayer",
		anyPlayerHookPaid, anyPlayerHookNobody, nil)
	testAnyPlayerHookPerPayer = RegisterAnyPlayerPays("test/4253-game-hook-anyplayer-per-payer",
		anyPlayerHookPaid, anyPlayerHookNobody,
		func(g *Game, payer uuid.UUID, carry []uuid.UUID) (PayAction, string) {
			return anyPlayerOfferOf(carry).PerPayer(g, payer)
		})
)

// anyPlayerOfferOf is the offer a carry names.
func anyPlayerOfferOf(carry []uuid.UUID) anyPlayerTestOffer {
	if len(carry) == 0 {
		return anyPlayerTestOffer{}
	}
	testHookMu.Lock()
	defer testHookMu.Unlock()
	return testAnyPlayerOffers[carry[0]]
}

func anyPlayerHookPaid(g *Game, r AnyPlayerAnswer) error {
	if o := anyPlayerOfferOf(r.Carry); o.OnPaid != nil {
		return o.OnPaid(g, r.Payer)
	}
	return nil
}

func anyPlayerHookNobody(g *Game, r AnyPlayerAnswer) error {
	if o := anyPlayerOfferOf(r.Carry); o.OnNobody != nil {
		return o.OnNobody(g)
	}
	return nil
}

// queueAnyPlayerHooked queues p as a keyed any-player offer.
func queueAnyPlayerHooked(g *Game, p anyPlayerTestOffer) error {
	id := uuid.New()
	testHookMu.Lock()
	testAnyPlayerOffers[id] = p
	testHookMu.Unlock()
	then := testAnyPlayerHook
	if p.PerPayer != nil {
		then = testAnyPlayerHookPerPayer
	}
	return g.QueueAnyPlayerPaysKeyedForEffect(AnyPlayerPaysKeyed{
		Source: p.Source, Controller: p.Controller, Cost: p.Cost, Action: p.Action,
		AcceptLabel: p.AcceptLabel, DeclineLabel: p.DeclineLabel, Question: p.Question,
		Opponents: p.Opponents, HelpsController: p.HelpsController, Guards: p.Guards,
		Then: then, Carry: []uuid.UUID{id},
	})
}

// playersDiscardHooked is a keyed discard run (PlayersDiscardThenKeyedForEffect)
// whose continuation is fn, handed what each seat really discarded.
func playersDiscardHooked(g *Game, players []uuid.UUID, p DiscardPrompt, fn func(g *Game, discarded PromptedDiscards) error) error {
	data := runHook(func(g *Game, r RunDone) error { return fn(g, r.Discards()) })
	return g.PlayersDiscardThenKeyedForEffect(players, p, testRunHookThen, data.Carry)
}

// eachPlayerDiscardsHooked is the APNAP fan-out of playersDiscardHooked.
func eachPlayerDiscardsHooked(g *Game, except uuid.UUID, p DiscardPrompt, fn func(g *Game, discarded PromptedDiscards) error) error {
	data := runHook(func(g *Game, r RunDone) error { return fn(g, r.Discards()) })
	return g.EachPlayerDiscardsThenKeyedForEffect(except, p, testRunHookThen, data.Carry)
}

// eachPlayerSacrificesHooked is a keyed sacrifice run
// (EachPlayerSacrificesThenKeyedForEffect) whose continuation is fn, handed
// what each seat really sacrificed.
func eachPlayerSacrificesHooked(g *Game, source, except uuid.UUID, spec *TargetSpec, reason string, fn func(g *Game, sacrificed PromptedSacrifices) error) error {
	data := runHook(func(g *Game, r RunDone) error { return fn(g, r.Sacrifices()) })
	return g.EachPlayerSacrificesThenKeyedForEffect(source, except, spec, reason, testRunHookThen, data)
}
