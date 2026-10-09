package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_opponents_test.go — #2745, CR 702.16i over CR 702.16k:
// "you have protection from each of your opponents" (Absolute Virtue).
//
// Like the chosen-player quality next door, the source is tested by
// who CONTROLS it, and every test checks both halves: an opponent is
// refused, and the protected player's own source is not.

// withOpponentsProtection gives `holder` the derived grant through a
// stubbed catalog permanent, the way Absolute Virtue's static does.
func withOpponentsProtection(t *testing.T, g *Game, holder *Player) uuid.UUID {
	t.Helper()
	const virtue = "test-protection-from-each-of-your-opponents"
	withCatalogPlayerKeywords(t, func(id string) []string {
		if id == virtue {
			return []string{ProtectionFromEachOfYourOpponents}
		}
		return nil
	})
	return pushLeyline(t, g, holder, virtue)
}

// TestTheOpponentsQualityIsResolvedByTheReader — the token names no
// seat, so a bare parse protects from nobody, and both readers bind
// "your" to the one who has the ability.
func TestTheOpponentsQualityIsResolvedByTheReader(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bare, ok := ParseProtectionQuality(ProtectionFromEachOfYourOpponents)
	if !ok {
		t.Fatal("the grammar accepts the token")
	}
	if bare.Kind != ProtectionQualityOpponents || bare.Kind.String() != "opponents" {
		t.Errorf("kind %v (%q), want opponents", bare.Kind, bare.Kind.String())
	}
	if bare.Printed != "each of your opponents" {
		t.Errorf("printed %q", bare.Printed)
	}
	if bare.Matches(&Characteristic{Controller: opp.ID}) {
		t.Error("an unbound opponents quality matches nothing")
	}

	bound := bindPlayerProtectionQuality(me.ID, bare)
	if !bound.Matches(&Characteristic{Controller: opp.ID}) {
		t.Error("a source an opponent controls has the quality")
	}
	if bound.Matches(&Characteristic{Controller: me.ID}) {
		t.Error("a source you control does not")
	}
	if bound.Matches(&Characteristic{}) {
		t.Error("a source with no controller matches nothing")
	}

	// On a permanent, "your" is its controller.
	id := pushColouredCreature(g, opp, "Test Virtue", []string{"W"}, ProtectionFromEachOfYourOpponents)
	card, _ := battlefieldCardByID(g, id)
	qs := ProtectionQualities(&card)
	if len(qs) != 1 || qs[0].Holder != opp.ID {
		t.Fatalf("the reader bound %+v, want the controller %v", qs, opp.ID)
	}
	if !ProtectedFrom(&card, &Characteristic{Controller: me.ID}) {
		t.Error("the permanent is protected from its controller's opponent")
	}
	if ProtectedFrom(&card, &Characteristic{Controller: opp.ID}) {
		t.Error("the permanent is not protected from its own controller")
	}
}

// TestProtectionFromOpponentsRefusesTheirTargetsAndNotYours is CR
// 702.16b: no opponent's spell may target the player, their own spell
// still may, and the protection is theirs alone.
func TestProtectionFromOpponentsRefusesTheirTargetsAndNotYours(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	const bolt = "test-protection-opponents-bolt"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})
	withOpponentsProtection(t, g, me)

	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at the protected player: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, me.ID); err != nil {
		t.Fatalf("your own spell at yourself: %v", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the unprotected opponent: %v", err)
	}
}

// TestProtectionFromOpponentsPreventsOnlyTheirDamage is CR 702.16e,
// both ways: an opponent's source deals the player nothing, and the
// player's own source still hurts them.
func TestProtectionFromOpponentsPreventsOnlyTheirDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	withOpponentsProtection(t, g, me)

	theirs := NewCard("Their Pinger", opp.ID)
	theirs.TypeLine = "Artifact"
	mine := NewCard("My Pinger", me.ID)
	mine.TypeLine = "Artifact"
	start := me.Life
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(theirs)
		g.Battlefield.PushTop(mine)
		if err := g.DealDamageToPlayerForEffect(theirs.InstanceID, me.ID, 3); err != nil {
			t.Fatalf("an opponent's source: %v", err)
		}
	})
	if me.Life != start {
		t.Errorf("an opponent's source dealt %d damage; CR 702.16e prevents it", start-me.Life)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine.InstanceID, me.ID, 2); err != nil {
			t.Fatalf("your own source: %v", err)
		}
	})
	if me.Life != start-2 {
		t.Errorf("life %d after your own source dealt 2, want %d", me.Life, start-2)
	}

	attacker := pushColouredCreature(g, opp, "Their Attacker", []string{"G"})
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, attacker, 5, "") })
	if me.Life != start-2 {
		t.Errorf("an opponent's attacker dealt combat damage: life %d, want %d", me.Life, start-2)
	}
}

// TestProtectionFromOpponentsDropsTheirAuraAndKeepsYours is CR 702.16c:
// an opponent's Curse falls off the player, the player's own Aura on
// themselves stays.
func TestProtectionFromOpponentsDropsTheirAuraAndKeepsYours(t *testing.T) {
	// Four seats, so a state-based action pass is not a game over.
	g := newFourPlayerActiveGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	theirs := pushAttachTestCard(g, foe.ID, "Their Curse", "Enchantment — Aura Curse")
	mine := pushAttachTestCard(g, me.ID, "My Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{theirs, mine} {
			if err := g.AttachForEffect(id, TargetRef{Kind: TargetPlayer, ID: me.ID}); err != nil {
				t.Fatalf("AttachForEffect: %v", err)
			}
		}
		g.runStateChecksLocked()
	})

	withOpponentsProtection(t, g, me)
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if _, ok := battlefieldCardByID(g, theirs); ok {
		t.Error("an opponent's Aura on a player protected from their opponents must fall off")
	}
	if got, ok := battlefieldCardByID(g, mine); !ok || !got.IsAttached() {
		t.Errorf("the player's own Aura must stay (onBattlefield=%v)", ok)
	}
}
