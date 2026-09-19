package game

import (
	"testing"

	"github.com/google/uuid"
)

// sacrificed_this_way_test.go pins #910: sacrifice is a batch with a
// continuation, on the one body every other verb already shared.
//
// Destroy got it in #815, exile and bounce in #866, mill in #893.
// Sacrifice had only the per-card call, so Living Death's second pass
// fired and forgot and God-Eternal Bontu drew a card for a commander
// whose owner was still being ASKED about the command zone.
//
// Sacrifice itself is not replaceable (CR 701.17a — it is not a
// destruction, so nothing that replaces destruction touches it), but
// the MOVE it makes is an ordinary zone change: the CR 614 window
// opens over it, a commander's CR 903.9 prompt can pause a leg, and a
// "graveyard becomes exile" replacement can rewrite where the card
// goes. What "sacrificed this way" counts is therefore the permanent
// that LEFT THE BATTLEFIELD, wherever it landed — which is where this
// verb parts company with destroy, and sacrificedThisWayLocked says
// why.

// sacrificeAllThen runs the sweep under the write lock and returns the
// sacrificed list the continuation was handed, plus how many times it
// ran.
func sacrificeAllThen(t *testing.T, g *Game, ids []uuid.UUID) (*[]uuid.UUID, *int) {
	t.Helper()
	var got []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.SacrificeAllThenForEffect(uuid.Nil, ids, func(_ *Game, sacrificed []uuid.UUID) error {
			ran++
			got = sacrificed
			return nil
		})
		if err != nil {
			t.Fatalf("SacrificeAllThenForEffect: %v", err)
		}
	})
	return &got, &ran
}

// TestSacrificeAllThenReportsWhatLeftTheBattlefield is the plain case:
// two creatures, both sacrificed, both reported, and each announced
// before it moved.
func TestSacrificeAllThenReportsWhatLeftTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	first := pushBear(g, me.ID)
	second := pushBear(g, me.ID)

	got, ran := sacrificeAllThen(t, g, []uuid.UUID{first, second})

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{first, second}) {
		t.Errorf("sacrificed this way = %v, want both", *got)
	}
	for _, id := range []uuid.UUID{first, second} {
		if !me.Graveyard.Contains(id) {
			t.Error("a sacrificed permanent goes to its owner's graveyard (CR 701.17a)")
		}
		if !hasEvent(g, EventSacrifice, id) {
			t.Error("every leg of the batch announces its own EventSacrifice")
		}
	}
}

// TestASacrificedCommanderIsStillSacrificed is the CR 903.9 half and
// the issue's shape: the list cannot be taken while the prompt is
// open, so the rest of the batch waits for it — and the commander
// counts on both answers, because the sacrifice happened either way
// and only where the card went was replaced.
func TestASacrificedCommanderIsStillSacrificed(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			commander := seatCommander(t, g.Battlefield, me)
			bear := pushBear(g, me.ID)

			got, ran := sacrificeAllThen(t, g, []uuid.UUID{commander, bear})

			if *ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", *ran)
			}
			if findBattlefieldCard(g, bear) == nil {
				t.Error("the rest of the batch waits for the paused leg")
			}
			if findBattlefieldCard(g, commander) == nil {
				t.Error("a paused leg has moved nothing")
			}

			prompt := expectCommanderPrompt(t, g, me)
			if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			if *ran != 1 {
				t.Fatalf("the continuation ran %d times after the answer, want 1", *ran)
			}
			if !idsEqual(*got, []uuid.UUID{commander, bear}) {
				t.Errorf("sacrificed this way = %v, want both — the commander was sacrificed "+
					"whichever zone it went to (CR 701.17a)", *got)
			}
			landed := me.Graveyard
			if tc.commandZone {
				landed = me.Command
			}
			if !landed.Contains(commander) {
				t.Errorf("the commander is in the %s", landed.Kind)
			}
			if !me.Graveyard.Contains(bear) {
				t.Error("the rest of the batch lands once the prompt is answered")
			}
		})
	}
}

// TestASacrificeAReplacementExiledIsStillSacrificed is the rule that
// parts company with destroy, stated on its own because it is a
// judgement rather than an obvious bug. CR 701.7a defines a
// DESTRUCTION by the graveyard, so a permanent sent to exile instead
// was not destroyed; CR 701.17a's sacrifice is the controller's move
// OFF the battlefield, and Rest in Peace taking the card does not undo
// it.
func TestASacrificeAReplacementExiledIsStillSacrificed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stolen := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(stolen,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	got, ran := sacrificeAllThen(t, g, []uuid.UUID{stolen})

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !g.Exile.Contains(stolen) {
		t.Fatal("the replacement rewrote the destination")
	}
	if !idsEqual(*got, []uuid.UUID{stolen}) {
		t.Errorf("sacrificed this way = %v, want the exiled permanent: it was sacrificed, "+
			"and only where it went was replaced", *got)
	}
	if !hasEvent(g, EventSacrifice, stolen) {
		t.Error("the announcement happens before the window, so it fires whatever the window does")
	}
}

// TestACancelledSacrificeIsNotASacrifice is the other end: the CR 614
// window cancels the move outright and the permanent never leaves, so
// nothing was sacrificed however loudly it was announced.
func TestACancelledSacrificeIsNotASacrifice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	saved := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(saved,
			"it isn't put into a graveyard", func(ev *ReplacementEvent) { ev.Cancel() }))
	})

	got, ran := sacrificeAllThen(t, g, []uuid.UUID{saved})

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if findBattlefieldCard(g, saved) == nil {
		t.Fatal("a cancelled move leaves the permanent on the battlefield")
	}
	if len(*got) != 0 {
		t.Errorf("sacrificed this way = %v, want nothing: it never left the battlefield", *got)
	}
}

// TestUndoAcrossAPausedSacrificeLegReplays is the undo contract every
// continuation in the engine signs.
func TestUndoAcrossAPausedSacrificeLegReplays(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	commander := seatCommander(t, g.Battlefield, me)
	bear := pushBear(g, me.ID)

	got, ran := sacrificeAllThen(t, g, []uuid.UUID{commander, bear})
	promptOpen := g.Clone()

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := append([]uuid.UUID(nil), (*got)...)
	if !idsEqual(first, []uuid.UUID{commander, bear}) {
		t.Fatalf("first run sacrificed %v, want both", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if findBattlefieldCard(g, commander) == nil || findBattlefieldCard(g, bear) == nil {
		t.Fatal("the rewind puts both permanents back on the battlefield")
	}

	replay := expectCommanderPrompt(t, g, g.Seats[0])
	if err := g.ResolveOptionalReplacement(replay.ID, g.Seats[0].ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed sacrificed this way = %v, want the first run's %v", *got, first)
	}
}

// TestFireAndForgetSacrificeCountsWhatSettled — the batch's synchronous
// twin keeps the posture the other fire-and-forget sweeps have: it
// proceeds AROUND a paused leg, and the count is the legs that really
// left by the time it returns. A card that reads the number uses the
// Then form, which waits.
func TestFireAndForgetSacrificeCountsWhatSettled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	commander := seatCommander(t, g.Battlefield, me)
	bear := pushBear(g, me.ID)

	var n int
	g.WithWriteLock(func() { n = g.SacrificeAllForEffect(uuid.Nil, []uuid.UUID{commander, bear}) })

	if n != 1 {
		t.Errorf("SacrificeAllForEffect = %d, want 1 — the commander is still being asked", n)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the sweep proceeds around the paused leg")
	}
	if findBattlefieldCard(g, commander) == nil {
		t.Error("the paused leg has moved nothing")
	}

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Graveyard.Contains(commander) {
		t.Error("the answered leg lands afterwards, outside the returned count")
	}
}

// TestSingleSacrificeIsUnchanged — SacrificePermanentForEffect is one
// leg of the same body now, so the contract every caller relies on has
// to be exactly what it was: the announcement, the move, and
// ErrCardNotFound for a permanent that is not there.
func TestSingleSacrificeIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
		if err := g.SacrificePermanentForEffect(uuid.New()); err != ErrCardNotFound {
			t.Errorf("sacrificing a permanent that is not on the battlefield = %v, want ErrCardNotFound", err)
		}
	})

	if !me.Graveyard.Contains(bear) {
		t.Error("the single sacrifice still routes to the graveyard")
	}
	if !hasEvent(g, EventSacrifice, bear) {
		t.Error("the single sacrifice still announces")
	}
}
