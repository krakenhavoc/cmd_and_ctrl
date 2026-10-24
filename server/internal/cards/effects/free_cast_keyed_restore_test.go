package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_cast_keyed_restore_test.go — #4246: the free-cast offers are keyed
// choose-cards picks, so a table waiting on one is a restore point. Each
// test takes a STRICT restore (restoreRoundTrip with strict, which fails on a
// table that is not a restore point) while the offer is open, then answers it
// on the restored table.

// cecilyOfferTable is Cecily, Haunted Mage attacking with ten cards in hand,
// one of them Mind Rot, and a card on top of the library: the draw makes
// eleven, so the trigger ends on the free-cast offer.
func cecilyOfferTable(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cecily := pushCatalogPermanent(g, me.ID, "Cecily, Haunted Mage", "Legendary Creature — Human Wizard", cecilyOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	rot := handCardFull(me, "Mind Rot", "Sorcery", "{2}{B}", mindRotOracle, []string{"B"})
	for i := 1; i < 10; i++ {
		handCardFull(me, "Swamp", "Basic Land — Swamp", "", "", nil)
	}
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Drawn", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	declareAttack(t, g, opp.ID, cecily)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != rot {
		t.Fatalf("the free-cast offer is %+v, want Mind Rot alone", pick)
	}
	return g, rot
}

// Cecily's offer, restored strictly while it is open: the restored table asks
// the same question, and its answer opens the free cast, which is cast in the
// declare attackers step (CR 608.2g). Declining on a restored table opens
// nothing.
func TestCecilyFreeCastOfferIsARestorePoint(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g, rot := cecilyOfferTable(t)
		g = restoreRoundTrip(t, g, true)
		me, opp := g.Seats[0], g.Seats[1]
		pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if pick == nil {
			t.Fatal("the restored table lost the free-cast offer")
		}
		if pick.PickThen != cecilyFreeCastOffer.then.Key() || pick.ChooseDestination != game.ChooseCastFree {
			t.Fatalf("restored offer: key %q destination %q, want %q and %q",
				pick.PickThen, pick.ChooseDestination, cecilyFreeCastOffer.then.Key(), game.ChooseCastFree)
		}
		if !cast {
			if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			if perm := g.CastPermissionOnCardByIDForEffect(rot); perm != nil {
				t.Fatalf("declining opened a cast: %+v", perm)
			}
			continue
		}
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rot}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		perm := g.CastPermissionOnCardByIDForEffect(rot)
		if perm == nil || perm.Cost != "{0}" || perm.Player != me.ID || perm.SourceName != "Cecily, Haunted Mage" {
			t.Fatalf("permission after the restored answer: %+v", perm)
		}
		if err := g.CastSpell(me.ID, rot, game.CastSpellParams{
			Strict:          true,
			Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			AlternativeCost: game.AltCostKeyGrantedCost,
		}); err != nil {
			t.Fatalf("the free cast on the restored table: %v", err)
		}
		if me.Hand.Contains(rot) {
			t.Fatal("Mind Rot did not leave the hand")
		}
	}
}

// A binary that does not know the offer's key (one from before #4246, or a
// rollback) refuses the file rather than restoring a question whose answer
// would run nothing.
func TestCecilyFreeCastOfferWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _ := cecilyOfferTable(t)
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].PickThen == cecilyFreeCastOffer.then.Key() {
			snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no prompt naming the offer's key")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// Buster Sword's offer in the combat damage step, restored strictly: the
// picked instant is cast free with its target on the restored table.
func TestBusterSwordFreeCastOfferIsARestorePoint(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.g = restoreRoundTrip(t, tb.g, true)
	tb.me, tb.opp = tb.g.Seats[0], tb.g.Seats[1]
	tb.pick(t, tb.bolt)
	if perm := tb.permissionOn(tb.bolt); perm == nil || perm.Cost != "{0}" || perm.SourceName != "Buster Sword" {
		t.Fatalf("permission after the restored answer: %+v", perm)
	}
	life := tb.opp.Life
	if err := tb.castFree(tb.bolt, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID}); err != nil {
		t.Fatalf("free cast on the restored table: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if tb.opp.Life != life-3 {
		t.Fatalf("opponent life %d, want %d (the bolt resolved)", tb.opp.Life, life-3)
	}
}

// A keyed pick has no live-zone re-check of its own, so the answer checks the
// hand: a picked card that is no longer in the chooser's hand is granted
// nothing, here one that went to another player's hand before the answer.
func TestFreeHandCastPickNoLongerInTheChoosersHandIsGrantedNothing(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.g.WithWriteLock(func() {
		for i, c := range tb.me.Hand.Cards {
			if c.InstanceID == tb.bolt {
				tb.opp.Hand.Cards = append(tb.opp.Hand.Cards, c)
				tb.me.Hand.Cards = append(tb.me.Hand.Cards[:i], tb.me.Hand.Cards[i+1:]...)
				break
			}
		}
	})
	tb.pick(t, tb.bolt)
	if perm := tb.permissionOn(tb.bolt); perm != nil {
		t.Fatalf("a card that left the chooser's hand was granted a free cast: %+v", perm)
	}
}

// Aetherworks Marvel's look-and-cast pick, restored strictly: the looked-at
// cards are the prompt's carried data, so the restored answer still bottoms
// the rest and leaves the pick on top with its free cast.
func TestAetherworksMarvelFreeCastPickIsARestorePoint(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	spells, lands, deep := aetherworksLibrary(me)
	aetherworksLook(t, g, me, marvel)
	g = restoreRoundTrip(t, g, true)
	me = g.PlayerByIDForEffect(me.ID)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil || c.PickThen != aetherworksMarvelFreeCast.then.Key() {
		t.Fatalf("the restored table's pick is %+v, want the keyed look", c)
	}
	wurm := spells[1]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{wurm}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if libraryIndex(me, wurm) != len(me.Library.Cards)-1 {
		t.Errorf("the pick waits on top of the library, at %d", libraryIndex(me, wurm))
	}
	for _, id := range append(append([]uuid.UUID{}, lands...), spells[0], spells[2], spells[3]) {
		if i := libraryIndex(me, id); i < 0 || i > 4 {
			t.Errorf("the rest go to the bottom (indices 0-4), found %d", i)
		}
	}
	if i := libraryIndex(me, deep); i != len(me.Library.Cards)-2 {
		t.Errorf("the deep card sits directly under the pick, found %d", i)
	}
	if perm := g.CastPermissionOnCardByIDForEffect(wurm); perm == nil || perm.Cost != "{0}" || perm.SourceName != "Aetherworks Marvel" {
		t.Fatalf("permission over the pick after the restored answer: %+v", perm)
	}
}
