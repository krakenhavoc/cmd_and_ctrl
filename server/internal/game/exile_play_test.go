package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_play_test.go — S21 sub-PR 6: impulse exile. Exile is a
// shared, public zone, so the interesting cases are all about who
// may touch what, and for how long.

// seedLibraryTop puts a card on top of p's library and returns it.
func seedLibraryTop(p *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{R}"
	p.Library.PushTop(c)
	return c.InstanceID
}

// impulseExile runs the primitive: exiles the top of victim's
// library and grants thief permission for this turn.
func impulseExile(t *testing.T, g *Game, victim, thief *Player, perm CastPermission) uuid.UUID {
	t.Helper()
	var ids []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		ids, err = g.ExileTopWithPermissionForEffect(victim.ID, thief.ID, 1, perm)
	})
	if err != nil {
		t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("exiled %d cards, want 1", len(ids))
	}
	return ids[0]
}

func toMainPhase(t *testing.T, g *Game) {
	t.Helper()
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
}

// beginALaterTurnFor makes it look as though `p` has begun another
// turn since — Player.TurnsBegun is the seat-turn counter every
// ADR 0063 duration is stamped against, and a real rotation bumps it
// in noteTurnBegunLocked. Tests that only need "time has passed for
// this seat" say so here rather than driving four turns of play.
func beginALaterTurnFor(g *Game, p *Player) {
	g.WithWriteLock(func() { p.TurnsBegun++ })
}

// TestCastPermissionActive pins the one liveness test (#945): the
// permission names you, its CR 702.185a floor has been reached, and
// its CR 611.2 duration has not run out.
func TestCastPermissionActive(t *testing.T) {
	g := newActiveGame(t)
	me, you := g.Seats[0], g.Seats[1]
	var thisTurn, myNextTurn Duration
	g.WithWriteLock(func() {
		thisTurn = g.UntilEndOfTurnDuration()
		myNextTurn = g.UntilEndOfYourNextTurnDuration(me.ID)
	})
	cases := []struct {
		name  string
		perm  CastPermission
		who   uuid.UUID
		later bool // a later turn has begun for the holder
		want  bool
	}{
		{"zero value grants nothing", CastPermission{}, me.ID, false, false},
		{"holder, this turn", CastPermission{Player: me.ID, Duration: thisTurn}, me.ID, false, true},
		{"holder, a turn later", CastPermission{Player: me.ID, Duration: thisTurn}, me.ID, true, false},
		{"holder, until end of next turn", CastPermission{Player: me.ID, Duration: myNextTurn}, me.ID, true, true},
		{"while in zone outlasts every turn", CastPermission{Player: me.ID, Duration: WhileInZoneDuration()}, me.ID, true, true},
		{"somebody else", CastPermission{Player: me.ID, Duration: thisTurn}, you.ID, false, false},
		{"the floor is not reached yet", CastPermission{
			Player: me.ID, Duration: WhileInZoneDuration(), NotBeforeSeq: g.Turn.Seq + 1,
		}, me.ID, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			probe := g
			if tc.later {
				probe = g.Clone()
				beginALaterTurnFor(probe, probe.Seats[0])
			}
			perm := tc.perm
			var got bool
			probe.ReadSnapshot(func() { got = probe.CastPermissionActiveForEffect(&perm, tc.who) })
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExileTopWithPermissionMovesAndStamps(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	loot := seedLibraryTop(opp, "Stolen Bolt", "Instant")

	id := impulseExile(t, g, opp, me, CastPermission{CastOnly: true})
	if id != loot {
		t.Fatalf("exiled %v, want the top card %v", id, loot)
	}
	if !g.Exile.Contains(loot) {
		t.Fatalf("card never reached exile")
	}
	got := CastPermission{}
	if perm := g.CastPermissionOnCardByIDForEffect(loot); perm != nil {
		got = *perm
	}
	if got.Player != me.ID {
		t.Errorf("permission holder = %v, want the thief %v", got.Player, me.ID)
	}
	if got.Duration.Kind != UntilEndOfTurn || got.Duration.Player != me.ID {
		t.Errorf("duration = %v (player %v), want until end of this turn for %v",
			got.Duration.Kind, got.Duration.Player, me.ID)
	}
	if !got.CastOnly {
		t.Errorf("CastOnly should have ridden through")
	}
	// An empty library is not an error — you exile what's there.
	opp.Library.Cards = nil
	g.WithWriteLock(func() {
		ids, err := g.ExileTopWithPermissionForEffect(opp.ID, me.ID, 2, CastPermission{})
		if err != nil || len(ids) != 0 {
			t.Errorf("empty library: ids %v err %v, want none and no error", ids, err)
		}
	})
}

func TestCastFromExileNeedsALivePermission(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil

	// No permission at all.
	bare := seedLibraryTop(opp, "Bare", "Instant")
	g.WithWriteLock(func() { _, _ = MoveCard(opp.Library, g.Exile, bare) })
	if err := g.CastSpell(me.ID, bare, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Errorf("no permission: %v, want ErrNoPlayPermission", err)
	}

	// Granted to somebody else.
	theirs := seedLibraryTop(opp, "Theirs", "Instant")
	g.WithWriteLock(func() {
		_, _ = g.ExileTopWithPermissionForEffect(opp.ID, opp.ID, 1, CastPermission{})
	})
	if err := g.CastSpell(me.ID, theirs, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Errorf("someone else's grant: %v, want ErrNoPlayPermission", err)
	}

	for _, id := range []uuid.UUID{bare, theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("a rejected cast moved a card out of exile")
		}
	}
}

// "Until end of turn" means exactly that: the card is still sitting
// in exile next turn, and still unplayable.
func TestExilePermissionLapsesWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	seedLibraryTop(opp, "Stolen Bolt", "Instant")
	loot := impulseExile(t, g, opp, me, CastPermission{})

	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("this turn it should be castable: %v", err)
	}
	// Put it back and roll the turn over.
	g.WithWriteLock(func() {
		_, _ = MoveCard(g.Stack, g.Exile, loot)
		g.GrantCastPermissionOverCardForEffect(loot, CastPermission{Player: me.ID, Duration: g.UntilEndOfTurnDuration()})
		delete(g.StackMeta, loot)
		g.Turn.Seq++
		// #945: the window is a seat-turn, not a round, so this is
		// what makes it a LATER turn for the grant's holder.
		me.TurnsBegun++
	})
	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Errorf("next turn: %v, want ErrNoPlayPermission", err)
	}
}

func TestCastFromExileSpendsThePermission(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	seedLibraryTop(opp, "Stolen Bolt", "Instant")
	loot := impulseExile(t, g, opp, me, CastPermission{CastOnly: true})

	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if !g.Stack.Contains(loot) {
		t.Fatalf("card is not on the stack")
	}
	item := g.StackMeta[loot]
	if item == nil || item.Controller != me.ID {
		t.Errorf("the thief should control the spell, not its owner")
	}
	// The grant is spent, so re-exiling this card later can't
	// resurrect it.
	if perm := g.CastPermissionOnCardByIDForEffect(loot); perm.Granted() {
		t.Errorf("permission survived the cast")
	}
}

// "You may CAST that card" strands a land; "you may PLAY those
// cards" does not (CR 305.1 — playing a land is not casting).
func TestCastOnlyStrandsALandButPlayDoesNot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil

	seedLibraryTop(opp, "Stolen Island", "Basic Land — Island")
	stranded := impulseExile(t, g, opp, me, CastPermission{CastOnly: true})
	if err := g.CastSpell(me.ID, stranded, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Fatalf("cast-only land: %v, want ErrNoPlayPermission", err)
	}
	if !g.Exile.Contains(stranded) {
		t.Errorf("a rejected land left exile")
	}

	seedLibraryTop(opp, "Playable Island", "Basic Land — Island")
	playable := impulseExile(t, g, opp, me, CastPermission{})
	if err := g.CastSpell(me.ID, playable, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play-permission land: %v", err)
	}
	if !g.Battlefield.Contains(playable) {
		t.Fatalf("the land never reached the battlefield")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == playable {
			if c.Controller != me.ID {
				t.Errorf("controller = %v, want the thief", c.Controller)
			}
			if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); perm.Granted() {
				t.Errorf("permission survived the land drop")
			}
		}
	}
}

func TestCleanupClearsExpiredExilePermissions(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	seedLibraryTop(opp, "This Turn", "Instant")
	thisTurn := impulseExile(t, g, opp, me, CastPermission{})
	seedLibraryTop(opp, "Next Turn", "Instant")
	var nextTurnWindow Duration
	g.WithWriteLock(func() { nextTurnWindow = g.UntilEndOfYourNextTurnDuration(me.ID) })
	nextTurn := impulseExile(t, g, opp, me, CastPermission{Duration: nextTurnWindow})

	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(true) })

	if perm := g.CastPermissionOnCardByIDForEffect(thisTurn); perm.Granted() {
		t.Errorf("a this-turn grant survived cleanup")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(nextTurn); !perm.Granted() {
		t.Errorf("a grant with a later window was cleared early")
	}
}

func TestAsAnyColorCostFoldsColorsButNotColorless(t *testing.T) {
	cost, err := ParseCost("{2}{W}{U}")
	if err != nil {
		t.Fatal(err)
	}
	any := asAnyColorCost(cost)
	if any.Generic != 4 || len(any.Required) != 0 {
		t.Errorf("{2}{W}{U} → generic %d, required %d; want 4 and 0", any.Generic, len(any.Required))
	}
	// A mono-red pool can now pay a white-blue spell.
	pool := ManaPool{{Color: "R"}, {Color: "R"}, {Color: "R"}, {Color: "R"}}
	if pool.CanPay(cost, 0) {
		t.Errorf("four red should not pay {2}{W}{U} normally")
	}
	if !pool.CanPay(any, 0) {
		t.Errorf("four red should pay it as any-color")
	}
	// "Mana of any color" is not colorless (CR 106.1b), so a {C}
	// requirement is left alone.
	c, err := ParseCost("{1}{C}")
	if err != nil {
		t.Fatal(err)
	}
	folded := asAnyColorCost(c)
	if len(folded.Required) != 1 {
		t.Errorf("{C} should survive the fold, got %+v", folded)
	}
	if (ManaPool{{Color: "R"}, {Color: "R"}}).CanPay(folded, 0) {
		t.Errorf("two red should not pay {1}{C} even under any-color")
	}
}

// TestExileTopWithPermissionTakesTheTopNotTheBottom pins the
// library-order bug the roadmap's batch 01 found: the helper read
// Cards[0], which is the BOTTOM of the library (PopTop, draw and
// mill all take the last element), and every earlier test seeded a
// one-card library where the two coincide. Ragavan and Breeches
// were exiling the bottom card.
func TestExileTopWithPermissionTakesTheTopNotTheBottom(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	bottom := uuid.New()
	top := uuid.New()
	me.Library.PushTop(Card{InstanceID: bottom, Name: "Bottom", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(Card{InstanceID: top, Name: "Top", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	var got []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		got, err = g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, CastPermission{})
	})
	if err != nil {
		t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
	}
	if len(got) != 1 || got[0] != top {
		t.Fatalf("exiled %v, want the top card %v", got, top)
	}
	if !g.Exile.Contains(top) || g.Exile.Contains(bottom) {
		t.Error("the wrong end of the library was exiled")
	}
	if c, err := me.Library.Top(); err != nil || c.InstanceID != bottom {
		t.Errorf("library top after the exile = %v, want the former bottom card", c.InstanceID)
	}
}

// --- CR 903.9 (#1587) ---------------------------------------------
//
// Before #1587, ExileTopWithPermissionForEffect moved cards with a
// raw MoveCard and never opened the CR 614 replacement window, so a
// commander impulse-exiled off the top of its owner's library was
// never offered the command zone. These pin the fix: the prompt is
// offered, nothing moves until it's answered, and the exile-play
// grant lands only on a card that actually reaches exile.

// TestImpulseExiledCommanderOffersCommandZone is #1587's own repro:
// Ragavan-shaped impulse exile off the top of the victim's library,
// with a commander on top.
func TestImpulseExiledCommanderOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	owner.Library.Cards = nil
	cmdID := seatCommander(t, owner.Library, owner)

	var ids []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		ids, err = g.ExileTopWithPermissionForEffect(owner.ID, thief.ID, 1, CastPermission{})
	})
	if err != nil {
		t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("landed %v before the CR 903.9 prompt was answered, want none", ids)
	}
	if g.Exile.Contains(cmdID) {
		t.Fatalf("impulse-exiled commander hit exile before the prompt was answered")
	}
	if !owner.Library.Contains(cmdID) {
		t.Fatalf("commander left the library before the prompt was answered")
	}

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, g.Exile, owner.Library)

	// The commander went to the command zone, not to exile — CR 400.7
	// says it is not the object the exile grant names, so it gets no
	// grant at all.
	if perm := g.CastPermissionOnCardByIDForEffect(cmdID); perm.Granted() {
		t.Errorf("a commander sent to the command zone should not carry an exile-play grant")
	}
}

// TestImpulseExiledCommanderDeclineGrantsPlay is the other half: the
// owner may decline the command zone, and then the card stays in
// exile and gets exactly the grant it would have gotten if it were
// any other card.
func TestImpulseExiledCommanderDeclineGrantsPlay(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	owner.Library.Cards = nil
	cmdID := seatCommander(t, owner.Library, owner)

	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(owner.ID, thief.ID, 1, CastPermission{CastOnly: true}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, g.Exile, owner.Command, owner.Library)

	perm := g.CastPermissionOnCardByIDForEffect(cmdID)
	if perm == nil || !perm.Granted() {
		t.Fatalf("a commander that stayed in exile should carry the exile-play grant")
	}
	if perm.Player != thief.ID {
		t.Errorf("permission holder = %v, want the thief %v", perm.Player, thief.ID)
	}
	if !perm.CastOnly {
		t.Errorf("CastOnly should have ridden through the grant")
	}
}

// TestImpulseExileOfANonCommanderIsUnchanged is the regression: an
// ordinary card still lands and is still granted immediately, with no
// prompt at all — #1587 must not have slowed down the common case.
func TestImpulseExileOfANonCommanderIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	owner.Library.Cards = nil
	loot := seedLibraryTop(owner, "Stolen Bolt", "Instant")

	var ids []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		ids, err = g.ExileTopWithPermissionForEffect(owner.ID, thief.ID, 1, CastPermission{CastOnly: true})
	})
	if err != nil {
		t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
	}
	if len(ids) != 1 || ids[0] != loot {
		t.Fatalf("exiled %v, want [%v] landed immediately", ids, loot)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a non-commander exile queued %d prompts, want 0", len(g.PendingChoices))
	}
	if !g.Exile.Contains(loot) {
		t.Fatalf("card never reached exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(loot)
	if perm == nil || !perm.Granted() {
		t.Fatalf("the exile-play grant should have landed immediately")
	}
	if perm.Player != thief.ID || !perm.CastOnly {
		t.Errorf("grant = %+v, want holder %v and CastOnly", perm, thief.ID)
	}
}

// TestImpulseExileThenWaitsForTheWholeBatch pins the continuation
// form Bonehoard Dracosaur needs: with a commander among the exiled
// cards, `then` does not run until the CR 903.9 prompt is answered,
// and it is handed the true final landed list — not the commander,
// which went to the command zone.
func TestImpulseExileThenWaitsForTheWholeBatch(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	owner.Library.Cards = nil
	loot := seedLibraryTop(owner, "Stolen Bolt", "Instant")
	cmdID := seatCommander(t, owner.Library, owner) // pushed on top, after loot

	var thenCalls int
	var gotLanded []uuid.UUID
	g.WithWriteLock(func() {
		err := g.ExileTopWithPermissionThenForEffect(owner.ID, thief.ID, 2, CastPermission{}, func(_ *Game, landed []uuid.UUID) error {
			thenCalls++
			gotLanded = landed
			return nil
		})
		if err != nil {
			t.Fatalf("ExileTopWithPermissionThenForEffect: %v", err)
		}
	})
	if thenCalls != 0 {
		t.Fatalf("then ran %d times before the CR 903.9 prompt was answered, want 0", thenCalls)
	}
	if g.Exile.Contains(loot) {
		t.Fatalf("the batch landed before the commander's prompt was answered")
	}

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if thenCalls != 1 {
		t.Fatalf("then ran %d times, want exactly 1", thenCalls)
	}
	if len(gotLanded) != 1 || gotLanded[0] != loot {
		t.Fatalf("then landed = %v, want [%v] — the commander went to the command zone", gotLanded, loot)
	}
	assertOnlyIn(t, cmdID, owner.Command, g.Exile, owner.Library)
	assertOnlyIn(t, loot, g.Exile, owner.Command, owner.Library)
	if perm := g.CastPermissionOnCardByIDForEffect(loot); !perm.Granted() {
		t.Errorf("the landed card should carry the exile-play grant")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(cmdID); perm.Granted() {
		t.Errorf("the commander in the command zone should not carry a grant")
	}
}
