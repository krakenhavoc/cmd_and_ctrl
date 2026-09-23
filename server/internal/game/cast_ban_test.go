package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_ban_test.go — #1316, ADR 0066's 2026-09-23 amendment and ADR
// 0085's: a CAST RESTRICTION created by a resolving spell, with a
// CR 611.2 duration.
//
// The model half (this file); the two catalogued cards — Avatar's
// Wrath, Mandate of Peace — are pinned end to end in
// cards/effects/avatars_wrath_test.go and
// cards/effects/mandate_of_peace_test.go.

// banOpponentOutright grants CastBanOutright to p, with the given
// exception zone (empty for none), for the given duration.
func banOpponentOutright(g *Game, p *Player, exceptZone ZoneKind, d Duration) {
	g.WithWriteLock(func() {
		g.GrantCastBanForEffect(p.ID, CastBanRule{Kind: CastBanOutright, ExceptFromZone: exceptZone},
			"Test — can't cast spells", uuid.Nil, d)
	})
}

func banMaxPerTurn(g *Game, p *Player, n int, d Duration) {
	g.WithWriteLock(func() {
		g.GrantCastBanForEffect(p.ID, CastBanRule{Kind: CastBanMaxPerTurn, MaxPerTurn: n},
			"Test — can't cast more than N spells", uuid.Nil, d)
	})
}

func untilYourNextTurn(g *Game, p *Player) Duration {
	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(p.ID) })
	return d
}

func castGate(g *Game, p *Player, card Card, zone ZoneKind) error {
	var err error
	g.ReadSnapshot(func() { err = g.CastGateLocked(p.ID, card, zone, CastSpellParams{}) })
	return err
}

func instantCard(owner uuid.UUID) Card {
	c := NewCard("Test Spell", owner)
	c.TypeLine = "Instant"
	return c
}

// --- the outright ban -------------------------------------------------

// TestOutrightBanRefusesEveryZone is Mandate of Peace's shape: no
// exception, so a cast from any zone the gate is asked about is
// refused.
func TestOutrightBanRefusesEveryZone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	banOpponentOutright(g, me, "", g.UntilEndOfTurnDuration())

	card := instantCard(me.ID)
	for _, zone := range []ZoneKind{ZoneHand, ZoneGraveyard, ZoneExile} {
		var banErr *CantCastError
		if err := castGate(g, me, card, zone); !errors.As(err, &banErr) {
			t.Errorf("cast from %s: got %v, want a *CantCastError", zone, err)
		}
	}
}

// TestOutrightBanWithAnExceptionAllowsOnlyThatZone is Avatar's Wrath's
// shape: everywhere but hand is refused, and hand is not.
func TestOutrightBanWithAnExceptionAllowsOnlyThatZone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	banOpponentOutright(g, me, ZoneHand, untilYourNextTurn(g, me))

	card := instantCard(me.ID)
	if err := castGate(g, me, card, ZoneHand); err != nil {
		t.Errorf("cast from hand should be allowed: %v", err)
	}
	for _, zone := range []ZoneKind{ZoneGraveyard, ZoneExile} {
		var banErr *CantCastError
		if err := castGate(g, me, card, zone); !errors.As(err, &banErr) {
			t.Errorf("cast from %s: got %v, want a *CantCastError", zone, err)
		}
	}
}

// TestCastBanRefusalNamesTheGrant is the attribution half: the refusal
// carries the PlayerStatic's own Label and Source, so a toast can name
// the card that said no, exactly as a battlefield CastRestriction's
// refusal does.
func TestCastBanRefusalNamesTheGrant(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	source := uuid.New()
	g.WithWriteLock(func() {
		g.GrantCastBanForEffect(me.ID, CastBanRule{Kind: CastBanOutright},
			"Mandate of Peace — can't cast spells this turn", source, Duration{})
	})
	var banErr *CantCastError
	err := castGate(g, me, instantCard(me.ID), ZoneHand)
	if !errors.As(err, &banErr) {
		t.Fatalf("got %v, want a *CantCastError", err)
	}
	if banErr.Reason != "Mandate of Peace — can't cast spells this turn" {
		t.Errorf("reason = %q, want the granted label", banErr.Reason)
	}
	if banErr.Source != source {
		t.Errorf("source = %v, want %v", banErr.Source, source)
	}
}

// TestCastBanDoesNotReachAnotherSeat — the grant is per player, and a
// statement about "your opponents" is one grant per opponent
// (RestrictCasting's own doc comment), not one statement about the
// table.
func TestCastBanDoesNotReachAnotherSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]
	banOpponentOutright(g, me, "", g.UntilEndOfTurnDuration())

	if err := castGate(g, other, instantCard(other.ID), ZoneHand); err != nil {
		t.Errorf("an unbanned seat was refused: %v", err)
	}
}

// --- the count shape (#1316's third named case) ------------------------

// TestMaxPerTurnBanReadsTheSameTallyTheStaticRestrictionDoes —
// "each player can't cast more than one spell each turn", granted
// rather than printed on a permanent. No catalogued card ships this
// shape yet; the seam issue named it explicitly, so it is pinned here
// against the model rather than against a card.
func TestMaxPerTurnBanReadsTheSameTallyTheStaticRestrictionDoes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	banMaxPerTurn(g, me, 1, g.UntilEndOfTurnDuration())

	card := instantCard(me.ID)
	if err := castGate(g, me, card, ZoneHand); err != nil {
		t.Errorf("the first spell of the turn should be allowed: %v", err)
	}
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]CastTally)
		}
		g.SpellsCastThisTurn[me.ID] = CastTally{Total: 1}
	})
	var banErr *CantCastError
	if err := castGate(g, me, card, ZoneHand); !errors.As(err, &banErr) {
		t.Errorf("the second spell of the turn: got %v, want a *CantCastError", err)
	}
}

// --- the fast negative --------------------------------------------------

// TestAnyCastRestrictionsForEffectSeesAGrantedBan is the enumerator's
// and the view's fast negative, extended to the new registry: a table
// with nothing on the battlefield but a live granted ban must not
// answer false.
func TestAnyCastRestrictionsForEffectSeesAGrantedBan(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	if g.AnyCastRestrictionsForEffect() {
		t.Fatal("an empty table already answers true")
	}
	banOpponentOutright(g, me, "", g.UntilEndOfTurnDuration())
	if !g.AnyCastRestrictionsForEffect() {
		t.Error("a live granted ban was not seen")
	}
}

// --- the duration ---------------------------------------------------

// TestCastBanEndsAsYourNextTurnBegins mirrors the life-lock boundary
// test: four seats, so the ban sits through three opponents' turns and
// is gone the instant its own player's turn begins.
func TestCastBanEndsAsYourNextTurnBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	banOpponentOutright(g, me, "", untilYourNextTurn(g, me))
	card := instantCard(me.ID)

	for i := 0; i < 3; i++ {
		advanceOneTurn(t, g)
		var banErr *CantCastError
		if err := castGate(g, me, card, ZoneHand); !errors.As(err, &banErr) {
			t.Fatalf("the ban ended after %d opponent turns; it lasts until YOUR next turn", i+1)
		}
	}
	advanceOneTurn(t, g)
	if err := castGate(g, me, card, ZoneHand); err != nil {
		t.Errorf("the ban survived the beginning of its own player's next turn: %v", err)
	}
	if n := len(me.Statics); n != 0 {
		t.Errorf("Player.Statics still holds %d expired entries; the sweep did not run", n)
	}
}

// TestAnUnstampedCastBanIsUntilEndOfTurn — the grant's default. A
// card file that forgot its window gets the narrowest real one rather
// than a permanent ban.
func TestAnUnstampedCastBanIsUntilEndOfTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	banOpponentOutright(g, me, "", Duration{})
	if got := me.Statics[0].Duration.Kind; got != UntilEndOfTurn {
		t.Fatalf("an unstamped ban came back as %v, want until end of turn", got)
	}
	advanceOneTurn(t, g)
	if err := castGate(g, me, instantCard(me.ID), ZoneHand); err != nil {
		t.Errorf("an until-end-of-turn ban outlived its turn: %v", err)
	}
}

// TestExpiredCastBanIsRefusedBeforeTheSweepRuns — the reader tests the
// duration itself, not just the sweep. Stamped into the past by hand.
func TestExpiredCastBanIsRefusedBeforeTheSweepRuns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.GrantCastBanForEffect(me.ID, CastBanRule{Kind: CastBanOutright}, "Test — stale ban", uuid.Nil,
			Duration{Kind: UntilYourNextTurn, Player: me.ID, ExpiresAtTurnsBegun: 1})
	})
	if len(me.Statics) != 1 {
		t.Fatal("the stale ban was not stored; the test is vacuous")
	}
	if err := castGate(g, me, instantCard(me.ID), ZoneHand); err != nil {
		t.Errorf("an expired ban still refuses: %v", err)
	}
}

// --- undo, clone and the snapshot ------------------------------------

// TestCastBanSurvivesCloneUndoAndSnapshot — the grant is plain data on
// PlayerStatic, so it has to rewind and it has to persist, exactly as
// the life-total lock and the timing statement beside it do.
func TestCastBanSurvivesCloneUndoAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	window := untilYourNextTurn(g, me)
	banOpponentOutright(g, me, ZoneHand, window)

	var before *Game
	g.ReadSnapshot(func() { before = g.cloneLocked() })
	got := before.Seats[0].Statics
	if len(got) != 1 || got[0].CastBan.Kind != CastBanOutright {
		t.Fatalf("the clone holds %+v, want one cast ban", got)
	}
	if got[0].Duration != window {
		t.Errorf("the clone's duration is %+v, want %+v", got[0].Duration, window)
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Errorf("a granted cast ban cost the table its restore point: %+v", snap.Continuations)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if err := castGate(restored, restored.Seats[0], instantCard(restored.Seats[0].ID), ZoneGraveyard); err == nil {
		t.Error("the ban did not survive a snapshot round-trip")
	}

	g.RestoreFrom(before)
	if n := len(g.Seats[0].Statics); n != 1 {
		t.Errorf("after the undo Player.Statics holds %d entries, want 1", n)
	}
}
