package legal_test

import (
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_door_hooks_test.go — #4253's last part deleted the engine closure
// doors only tests still called (QueueConfirmForEffect, PermanentsPickedThenForEffect, QueueAnyPlayerPaysForEffect). The tests here that used them
// now ask through the keyed siblings, and hear the continuation through
// these test-only registered bodies: each looks up, by the word or ID its
// data carries, a Go function the test handed it. A production body never
// captures anything; a test's hook may, because a test process keeps the
// map for its whole run.

// testHookMu guards the hook maps below.
var testHookMu sync.Mutex

var (
	// testRunHooks are the run hooks, by the ID the run carries first.
	testRunHooks = map[uuid.UUID]func(g *game.Game, r game.RunDone) error{}

	// testRunHookThen runs the run hook its run's first carried ID names.
	testRunHookThen = game.RegisterPromptRunThen("test/4253-legal-hook-run", func(g *game.Game, r game.RunDone) error {
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
	// testConfirmHooks are the confirm branches, accept then decline.
	testConfirmHooks = map[string][2]func(g *game.Game) error{}

	// testConfirmHook runs the branch its prompt's word names.
	testConfirmHook = game.RegisterConfirmThen("test/4253-legal-hook-confirm", func(g *game.Game, r game.ConfirmAnswer) error {
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

// anyPlayerTestOffer is an any-player offer whose consequences are Go
// functions, for a test: queueAnyPlayerHooked files them under an ID the
// offer carries and queues the keyed offer
// (QueueAnyPlayerPaysKeyedForEffect).
type anyPlayerTestOffer struct {
	Source, Controller         uuid.UUID
	Cost                       string
	Action                     *game.PayAction
	PerPayer                   func(g *game.Game, payer uuid.UUID) (game.PayAction, string)
	AcceptLabel, DeclineLabel  string
	Question                   string
	Opponents, HelpsController bool
	OnPaid                     func(g *game.Game, payer uuid.UUID) error
	OnNobody                   func(g *game.Game) error
	Guards                     uuid.UUID
}

var (
	// testAnyPlayerOffers are the offers the tests queued, by the ID each
	// carries.
	testAnyPlayerOffers = map[uuid.UUID]anyPlayerTestOffer{}

	testAnyPlayerHook = game.RegisterAnyPlayerPays("test/4253-legal-hook-anyplayer",
		anyPlayerHookPaid, anyPlayerHookNobody, nil)
	testAnyPlayerHookPerPayer = game.RegisterAnyPlayerPays("test/4253-legal-hook-anyplayer-per-payer",
		anyPlayerHookPaid, anyPlayerHookNobody,
		func(g *game.Game, payer uuid.UUID, carry []uuid.UUID) (game.PayAction, string) {
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

func anyPlayerHookPaid(g *game.Game, r game.AnyPlayerAnswer) error {
	if o := anyPlayerOfferOf(r.Carry); o.OnPaid != nil {
		return o.OnPaid(g, r.Payer)
	}
	return nil
}

func anyPlayerHookNobody(g *game.Game, r game.AnyPlayerAnswer) error {
	if o := anyPlayerOfferOf(r.Carry); o.OnNobody != nil {
		return o.OnNobody(g)
	}
	return nil
}

// queueAnyPlayerHooked queues p as a keyed any-player offer.
func queueAnyPlayerHooked(g *game.Game, p anyPlayerTestOffer) error {
	id := uuid.New()
	testHookMu.Lock()
	testAnyPlayerOffers[id] = p
	testHookMu.Unlock()
	then := testAnyPlayerHook
	if p.PerPayer != nil {
		then = testAnyPlayerHookPerPayer
	}
	return g.QueueAnyPlayerPaysKeyedForEffect(game.AnyPlayerPaysKeyed{
		Source: p.Source, Controller: p.Controller, Cost: p.Cost, Action: p.Action,
		AcceptLabel: p.AcceptLabel, DeclineLabel: p.DeclineLabel, Question: p.Question,
		Opponents: p.Opponents, HelpsController: p.HelpsController, Guards: p.Guards,
		Then: then, Carry: []uuid.UUID{id},
	})
}

// testNoopDecline is a keyed pay-unless decline that does nothing: a
// guarded pay-unless whose "or else" is not a counter (Divert's shape),
// where the test reads only the moves (#4361 deleted the OnDecline
// closure it used to pass).
var testNoopDecline = game.RegisterPayUnlessElse("test/4361-legal-noop-decline",
	func(*game.Game, game.PayUnlessAnswer) error { return nil })
