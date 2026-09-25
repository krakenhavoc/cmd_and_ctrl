package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// control_effects_test.go pins the engine half of #756: gain and
// exchange of control from a spell or an ability. The Aura half
// (Mind Control) is pinned in layer2_control_test.go and
// cards/effects/mind_control_test.go; the card half of this one is in
// cards/effects.

// TestGainControlMovesTheePermanentAndRevertsOnExpiry is the whole
// primitive in one test: a layer-2 scoped static takes the creature,
// the cleanup step ends it, and control goes back to the baseline
// without anything having remembered it.
func TestGainControlMovesThePermanentAndRevertsOnExpiry(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), victim, me.ID,
			g.UntilEndOfTurnDuration(), "test — Act of Treason") {
			t.Fatal("GainControlForEffect refused a battlefield permanent")
		}
	})
	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Fatalf("stolen: controller %s, want %s", got, me.ID)
	}

	advancePastScopedCleanup(t, g)
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup: controller %s, want the baseline %s", got, opp.ID)
	}
}

// TestGainControlRefusesAPermanentThatIsNotThere — nothing to change
// the control of, so nothing is registered and the registry stays
// clean (an entry that can never apply would sit in the snapshot
// census forever).
func TestGainControlRefusesAPermanentThatIsNotThere(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.GainControlForEffect(uuid.New(), uuid.New(), g.Seats[0].ID,
			IndefiniteDuration(), "test") {
			t.Error("GainControlForEffect accepted a permanent that is not on the battlefield")
		}
	})
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("registry holds %d entries after a refused theft", n)
	}
}

// TestGainControlGrantsSummoningSicknessPerController — CR 302.6. The
// creature has been on the battlefield all game and is still sick,
// because sickness is about the CONTROLLER's most recent turn, not the
// permanent's time on the board. This is why Act of Treason prints
// "It gains haste".
func TestGainControlGrantsSummoningSicknessPerController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(victim); ok {
			c.SummonedThisTurn = false
		}
	})
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal")
	})
	g.ReadSnapshot(func() {})

	g.WithWriteLock(func() {
		c, ok := g.battlefieldCardLocked(victim)
		if !ok {
			t.Fatal("the victim left the battlefield")
		}
		if !c.SummonedThisTurn {
			t.Error("a creature that changed control is not summoning-sick under its new controller (CR 302.6)")
		}
	})
}

// TestGainControlRemovesThePermanentFromCombatAndItsAnnouncement is
// CR 506.4 plus the half that was missing (#871): clearing
// AttackingTarget without clearing the announcement left the creature
// marked as already-declared, so its next attack declaration in the
// same combat fired no "whenever ~ attacks" trigger.
func TestGainControlRemovesThePermanentFromCombatAndItsAnnouncement(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(attacker); ok {
			c.AttackingTarget = me.ID
		}
		g.noteAttackAnnouncedLocked(attacker)
	})

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), attacker, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal mid-combat")
	})
	g.ReadSnapshot(func() {})

	g.WithWriteLock(func() {
		c, ok := g.battlefieldCardLocked(attacker)
		if !ok {
			t.Fatal("the attacker left the battlefield")
		}
		if c.AttackingTarget != uuid.Nil {
			t.Error("a permanent that changed control is still attacking (CR 506.4)")
		}
		if g.announcedAttacks[attacker] {
			t.Error("the attack ANNOUNCEMENT survived the control change, so the creature can never announce an attack again this combat (#871)")
		}
	})
}

// TestTwoControlEffectsSortByTimestamp is CR 613.7 through the layer
// engine's own sort: the later theft wins, and when it ends the
// earlier one is applying again — so the creature goes back to the
// FIRST thief, not to its owner.
func TestTwoControlEffectsSortByTimestamp(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	first, second, owner := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := pushScopedTestCreature(g, owner.ID, 2, 2)

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, first.ID,
			IndefiniteDuration(), "first theft")
	})
	if got := controllerOfCard(t, g, victim); got != first.ID {
		t.Fatalf("after the first theft: controller %s, want %s", got, first.ID)
	}

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, second.ID,
			g.UntilEndOfTurnDuration(), "second theft")
	})
	if got := controllerOfCard(t, g, victim); got != second.ID {
		t.Fatalf("the later control effect did not win: controller %s, want %s", got, second.ID)
	}

	advancePastScopedCleanup(t, g)
	if got := controllerOfCard(t, g, victim); got != first.ID {
		t.Errorf("after the later effect ended: controller %s, want the earlier thief %s", got, first.ID)
	}
}

// TestExchangeControlSwapsBothWays — CR 701.12a.
func TestExchangeControlSwapsBothWays(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 3, 3)

	g.WithWriteLock(func() {
		if !g.ExchangeControlForEffect(uuid.New(), mine, theirs, "test — Switcheroo") {
			t.Fatal("ExchangeControlForEffect refused two battlefield permanents")
		}
	})
	if got := controllerOfCard(t, g, mine); got != opp.ID {
		t.Errorf("mine: controller %s, want %s", got, opp.ID)
	}
	if got := controllerOfCard(t, g, theirs); got != me.ID {
		t.Errorf("theirs: controller %s, want %s", got, me.ID)
	}
}

// TestExchangeControlIsOneEffectWithOneTimestamp — CR 701.12 makes an
// exchange a single effect, and CR 613.7 orders effects by timestamp.
// Two clock reads would let a third control-changer resolve "between"
// the halves.
func TestExchangeControlIsOneEffectWithOneTimestamp(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 3, 3)

	g.WithWriteLock(func() {
		g.ExchangeControlForEffect(uuid.New(), mine, theirs, "test — Switcheroo")
	})
	if n := len(g.ScopedEffects); n != 2 {
		t.Fatalf("an exchange registered %d entries, want 2", n)
	}
	if a, b := g.ScopedEffects[0].Timestamp, g.ScopedEffects[1].Timestamp; a != b {
		t.Errorf("the two halves have different timestamps (%d, %d); CR 701.12 makes them one effect", a, b)
	}
}

// TestExchangeControlFailsWholeWhenOneObjectIsGone — CR 701.12b. All
// or nothing: the surviving permanent does not change controller, and
// nothing is registered.
func TestExchangeControlFailsWholeWhenOneObjectIsGone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)

	g.WithWriteLock(func() {
		if g.ExchangeControlForEffect(uuid.New(), mine, uuid.New(), "test — Switcheroo") {
			t.Error("an exchange with a missing object reported success")
		}
	})
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("a failed exchange registered %d entries; CR 701.12b says none", n)
	}
	if got := controllerOfCard(t, g, mine); got != me.ID {
		t.Errorf("the surviving permanent changed controller: %s, want %s", got, me.ID)
	}
}

// TestGainControlEndsWhenTheStolenPermanentIsFlickered — CR 400.7. The
// creature that comes back is a new object, so the theft does not
// follow it, and the entry is dropped rather than left to sit in the
// registry for the rest of the game.
func TestGainControlEndsWhenTheStolenPermanentIsFlickered(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID, IndefiniteDuration(), "indefinite theft")
	})
	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, g.Exile, victim); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: victim,
			OldZone: ZoneBattlefield, NewZone: ZoneExile})
		if _, err := MoveCard(g.Exile, g.Battlefield, victim); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: victim,
			OldZone: ZoneExile, NewZone: ZoneBattlefield})
	})

	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Errorf("a flickered permanent is still stolen: controller %s, want %s", got, opp.ID)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("the theft outlived the object it was pinned to (%d entries)", n)
	}
}

// TestAStolenCommanderKeepsItsCommanderState — a commander is a
// commander because of who OWNS it (CR 903.3), not who controls it.
// Stealing one and letting it die must still offer its owner the
// command-zone replacement.
func TestAStolenCommanderKeepsItsCommanderState(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(cmdr); ok {
			c.IsCommander = true
		}
	})

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), cmdr, me.ID, IndefiniteDuration(), "steal a commander")
	})
	if got := controllerOfCard(t, g, cmdr); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	g.WithWriteLock(func() {
		c, ok := g.battlefieldCardLocked(cmdr)
		if !ok {
			t.Fatal("the commander left the battlefield")
		}
		if !c.IsCommander {
			t.Error("a stolen commander stopped being a commander")
		}
		if c.Owner != opp.ID {
			t.Errorf("a stolen commander's owner changed: %s, want %s", c.Owner, opp.ID)
		}
	})
}

// TestUndoAcrossAControlChange — the undo stack carries the registry,
// so rewinding past the spell that stole the creature hands it back.
func TestUndoAcrossAControlChange(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.ReadSnapshot(func() {})

	snap := g.Clone()

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID, IndefiniteDuration(), "theft")
	})
	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("undo left %d scoped statics", n)
	}
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Errorf("after undo: controller %s, want %s", got, opp.ID)
	}
}

// TestUndoAcrossAnExpiry is the other direction: a snapshot taken
// while the theft was live must still hold it after the original's
// sweep has dropped it, which is what the fresh-slice sweep buys.
func TestUndoAcrossAnExpiry(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID,
			g.UntilEndOfTurnDuration(), "theft until end of turn")
	})
	snap := g.Clone()

	advancePastScopedCleanup(t, g)
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Fatalf("setup: the theft did not expire (controller %s)", got)
	}
	if len(snap.ScopedEffects) != 1 {
		t.Fatalf("the sweep reached into the snapshot (%d entries)", len(snap.ScopedEffects))
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Errorf("after undoing the expiry: controller %s, want %s", got, me.ID)
	}
}

// TestSnapshotRoundTripCarriesTheTheft — since ADR 0041 phase 3
// (#1497) a theft is a data record, so a game holding one is a restore
// point: the snapshot carries the record, the census stays empty, and
// the restored game still has the permanent under the thief's control
// after its own layer pass — for the rest of the game, which is the
// case that used to freeze the restore point for the rest of the game.
func TestSnapshotRoundTripCarriesTheTheft(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID, IndefiniteDuration(), "Agent of Treachery")
	})
	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding only a data-backed theft is not a restore point: %+v", snap.Continuations)
	}
	if len(snap.ScopedEffects) != 1 {
		t.Fatalf("the snapshot carried %d scoped effects, want 1", len(snap.ScopedEffects))
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	if got := controllerOfCard(t, restored, victim); got != me.ID {
		t.Errorf("restored controller %s, want the thief %s", got, me.ID)
	}
	found := false
	restored.ReadSnapshot(func() {
		for _, c := range restored.Battlefield.Cards {
			if c.InstanceID == victim {
				found = true
				if c.Owner != opp.ID {
					t.Errorf("restored owner %s, want %s", c.Owner, opp.ID)
				}
			}
		}
	})
	if !found {
		t.Error("the stolen permanent did not survive the round trip")
	}
	if got := restored.Seats[0].TurnsBegun; got != g.Seats[0].TurnsBegun {
		t.Errorf("restored TurnsBegun = %d, want %d", got, g.Seats[0].TurnsBegun)
	}

	// And the restored record still ENDS: the permanent leaving takes
	// the theft with it (the pin, CR 400.7).
	if err := restored.MoveCardByID(ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, victim); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	if n := len(restored.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects survive the stolen permanent leaving the restored game, want 0", n)
	}
}
