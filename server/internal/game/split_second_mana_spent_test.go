package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// split_second_mana_spent_test.go — #4724: a split-second grant
// conditioned on the mana spent to cast the spell, Shadow the Hedgehog's
// "Each spell you cast has split second if mana from an artifact was
// spent to cast it" (SplitSecondGrant.YouCast and SpentFrom). The
// grant is still read live off the battlefield (ADR 0007, decision 16);
// what it adds is the spell's payment record (#1212) and its caster.

const chaosControlKey = "chaos-control-test"

// chaosControlGame seats a permanent declaring the grant under seat 0
// and stops at a main phase.
func chaosControlGame(t *testing.T) (*Game, *Player, *Player, uuid.UUID) {
	t.Helper()
	withCatalogDefs(t, map[string]*CardDef{
		chaosControlKey: {SplitSecondGrants: []SplitSecondGrant{{
			YouCast:   true,
			SpentFrom: ManaSourceArtifact,
			Label:     "Each spell you cast has split second if mana from an artifact was spent to cast it.",
		}}},
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: src, Name: "Chaos Source", TypeLine: "Legendary Creature — Hedgehog",
			OracleID: chaosControlKey, Owner: me.ID, Controller: me.ID})
	})
	return g, me, opp, src
}

// castPaidWith casts a {1}{R} instant for `p`, paid strictly from two
// red mana whose sources were `kinds` (one token each).
func castPaidWith(t *testing.T, g *Game, p *Player, kinds ...ManaSourceKinds) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		for _, k := range kinds {
			p.ManaPool.AddMana(ManaToken{Color: "R", Source: uuid.New(), SourceKinds: k})
		}
	})
	id := pushTypedCardToHandWithCost(p, "Test Shock", "Instant", "{1}{R}")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return id
}

func grantedSplitSecond(g *Game, spell uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.SpellHasSplitSecondForEffect(spell) })
	return out
}

// Mana from an artifact, even one of the two, gives the spell split
// second; the table is shut to a response, and the grant is not stamped
// on the item.
func TestSplitSecondGrantReadsArtifactManaSpent(t *testing.T) {
	g, me, opp, _ := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceArtifact, ManaSourceLand)
	if !g.SplitSecondActive || !grantedSplitSecond(g, spell) {
		t.Fatalf("a spell paid partly with artifact mana has no split second (table=%v)", g.SplitSecondActive)
	}
	if g.StackMeta[spell].SplitSecond {
		t.Error("the grant was stamped on the stack item: it must be read live")
	}
	bolt := pushTypedCardToHand(opp, "Lightning Bolt", "Instant")
	if err := g.CastSpell(opp.ID, bolt, CastSpellParams{}); !errors.Is(err, ErrSplitSecondActive) {
		t.Errorf("a response to the spell: got %v, want ErrSplitSecondActive", err)
	}
}

// A Treasure is an artifact: its mana sets both kinds, and counts.
func TestSplitSecondGrantCountsTreasureMana(t *testing.T) {
	g, me, _, _ := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceTreasure|ManaSourceArtifact, ManaSourceLand)
	if !grantedSplitSecond(g, spell) {
		t.Fatal("a spell paid with Treasure mana has no split second")
	}
}

// No artifact mana, no split second.
func TestSplitSecondGrantIgnoresANonArtifactPayment(t *testing.T) {
	g, me, _, _ := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceLand, ManaSourceLand|ManaSourceCreature)
	if g.SplitSecondActive || grantedSplitSecond(g, spell) {
		t.Fatal("a spell paid with land and creature mana has split second")
	}
}

// "Each spell YOU cast": an opponent's artifact-paid spell is untouched.
func TestSplitSecondGrantIgnoresAnOpponentsCast(t *testing.T) {
	g, _, opp, _ := chaosControlGame(t)
	spell := pushTypedCardToHandWithCost(opp, "Their Shock", "Instant", "{R}")
	g.WithWriteLock(func() {
		opp.ManaPool.AddMana(ManaToken{Color: "R", Source: uuid.New(), SourceKinds: ManaSourceArtifact})
	})
	if err := g.CastSpell(opp.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the opponent's cast: %v", err)
	}
	if g.SplitSecondActive || grantedSplitSecond(g, spell) {
		t.Fatal("the grant reached a spell its controller did not cast")
	}
}

// Still live (CR 611.3a): the source leaving takes split second off the
// spell.
func TestSplitSecondGrantEndsWhenTheSourceLeaves(t *testing.T) {
	g, me, _, src := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceArtifact, ManaSourceArtifact)
	if !g.SplitSecondActive {
		t.Fatal("setup: no split second")
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(src); err != nil {
			t.Fatalf("sacrifice the source: %v", err)
		}
	})
	if g.SplitSecondActive || grantedSplitSecond(g, spell) {
		t.Fatal("split second outlived the permanent that grants it")
	}
}

// "Each spell you CAST" reads the caster: a spell of yours an opponent
// has taken control of is still one you cast. A grant on "spells you
// control" (YouCast false) would have dropped it.
func TestSplitSecondGrantFollowsTheCasterNotTheController(t *testing.T) {
	g, me, opp, _ := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceArtifact, ManaSourceArtifact)
	g.WithWriteLock(func() {
		if !g.GainControlOfSpellForEffect(uuid.New(), spell, opp.ID, "test — take the spell") {
			t.Fatal("setup: the spell did not change hands")
		}
	})
	if g.StackMeta[spell].Controller != opp.ID {
		t.Fatal("setup: the opponent does not control the spell")
	}
	if !grantedSplitSecond(g, spell) {
		t.Error("a spell you cast lost the grant when an opponent took control of it")
	}
}

// A copy was not cast and spent nothing (CR 707.10): no split second for
// it, though the original keeps its own.
func TestSplitSecondGrantSkipsACopy(t *testing.T) {
	g, me, _, _ := chaosControlGame(t)
	spell := castPaidWith(t, g, me, ManaSourceArtifact, ManaSourceArtifact)
	var copyID uuid.UUID
	g.WithWriteLock(func() {
		id, err := g.CopySpellKeepingTargetsForEffect(spell, me.ID)
		if err != nil {
			t.Fatalf("copy: %v", err)
		}
		copyID = id
	})
	if grantedSplitSecond(g, copyID) {
		t.Error("the copy has split second")
	}
	if !grantedSplitSecond(g, spell) {
		t.Error("the original lost split second when it was copied")
	}
}
