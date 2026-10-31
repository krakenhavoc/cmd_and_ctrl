package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shadow_the_hedgehog_test.go — Shadow the Hedgehog (#4724): the dies
// trigger read as each creature last existed (CR 603.10a), and Chaos
// Control, a split-second grant read off the mana spent to cast the
// spell (#1212) and live off the battlefield (CR 611.3a).

const (
	shadowOracle        = "0c12abcc-6ada-43b3-8750-9cb6ce2b0785"
	shadowDamnationOrcl = "d57a8f0b-7989-4db5-8756-6f2690097252"
)

func pushShadow(g *game.Game, controller uuid.UUID) uuid.UUID {
	return b11Push(g, controller, "Shadow the Hedgehog", "Legendary Creature — Hedgehog Mercenary", shadowOracle, 4, 2)
}

func TestShadowHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	shadow := pushShadow(g, g.Seats[0].ID)
	if !effectiveAbilitiesContain(t, g, shadow, "haste") {
		t.Error("Shadow the Hedgehog lacks haste")
	}
}

// Which deaths draw: Shadow's own, and another creature you control that
// had flash or haste as it died. A plain creature of yours and an
// opponent's hasty creature do not.
func TestShadowDrawsForItselfAndForFlashOrHasteCreatures(t *testing.T) {
	for _, tc := range []struct {
		name  string
		seed  func(g *game.Game, me, opp *game.Player, shadow uuid.UUID) uuid.UUID
		draws int
	}{
		{"Shadow itself", func(_ *game.Game, _, _ *game.Player, shadow uuid.UUID) uuid.UUID { return shadow }, 1},
		{"a haste creature you control", func(g *game.Game, me, _ *game.Player, _ uuid.UUID) uuid.UUID {
			return pushKeywordCreature(g, me.ID, "Hasty", "haste")
		}, 1},
		{"a flash creature you control", func(g *game.Game, me, _ *game.Player, _ uuid.UUID) uuid.UUID {
			return pushKeywordCreature(g, me.ID, "Flashy", "flash")
		}, 1},
		{"a plain creature you control", func(g *game.Game, me, _ *game.Player, _ uuid.UUID) uuid.UUID {
			return pushKeywordCreature(g, me.ID, "Plain")
		}, 0},
		{"an opponent's haste creature", func(g *game.Game, _, opp *game.Player, _ uuid.UUID) uuid.UUID {
			return pushKeywordCreature(g, opp.ID, "Their Hasty", "haste")
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			shadow := pushShadow(g, me.ID)
			victim := tc.seed(g, me, opp, shadow)
			hand := me.Hand.Size()

			g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(victim) })
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
		})
	}
}

// A keyword is read as the creature last existed on the battlefield
// (CR 603.10a): haste an effect gave it counts.
func TestShadowReadsGrantedHasteAsItLastExisted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushShadow(g, me.ID)
	bear := pushKeywordCreature(g, me.ID, "Bear")
	g.WithWriteLock(func() {
		c, _ := battlefieldCard(g, bear)
		g.RegisterScopedEffectForEffect(bear,
			[]game.AffectedObject{game.PinObject(bear, c.EnteredBattlefieldAt)},
			[]game.Mod{game.AddKeywordsMod("haste")}, g.UntilEndOfTurnDuration(), "test — gains haste")
	})
	if !effectiveAbilitiesContain(t, g, bear, "haste") {
		t.Fatal("setup: the Bear has no haste")
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d for a creature with granted haste, want 1", got)
	}
}

// Leaves-the-battlefield triggers look back in time: Shadow dying in the
// same wipe as a hasty creature draws for both.
func TestShadowSeesCreaturesDyingBesideIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushShadow(g, me.ID)
	pushKeywordCreature(g, me.ID, "Hasty", "haste")
	pushKeywordCreature(g, me.ID, "Plain")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Damnation", "Sorcery", shadowDamnationOrcl, nil)
	passPriorityAroundTable(t, g)
	// Shadow and the hasty creature draw; the plain one does not.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand changed by %d after the wipe, want +2", got)
	}
}

// shadowCastShock casts a {1}{R} instant for `me`, paid with one mana
// from a cracked Treasure (an artifact) and one from elsewhere, or with
// two non-artifact mana.
func shadowCastShock(t *testing.T, g *game.Game, me *game.Player, treasure bool) uuid.UUID {
	t.Helper()
	if treasure {
		crackATreasureFor(t, g, me, "R")
		floatForTest(g, me, "R")
	} else {
		floatForTest(g, me, "RR")
	}
	return castFromHandForTest(t, g, me, "Test Shock", "Instant", "{1}{R}", "", game.CastSpellParams{Strict: true})
}

// Chaos Control: a spell paid partly with artifact mana has split second,
// and the table cannot respond (but can still make mana, CR 702.61b).
func TestShadowChaosControlGivesArtifactPaidSpellsSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushShadow(g, me.ID)
	advanceToMain(t, g)

	spell := shadowCastShock(t, g, me, true)
	if !g.SplitSecondActive || !hasSplitSecond(g, spell) {
		t.Fatalf("a Treasure-paid spell has no split second (table=%v)", g.SplitSecondActive)
	}
	if err := castBoltAt(t, g, opp, me.ID); !errors.Is(err, game.ErrSplitSecondActive) {
		t.Errorf("a response: got %v, want ErrSplitSecondActive", err)
	}
	passPriorityAroundTable(t, g)
	if g.SplitSecondActive {
		t.Error("split second outlived the spell")
	}
}

// No artifact mana, no split second.
func TestShadowChaosControlIgnoresOtherMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushShadow(g, me.ID)
	advanceToMain(t, g)

	spell := shadowCastShock(t, g, me, false)
	if g.SplitSecondActive || hasSplitSecond(g, spell) {
		t.Fatal("a spell paid without artifact mana has split second")
	}
}

// "Each spell YOU cast": an opponent's artifact-paid spell is untouched.
func TestShadowChaosControlIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushShadow(g, opp.ID)
	advanceToMain(t, g)

	spell := shadowCastShock(t, g, me, true)
	if g.SplitSecondActive || hasSplitSecond(g, spell) {
		t.Fatal("an opponent's Shadow gave my spell split second")
	}
}

// The grant is live (CR 611.3a): Shadow leaving takes split second off a
// spell already on the stack.
func TestShadowLeavingTakesSplitSecondOffTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shadow := pushShadow(g, me.ID)
	advanceToMain(t, g)

	spell := shadowCastShock(t, g, me, true)
	if !g.SplitSecondActive {
		t.Fatal("setup: no split second")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(shadow) })
	if g.SplitSecondActive || hasSplitSecond(g, spell) {
		t.Fatal("split second stayed on the spell after Shadow left")
	}
	if err := castBoltAt(t, g, opp, me.ID); err != nil {
		t.Errorf("a response after Shadow left: %v", err)
	}
}
