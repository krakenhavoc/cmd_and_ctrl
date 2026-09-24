package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_source_lki_test.go covers #1396 (ADR 0056 Decision 2 step 4,
// closed by the 2026-09-24 amendment): damage from a source that has
// LEFT the battlefield carries the lifelink and deathtouch the source
// had as it last existed there (CR 608.2h, CR 702.15b, CR 702.2b), read
// off the #1379 per-object record, and never the keywords or controller
// of a new object the same card has become since (CR 400.7).
//
// The live-source half is noncombat_keywords_test.go (#711), which this
// change must leave passing untouched; TestLiveDamageSourceStillReadsLive
// below pins the one new branch it gained (a ref that names the live
// object).

// pushPrintedKeywordCreature puts a creature on the battlefield with PRINTED
// keywords, so a layer recompute keeps them and the departure record
// (which reads Effective()) sees them. Owner and controller may differ,
// which is how these tests tell two objects of one card apart.
func pushPrintedKeywordCreature(g *Game, id uuid.UUID, owner, controller uuid.UUID, power, toughness int, keywords ...string) {
	c := NewCard("Keyword Creature", owner)
	c.InstanceID = id
	c.TypeLine = "Creature — Test"
	c.Power = power
	c.Toughness = toughness
	c.Controller = controller
	c.Keywords = append([]string(nil), keywords...)
	g.Battlefield.PushTop(c)
	g.RecomputeLayersIfStaleLocked()
}

// returnUnderOwner moves a card from its owner's graveyard back onto the
// battlefield as a new object under its OWNER's control, the way a
// reanimation "under its owner's control" would. The raw move keeps the
// last controller, so the test sets it.
func returnUnderOwner(t *testing.T, g *Game, id, owner uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: owner}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID back: %v", err)
	}
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.Controller, c.BaseController = owner, owner
		g.recomputeLayersLocked()
	})
}

// A lifelinker that has died still gains its controller the life when
// its damage lands afterwards — the "when this dies, it deals damage"
// shape, and a lifelinker killed in response to Warstorm Surge. Played
// through the paused and the unpaused CR 614 paths: the keyword rides
// the tail, so an ordering prompt answered later must credit the same.
func TestDepartedDamageSourceStillHasItsLifelink(t *testing.T) {
	srcID := uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushPrintedKeywordCreature(g, srcID, g.Seats[0].ID, g.Seats[0].ID, 3, 3, "lifelink")
			if err := g.DestroyPermanentForEffect(srcID); err != nil {
				t.Fatalf("DestroyPermanentForEffect: %v", err)
			}
			if findBattlefieldCard(g, srcID) != nil {
				t.Fatal("setup: the lifelinker should have left")
			}
		},
		deal: func(g *Game) {
			_ = g.DealDamageToPlayerForEffect(srcID, g.Seats[1].ID, 3)
		},
	})
	if got := g.Seats[1].Life; got != StartingLife-4 {
		t.Errorf("target life = %d, want %d", got, StartingLife-4)
	}
	if got := g.Seats[0].Life; got != StartingLife+4 {
		t.Errorf("controller life = %d, want %d — a departed lifelinker's damage "+
			"still gains life (CR 608.2h, #1396)", got, StartingLife+4)
	}
}

// A deathtouch creature that has died still destroys what its damage
// hits: one damage to a 1/10 is lethal.
func TestDepartedDamageSourceStillHasItsDeathtouch(t *testing.T) {
	srcID, victimID := uuid.New(), uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushPrintedKeywordCreature(g, srcID, g.Seats[0].ID, g.Seats[0].ID, 1, 1, "deathtouch")
			pushTestCreature(g, victimID, g.Seats[1], 1, 10)
			if err := g.DestroyPermanentForEffect(srcID); err != nil {
				t.Fatalf("DestroyPermanentForEffect: %v", err)
			}
		},
		deal: func(g *Game) {
			// (2-1)*2 = 2 damage, far short of 10 toughness: only
			// deathtouch kills.
			_ = g.DealDamageToCreatureForEffect(srcID, victimID, 2)
		},
	})
	if findBattlefieldCard(g, victimID) != nil {
		t.Error("the 1/10 survived damage from a departed deathtouch source " +
			"(CR 608.2h, #1396)")
	}
}

// The object-named entry point reads the same record.
func TestDepartedDamageSourceByObjectRef(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID, victimID := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, srcID, me.ID, me.ID, 2, 2, "deathtouch", "lifelink")
		pushTestCreature(g, victimID, opp, 5, 5)
	})
	ref := refOf(t, g, srcID)
	destroy(t, g, srcID)
	g.WithWriteLock(func() {
		if err := g.DealDamageFromObjectForEffect(ref, victimID, 1); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
		if err := g.DealDamageFromObjectForEffect(ref, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
	if findBattlefieldCard(g, victimID) != nil {
		t.Error("the 5/5 survived 1 damage from the departed deathtouch object")
	}
	if me.Life != StartingLife+3 {
		t.Errorf("controller life = %d, want %d: both hits carry the departed "+
			"object's lifelink", me.Life, StartingLife+3)
	}
}

// CR 400.7, the object-named path: the card left and came back. The
// permanent on the battlefield now is a new object under a different
// controller, and the damage named the OLD object — so the old
// object's controller is credited, not the new one's.
func TestReturnedDamageSourceIsNotTheObjectTheRefNamed(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	g.WithWriteLock(func() { pushTestCreature(g, victim, thief, 5, 5) })
	srcID := uuid.New()
	// The first object is controlled by the thief.
	g.WithWriteLock(func() { pushPrintedKeywordCreature(g, srcID, owner.ID, thief.ID, 2, 2, "lifelink") })
	first := refOf(t, g, srcID)
	destroy(t, g, srcID)
	returnUnderOwner(t, g, srcID, owner.ID)
	second := refOf(t, g, srcID)
	if second == first {
		t.Fatal("setup: the returned card should be a new object")
	}
	g.WithWriteLock(func() {
		back := findBattlefieldCard(g, srcID)
		if back == nil || back.Controller != owner.ID {
			t.Fatalf("setup: the new object should be the owner's, got %+v", back)
		}
		if err := g.DealDamageFromObjectForEffect(first, victim, 2); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
	})
	if thief.Life != StartingLife+2 {
		t.Errorf("thief life = %d, want %d: the departed object's controller is "+
			"credited", thief.Life, StartingLife+2)
	}
	if owner.Life != StartingLife {
		t.Errorf("owner life = %d, want %d: the new object dealt nothing", owner.Life, StartingLife)
	}
}

// CR 400.7, the instance-ID path: the card left under the thief, came
// back under its owner and left again. The source a bare ID names is
// the LAST object it was, so the owner is credited — the first object's
// record must not answer for the second.
func TestBareDamageSourceReadsTheMostRecentDepartedObject(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	g.WithWriteLock(func() { pushTestCreature(g, victim, thief, 5, 5) })
	srcID := uuid.New()
	g.WithWriteLock(func() { pushPrintedKeywordCreature(g, srcID, owner.ID, thief.ID, 2, 2, "lifelink") })
	destroy(t, g, srcID)
	returnUnderOwner(t, g, srcID, owner.ID)
	destroy(t, g, srcID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(srcID, victim, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if owner.Life != StartingLife+2 || thief.Life != StartingLife {
		t.Errorf("owner %d, thief %d; want the second object's controller (owner) "+
			"credited 2 and the first object's (thief) nothing", owner.Life, thief.Life)
	}
}

// A card that has moved on since it left is no longer the departed
// permanent: died, then exiled from the graveyard, it carries nothing.
// The bare-ID read only answers while the card sits where it went.
func TestBareDamageSourceThatMovedOnHasNoKeywords(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID := uuid.New()
	g.WithWriteLock(func() { pushPrintedKeywordCreature(g, srcID, me.ID, me.ID, 2, 2, "lifelink") })
	destroy(t, g, srcID)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, srcID); err != nil {
		t.Fatalf("MoveCardByID to exile: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(srcID, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if me.Life != StartingLife {
		t.Errorf("controller life = %d, want %d: a card exiled from the graveyard "+
			"is not the permanent that died", me.Life, StartingLife)
	}
	if opp.Life != StartingLife-2 {
		t.Errorf("target life = %d, want %d: the damage itself lands", opp.Life, StartingLife-2)
	}
}

// A token that died has ceased to exist (CR 704.5d) and is in no zone at
// all; it cannot come back (CR 111.8), so its last record answers for it.
func TestDepartedTokenDamageSourceStillHasItsLifelink(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var tok uuid.UUID
	g.WithWriteLock(func() {
		made, err := g.CreateTokensForEffect(me.ID, Card{
			Name: "Lifelink Spirit", TypeLine: "Token Creature — Spirit",
			Power: 1, Toughness: 1, Keywords: []string{"lifelink"},
		}, 1, TokenEntryOptions{})
		if err != nil || len(made) != 1 {
			t.Fatalf("CreateTokensForEffect: %v %v", made, err)
		}
		tok = made[0]
	})
	destroy(t, g, tok)
	g.WithWriteLock(func() {
		g.runStateChecksLocked()
		if g.cardObjectEpochLocked(tok) != -1 {
			t.Fatal("setup: the token should have ceased to exist")
		}
		if err := g.DealDamageToPlayerForEffect(tok, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if me.Life != StartingLife+2 {
		t.Errorf("controller life = %d, want %d: a token that has ceased to exist "+
			"still dealt its damage with lifelink", me.Life, StartingLife+2)
	}
}

// The live path is unchanged, and a ref naming the live object reads it
// live: a keyword the permanent has NOW counts.
func TestLiveDamageSourceStillReadsLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID := uuid.New()
	g.WithWriteLock(func() { pushPrintedKeywordCreature(g, srcID, me.ID, me.ID, 2, 2, "lifelink") })
	ref := refOf(t, g, srcID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(srcID, opp.ID, 1); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
		if err := g.DealDamageFromObjectForEffect(ref, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
	})
	if me.Life != StartingLife+3 {
		t.Errorf("controller life = %d, want %d", me.Life, StartingLife+3)
	}
	if opp.Life != StartingLife-3 {
		t.Errorf("target life = %d, want %d", opp.Life, StartingLife-3)
	}
}
