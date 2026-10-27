package actions

import (
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_door_hooks_test.go — #4253's last part deleted the engine closure
// doors only tests still called (QueueAnyPlayerPaysForEffect). The tests here that used them
// now ask through the keyed siblings, and hear the continuation through
// these test-only registered bodies: each looks up, by the word or ID its
// data carries, a Go function the test handed it. A production body never
// captures anything; a test's hook may, because a test process keeps the
// map for its whole run.

// testHookMu guards the hook maps below.
var testHookMu sync.Mutex

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

	testAnyPlayerHook = game.RegisterAnyPlayerPays("test/4253-actions-hook-anyplayer",
		anyPlayerHookPaid, anyPlayerHookNobody, nil)
	testAnyPlayerHookPerPayer = game.RegisterAnyPlayerPays("test/4253-actions-hook-anyplayer-per-payer",
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
