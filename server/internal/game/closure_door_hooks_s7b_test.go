package game

import "github.com/google/uuid"

// closure_door_hooks_s7b_test.go — #4361 S7b deleted the closure forms of
// the card-set pick (ChooseCardsPrompt.Then and .Validate), the discard
// prompt's (DiscardPrompt.Then and .Validate), the ordered collector's
// (ChoosePermanentsInOrderThenForEffect) and the per-object any-player
// offer's (AnyPlayerPaysForEach.OnUnsaved). Their tests ask through the
// keyed forms and hear the answer through these hooks: each registers, under
// a fresh test key, a continuation that runs the Go function the test
// handed it. A production continuation never captures anything; a test's
// may, because a test process keeps the registry for its whole run,
// restore or no restore.

// chooseCardsHooked is p with fn as its keyed continuation, handed the
// picks (nil when nobody chose), and validate, when non-nil, as the set
// rule registered with it. A Zone the test set is the pick's live-zone
// re-check, as it was for the closure pick.
func chooseCardsHooked(p ChooseCardsPrompt, fn func(g *Game, picked []uuid.UUID) error, validate func([]Card) bool) ChooseCardsPrompt {
	p.ThenKey = registerChooseCardsHook(fn, validate)
	if p.LiveZone == "" {
		p.LiveZone = p.Zone
	}
	return p
}

// registerChooseCardsHook registers fn (and validate, when non-nil) as a
// keyed choose_cards continuation under a fresh test key.
func registerChooseCardsHook(fn func(g *Game, picked []uuid.UUID) error, validate func([]Card) bool) ChooseCardsThen {
	key := "test/4361-s7b-pick-" + uuid.NewString()
	body := func(g *Game, r CardsChosen) error {
		if fn == nil {
			return nil
		}
		return fn(g, r.Picked)
	}
	if validate != nil {
		return RegisterChooseCardsThenRule(key, validate, body)
	}
	return RegisterChooseCardsThen(key, body)
}

// discardThenHooked queues p as a keyed discard-then
// (PlayerDiscardsThenKeyedForEffect) whose rest is fn, handed the seat and
// what it really discarded, with validate, when non-nil, as its set rule.
func discardThenHooked(g *Game, p DiscardPrompt, fn func(g *Game, seat uuid.UUID, discarded []uuid.UUID) error, validate func([]Card) bool) error {
	key := "test/4361-s7b-discard-" + uuid.NewString()
	body := func(g *Game, d DiscardsDone) error {
		if fn == nil {
			return nil
		}
		return fn(g, d.Player, d.Discarded.By(d.Player))
	}
	return g.PlayerDiscardsThenKeyedForEffect(p, RegisterDiscardThenRule(key, validate, body), nil)
}

// queueEffectDiscardThen is queueEffectDiscard for a prompt whose rest is fn
// (and whose set rule is validate, when non-nil): a keyed discard-then
// (discardThenHooked). It returns the queued prompt's ID, or uuid.Nil when
// none went up.
func queueEffectDiscardThen(g *Game, p DiscardPrompt, fn func(g *Game, seat uuid.UUID, discarded []uuid.UUID) error, validate func([]Card) bool) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		before := len(g.PendingChoices)
		if err := discardThenHooked(g, p, fn, validate); err != nil {
			panic(err)
		}
		if len(g.PendingChoices) > before {
			id = g.PendingChoices[len(g.PendingChoices)-1].ID
		}
	})
	return id
}

// discardLegHook is a keyed run's per-seat tail (RegisterPromptRunThenLeg)
// that runs fn, handed the seat and what it really discarded, and a run
// continuation that does nothing; queue the run with
// EachPlayerDiscardsThenKeyedForEffect or PlayersDiscardThenKeyedForEffect.
func discardLegHook(fn func(g *Game, seat uuid.UUID, discarded []uuid.UUID) error) PromptRunThen {
	key := "test/4361-s7b-leg-" + uuid.NewString()
	return RegisterPromptRunThenLeg(key,
		func(g *Game, l RunLeg) error { return fn(g, l.Seat, l.Landed) },
		func(*Game, RunDone) error { return nil })
}

// anyPlayerEachHook is a per-object any-player offer's registered
// consequence (RegisterAnyPlayerPaysEach) that runs fn on the objects
// nobody paid for.
func anyPlayerEachHook(fn func(g *Game, unsaved []uuid.UUID) error) AnyPlayerEachThen {
	return RegisterAnyPlayerPaysEach("test/4361-s7b-each-"+uuid.NewString(), func(g *Game, r AnyPlayerEachAnswer) error {
		return fn(g, r.Unsaved)
	})
}

// outsidePickHooked is p as a keyed outside pick (RegisterOutsidePickRule)
// whose continuation is fn, handed the picks, and whose set rule is
// validate, when non-nil: a pick over a sideboard is keyed since #4361 S7b.
func outsidePickHooked(p OutsidePickPrompt, fn func(g *Game, picked []uuid.UUID) error, validate func([]Card) bool) OutsidePickPrompt {
	p.Keyed = RegisterOutsidePickRule("test/4361-s7b-outside-"+uuid.NewString(), nil, validate, func(g *Game, r CardsChosen) error {
		if fn == nil {
			return nil
		}
		return fn(g, r.Picked)
	})
	return p
}
