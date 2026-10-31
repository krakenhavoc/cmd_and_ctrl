package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_cast_keyed_restore_test.go — #4246 made the free-cast offers keyed
// choose-cards picks, so a table waiting on one is a restore point. #4250
// moved the offer inside the resolution, onto the resolution_cast prompt,
// which is plain data too: each test takes a STRICT restore
// (restoreRoundTrip with strict, which fails on a table that is not a
// restore point) while the offer is open, then answers it on the restored
// table. The keyed picks stay registered, so a table saved while waiting on
// one (before #4250) still restores and answers; those tests build the
// pick outside a resolution, the one place it is still asked.

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
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, rot)
	if len(offer.DiscardOptions) != 1 {
		t.Fatalf("the free-cast offer is %v, want Mind Rot alone", offer.DiscardOptions)
	}
	return g, rot
}

// Cecily's offer, restored strictly while it is open: the restored table asks
// the same question inside the same held resolution, and the free cast is
// made in the declare attackers step (CR 608.2g). Declining on a restored
// table opens nothing.
func TestCecilyFreeCastOfferIsARestorePoint(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g, rot := cecilyOfferTable(t)
		g = restoreRoundTrip(t, g, true)
		me, opp := g.Seats[0], g.Seats[1]
		mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, rot)
		perm := g.CastPermissionOnCardByIDForEffect(rot)
		if perm == nil || perm.Cost != "{0}" || perm.Player != me.ID || perm.SourceName != "Cecily, Haunted Mage" {
			t.Fatalf("permission on the restored table: %+v", perm)
		}
		if !cast {
			declineResolutionCast(t, g, me.ID)
			if perm := g.CastPermissionOnCardByIDForEffect(rot); perm != nil {
				t.Fatalf("the restored decline left a cast open: %+v", perm)
			}
			if !me.Hand.Contains(rot) || g.ResolutionPaused() {
				t.Fatal("the restored decline moved the card or held the resolution")
			}
			continue
		}
		if err := castThroughResolutionCast(t, g, me.ID, rot, game.CastSpellParams{
			Strict:  true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("the free cast on the restored table: %v", err)
		}
		if me.Hand.Contains(rot) || !g.Stack.Contains(rot) {
			t.Fatal("Mind Rot did not go from the hand to the stack")
		}
		castInsideTheResolution(t, g, rot)
	}
}

// A keyed pick from before #4250 (built here outside a resolution, where it
// is still the offer) restores and answers: the key is still registered.
// A binary that does not know the key refuses the file rather than
// restoring a question whose answer would run nothing.
func TestCecilyFreeCastOfferWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	rot := handCardFull(me, "Mind Rot", "Sorcery", "{2}{B}", mindRotOracle, []string{"B"})
	offerFrom(t, g, me, func(ctx *Context) error {
		offerFreeHandCast(ctx, me.ID, cecilyFreeCastOffer, "an instant or sorcery spell", freeInstantOrSorcery())
		return nil
	})
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); pick == nil || pick.PickThen != cecilyFreeCastOffer.then.Key() {
		t.Fatalf("outside a resolution the offer is the keyed pick, got %+v", pick)
	}
	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[0]
	pick := latestChoiceOfKindFor(restored, game.PendingChoiceChooseCards, rme.ID)
	if pick == nil {
		t.Fatal("the restored table lost the keyed pick")
	}
	if err := restored.ResolveChooseCards(pick.ID, rme.ID, []uuid.UUID{rot}); err != nil {
		t.Fatalf("the restored pick: %v", err)
	}
	if perm := restored.CastPermissionOnCardByIDForEffect(rot); perm == nil || perm.Cost != "{0}" {
		t.Fatalf("the restored pick opened no free cast: %+v", perm)
	}

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
// instant is cast free with its target on the restored table.
func TestBusterSwordFreeCastOfferIsARestorePoint(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.g = restoreRoundTrip(t, tb.g, true)
	tb.me, tb.opp = tb.g.Seats[0], tb.g.Seats[1]
	tb.offer(t)
	if perm := tb.permissionOn(tb.bolt); perm == nil || perm.Cost != "{0}" || perm.SourceName != "Buster Sword" {
		t.Fatalf("permission on the restored table: %+v", perm)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bolt := handCardFull(me, "Lightning Bolt", "Instant", "{R}", boltOracleCombat, []string{"R"})
	offerFrom(t, g, me, func(ctx *Context) error { return busterSwordOfferFreeCast(ctx, me.ID, 2) })
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no keyed pick outside a resolution")
	}
	g.WithWriteLock(func() {
		for i, c := range me.Hand.Cards {
			if c.InstanceID == bolt {
				opp.Hand.Cards = append(opp.Hand.Cards, c)
				me.Hand.Cards = append(me.Hand.Cards[:i], me.Hand.Cards[i+1:]...)
				break
			}
		}
	})
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if perm := g.CastPermissionOnCardByIDForEffect(bolt); perm != nil {
		t.Fatalf("a card that left the chooser's hand was granted a free cast: %+v", perm)
	}
}

// Aetherworks Marvel's look-and-cast offer, restored strictly: the looked-at
// cards are the prompt's carried data, so the restored cast still sends the
// rest to the bottom once it is made.
func TestAetherworksMarvelFreeCastPickIsARestorePoint(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	spells, lands, deep := aetherworksLibrary(me)
	aetherworksLook(t, g, me, marvel)
	g = restoreRoundTrip(t, g, true)
	me = g.PlayerByIDForEffect(me.ID)
	c := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spells...)
	if c.PickThen != lookedAtRestToBottom.Key() || len(c.OptionCarry) != 6 {
		t.Fatalf("the restored offer is %+v, want the look's continuation and its six cards", c)
	}
	wurm := spells[1]
	if err := castThroughResolutionCast(t, g, me.ID, wurm, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the restored cast: %v", err)
	}
	if !g.Stack.Contains(wurm) {
		t.Fatal("the restored cast did not reach the stack")
	}
	for _, id := range append(append([]uuid.UUID{}, lands...), spells[0], spells[2], spells[3]) {
		if i := libraryIndex(me, id); i < 0 || i > 4 {
			t.Errorf("the rest go to the bottom (indices 0-4), found %d", i)
		}
	}
	if i := libraryIndex(me, deep); i != len(me.Library.Cards)-1 {
		t.Errorf("the deep card is the top card once the rest go under, found %d", i)
	}
}

// A resolution_cast's continuation is a key: a binary that does not know it
// refuses the file rather than restoring a prompt whose answer would skip
// the rest of the resolution.
func TestResolutionCastOfferWithAnUnknownContinuationIsRefused(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	aetherworksLibrary(me)
	aetherworksLook(t, g, me, marvel)
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].Kind == game.PendingChoiceResolutionCast && snap.PendingChoices[i].PickThen != "" {
			snap.PendingChoices[i].PickThen = "resolution-cast/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no resolution_cast naming a continuation")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown continuation: err = %v, want ErrUnknownEffectKey", err)
	}
}
