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
func impulseExile(t *testing.T, g *Game, victim, thief *Player, perm ExilePlayPermission) uuid.UUID {
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

func TestExilePlayPermissionActive(t *testing.T) {
	me, you := uuid.New(), uuid.New()
	cases := []struct {
		name string
		perm ExilePlayPermission
		who  uuid.UUID
		turn int
		want bool
	}{
		{"zero value grants nothing", ExilePlayPermission{}, me, 1, false},
		{"holder, in window", ExilePlayPermission{Player: me, UntilTurn: 3}, me, 3, true},
		{"holder, earlier turn", ExilePlayPermission{Player: me, UntilTurn: 3}, me, 2, true},
		{"holder, expired", ExilePlayPermission{Player: me, UntilTurn: 3}, me, 4, false},
		{"somebody else", ExilePlayPermission{Player: me, UntilTurn: 3}, you, 3, false},
	}
	for _, tc := range cases {
		if got := tc.perm.Active(tc.who, tc.turn); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestExileTopWithPermissionMovesAndStamps(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	loot := seedLibraryTop(opp, "Stolen Bolt", "Instant")

	id := impulseExile(t, g, opp, me, ExilePlayPermission{CastOnly: true})
	if id != loot {
		t.Fatalf("exiled %v, want the top card %v", id, loot)
	}
	if !g.Exile.Contains(loot) {
		t.Fatalf("card never reached exile")
	}
	var got ExilePlayPermission
	for _, c := range g.Exile.Cards {
		if c.InstanceID == loot {
			got = c.ExilePlay
		}
	}
	if got.Player != me.ID {
		t.Errorf("permission holder = %v, want the thief %v", got.Player, me.ID)
	}
	if got.UntilTurn != g.Turn.Number {
		t.Errorf("UntilTurn = %d, want this turn %d", got.UntilTurn, g.Turn.Number)
	}
	if !got.CastOnly {
		t.Errorf("CastOnly should have ridden through")
	}
	// An empty library is not an error — you exile what's there.
	opp.Library.Cards = nil
	g.WithWriteLock(func() {
		ids, err := g.ExileTopWithPermissionForEffect(opp.ID, me.ID, 2, ExilePlayPermission{})
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
		_, _ = g.ExileTopWithPermissionForEffect(opp.ID, opp.ID, 1, ExilePlayPermission{})
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
	loot := impulseExile(t, g, opp, me, ExilePlayPermission{})

	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("this turn it should be castable: %v", err)
	}
	// Put it back and roll the turn over.
	g.WithWriteLock(func() {
		_, _ = MoveCard(g.Stack, g.Exile, loot)
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == loot {
				g.Exile.Cards[i].ExilePlay = ExilePlayPermission{Player: me.ID, UntilTurn: g.Turn.Number}
			}
		}
		delete(g.StackMeta, loot)
		g.Turn.Number++
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
	loot := impulseExile(t, g, opp, me, ExilePlayPermission{CastOnly: true})

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
	for _, c := range g.Stack.Cards {
		if c.InstanceID == loot && c.ExilePlay.Granted() {
			t.Errorf("permission survived the cast")
		}
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
	stranded := impulseExile(t, g, opp, me, ExilePlayPermission{CastOnly: true})
	if err := g.CastSpell(me.ID, stranded, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Fatalf("cast-only land: %v, want ErrNoPlayPermission", err)
	}
	if !g.Exile.Contains(stranded) {
		t.Errorf("a rejected land left exile")
	}

	seedLibraryTop(opp, "Playable Island", "Basic Land — Island")
	playable := impulseExile(t, g, opp, me, ExilePlayPermission{})
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
			if c.ExilePlay.Granted() {
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
	thisTurn := impulseExile(t, g, opp, me, ExilePlayPermission{})
	seedLibraryTop(opp, "Next Turn", "Instant")
	nextTurn := impulseExile(t, g, opp, me, ExilePlayPermission{UntilTurn: g.Turn.Number + 1})

	g.WithWriteLock(func() { g.clearExpiredExilePlayLocked() })

	for _, c := range g.Exile.Cards {
		switch c.InstanceID {
		case thisTurn:
			if c.ExilePlay.Granted() {
				t.Errorf("a this-turn grant survived cleanup")
			}
		case nextTurn:
			if !c.ExilePlay.Granted() {
				t.Errorf("a grant with a later window was cleared early")
			}
		}
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
		got, err = g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, ExilePlayPermission{})
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
