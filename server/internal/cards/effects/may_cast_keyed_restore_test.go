package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_cast_keyed_restore_test.go — #4361 S5: every may_cast offer the
// catalog makes is keyed (game.RegisterMayCastThen), so a table waiting on
// one is a strict restore point and the restored answer does what the
// card says.

// restoredMayCast round-trips g strictly and returns the restored table
// and the may_cast waiting on `chooser` there.
func restoredMayCast(t *testing.T, g *game.Game, chooser uuid.UUID) (*game.Game, *game.PendingChoice) {
	t.Helper()
	if latestChoiceOfKindFor(g, game.PendingChoiceMayCast, chooser) == nil {
		t.Fatalf("setup: no may_cast is waiting: %+v", g.PendingChoices)
	}
	restored := restoreRoundTrip(t, g, true)
	c := latestChoiceOfKindFor(restored, game.PendingChoiceMayCast, chooser)
	if c == nil {
		t.Fatal("the restored table lost the offer")
	}
	return restored, c
}

// --- Malcolm, Alluring Scoundrel ---------------------------------------------

// Malcolm's free-cast offer at four chorus counters: the restored "yes"
// still grants the {0} cast over the discarded card, and the restored "no"
// grants nothing.
func TestAMalcolmFreeCastOfferIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := pushMalcolm(t, g, me.ID, 3)
		advanceToMain(t, g)
		pitch := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: pitch, Name: "Pitched Spell", TypeLine: "Instant",
			ManaCost: "{4}{U}{U}", Owner: me.ID, Controller: me.ID,
		})
		dealCombatDamageToPlayer(g, id, opp.ID, 2)
		passPriorityAroundTable(t, g)
		answerDiscard(t, g, me.ID, pitch)
		passPriorityAroundTable(t, g)

		restored, c := restoredMayCast(t, g, me.ID)
		if c.MayCastCard != pitch {
			t.Fatalf("the restored offer names %v, want the discarded card %v", c.MayCastCard, pitch)
		}
		if err := restored.ResolveMayCast(c.ID, me.ID, accept); err != nil {
			t.Fatalf("ResolveMayCast: %v", err)
		}
		var perm *game.CastPermission
		restored.ReadSnapshot(func() { perm = restored.CastPermissionOnCardByIDForEffect(pitch) })
		if !accept {
			if perm != nil {
				t.Error("a declined offer granted a cast")
			}
			continue
		}
		if perm == nil || perm.Cost != "{0}" || perm.Player != me.ID {
			t.Fatalf("the restored yes granted %+v, want a {0} cast for the controller", perm)
		}
		if err := restored.CastSpell(me.ID, pitch, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
			t.Fatalf("the granted free cast on the restored table: %v", err)
		}
	}
}

// --- Hideaway (Windbrisk Heights) ---------------------------------------------

// A hidden land's "play it?" mid-resolution: the restored yes plays it as
// this turn's land drop.
func TestAHiddenLandOfferIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "Hidden Mountain", "C", "D")
	for i := range me.Library.Cards {
		if me.Library.Cards[i].InstanceID == top[1] {
			me.Library.Cards[i].TypeLine = "Basic Land — Mountain"
			me.Library.Cards[i].ManaCost = ""
		}
	}
	heights := game.NewCard("Windbrisk Heights", me.ID)
	heights.OracleID = "3589bcfc-42b0-414a-adce-bc690dc631c8"
	heights.TypeLine = "Land"
	enterWithHideaway(t, g, heights, top[1])
	g.WithWriteLock(func() { g.LandsPlayedThisTurn[me.ID] = 0 })
	a := pushVanillaCreature(g, me.ID, "A1", 1, 1)
	b := pushVanillaCreature(g, me.ID, "A2", 1, 1)
	c := pushVanillaCreature(g, me.ID, "A3", 1, 1)
	declareAttack(t, g, opp.ID, a, b, c)
	untapForTest(g, heights.InstanceID)
	floatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, heights.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	restored, offer := restoredMayCast(t, g, me.ID)
	if offer.MayCastCard != top[1] {
		t.Fatalf("the restored offer names %v, want the hidden land %v", offer.MayCastCard, top[1])
	}
	if err := restored.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if _, ok := cardOnBattlefield(restored, top[1]); !ok {
		t.Fatal("the restored yes did not play the hidden land")
	}
	if got := restored.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("lands played = %d, want 1 (CR 305.2a)", got)
	}
}

// --- Ziatora's Envoy (free play from the top) ---------------------------------

// The Envoy's land offer: the restored yes plays the top land, and the
// restored no puts it into the hand.
func TestAZiatorasEnvoyLandOfferIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		tb := newEnvoyTable(t)
		land := tb.top("Forest", "Basic Land — Forest", "", "")
		tb.attack(t)

		restored, c := restoredMayCast(t, tb.g, tb.me.ID)
		if c.MayCastKeyword != game.MayCastKeywordPlayOrHand || c.MayCastCard != land {
			t.Fatalf("restored offer %q over %v, want %q over the land", c.MayCastKeyword, c.MayCastCard, game.MayCastKeywordPlayOrHand)
		}
		if err := restored.ResolveMayCast(c.ID, tb.me.ID, accept); err != nil {
			t.Fatalf("ResolveMayCast: %v", err)
		}
		rme := restored.PlayerByIDForEffect(tb.me.ID)
		_, played := cardOnBattlefield(restored, land)
		switch {
		case accept && (!played || rme.Hand.Contains(land)):
			t.Error("the restored yes did not play the land")
		case !accept && (played || !rme.Hand.Contains(land)):
			t.Error("the restored no did not put the land into the hand")
		}
		if rme.Library.Contains(land) {
			t.Error("the land is still in the library")
		}
	}
}
