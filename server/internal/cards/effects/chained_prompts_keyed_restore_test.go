package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chained_prompts_keyed_restore_test.go — #4301: the prompts that lead into
// the keyed confirms of #4298 are keyed too, so a table is a restore point
// from the first prompt of Ponder, Portent, Sylvan Library and Lazav,
// Familiar Stranger (the London-mulligan bottom pick is tested in
// protocol/mulligan_bottom_view_test.go, where the views are). Each test
// takes a STRICT restore (restoreRoundTrip fails on a table that is not a
// restore point) while the prompt is open, checks the key the snapshot
// names, then answers it on the restored table and carries on to the end of
// the card.

// promptKeyOf is the key the snapshot writes for chooser's open prompt of
// the kind, "" when there is none.
func promptKeyOf(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) string {
	for _, c := range g.CaptureSnapshot().PendingChoices {
		if c.Kind == kind && c.Chooser == chooser {
			return c.PickThen
		}
	}
	return ""
}

// refuseUnknownPromptKey is refuseUnknownKey for a prompt of the kind: the
// snapshot names a key this binary does not register, and Restore refuses it.
func refuseUnknownPromptKey(t *testing.T, g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if c := &snap.PendingChoices[i]; c.Kind == kind && c.Chooser == chooser && c.PickThen != "" {
			c.PickThen = "never-registered/not-a-key"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no keyed %s prompt for %s", kind, chooser)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// Ponder: the look-at-top prompt, whose "then" asks the shuffle question.
// Restored at the look, then at the question, the draw is still on both
// branches and a kept order draws the card put on top.
func TestPonderLookIsARestorePoint(t *testing.T) {
	for _, shuffle := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedLibrary(me, "Worst", "Middling", "Best", "L4", "L5", "L6", "L7", "L8")
		castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
		passPriorityAroundTable(t, g)
		c := lookAtTopChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("Ponder queued no look-at prompt")
		}
		if key := promptKeyOf(g, game.PendingChoiceLookAtTop, me.ID); key != ponderAskShuffle.Key() {
			t.Fatalf("the look-at prompt is written under %q, want %q", key, ponderAskShuffle.Key())
		}
		seen := append([]uuid.UUID(nil), c.ScryCards...)

		g = restoreRoundTrip(t, g, true)
		me = seatOf(g, me.ID)
		c = lookAtTopChoiceFor(g, me.ID)
		if c == nil || len(c.ScryCards) != 3 || c.ScryCards[0] != seen[0] || c.ScryCards[2] != seen[2] {
			t.Fatalf("the restored look-at prompt is %+v, want the same three cards %v", c, seen)
		}
		if err := g.ResolveLookAtTop(c.ID, me.ID, []uuid.UUID{seen[2], seen[0], seen[1]}); err != nil {
			t.Fatal(err)
		}
		g, ask := restoreAtConfirm(t, g, me.ID, ponderShuffle.Key())
		me = seatOf(g, me.ID)
		hand := me.Hand.Size()
		if err := g.ResolveConfirm(ask.ID, me.ID, shuffle); err != nil {
			t.Fatal(err)
		}
		if me.Hand.Size() != hand+1 {
			t.Fatalf("shuffle=%v: hand %d → %d after the look was restored, want the draw on both branches", shuffle, hand, me.Hand.Size())
		}
		if drawn, _ := me.Hand.Top(); !shuffle && drawn.InstanceID != seen[2] {
			t.Errorf("kept the order but drew %q, want the card put on top", drawn.Name)
		}
	}
}

func TestPonderLookWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
	passPriorityAroundTable(t, g)
	refuseUnknownPromptKey(t, g, game.PendingChoiceLookAtTop, me.ID)
}

// A look-at-top with no "then" (Sensei's Divining Top, Soothsaying) holds
// no continuation at all, so it was always a restore point; it still is.
func TestLookAtTopWithoutAThenStaysARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E")
	g.WithWriteLock(func() {
		if err := (LookAtTop{Player: me.ID, N: 3}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	g = restoreRoundTrip(t, g, true)
	c := lookAtTopChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 3 {
		t.Fatalf("the restored look-at prompt is %+v", c)
	}
	if err := g.ResolveLookAtTop(c.ID, me.ID, c.ScryCards); err != nil {
		t.Fatal(err)
	}
}

// Portent: the put-in-library prompt over ANOTHER player's top three, whose
// "then" asks whether that player shuffles. The library's owner rides the
// tail, so the question names them on the restored table.
func TestPortentLookIsARestorePoint(t *testing.T) {
	for _, shuffle := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		lib := seedLibrary(opp, "A", "B", "C", "D")
		castCatalogSpell(t, g, "Portent", "Sorcery", portentOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		passPriorityAroundTable(t, g)
		if putInLibraryChoiceFor(g, me.ID) == nil {
			t.Fatal("no look prompt")
		}
		if key := promptKeyOf(g, game.PendingChoicePutInLibrary, me.ID); key != portentAskShuffle.Key() {
			t.Fatalf("the order prompt is written under %q, want %q", key, portentAskShuffle.Key())
		}

		g = restoreRoundTrip(t, g, true)
		me = seatOf(g, me.ID)
		ask := putInLibraryChoiceFor(g, me.ID)
		if ask == nil || len(ask.ScryCards) != 3 || ask.LibraryPlacement != game.LibraryPlaceTop {
			t.Fatalf("the restored order prompt is %+v, want the three cards to put back on top", ask)
		}
		if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{lib[2], lib[0], lib[1]}); err != nil {
			t.Fatal(err)
		}
		g, q := restoreAtConfirm(t, g, me.ID, portentShuffle.Key())
		if len(q.OptionCarry) != 2 || q.OptionCarry[1] != opp.ID {
			t.Errorf("the question carries %v, want the caster and the library's owner", q.OptionCarry)
		}
		if err := g.ResolveConfirm(q.ID, me.ID, shuffle); err != nil {
			t.Fatal(err)
		}
		for _, id := range lib[:3] {
			if knownTo(g, id, me.ID) == shuffle {
				t.Errorf("shuffle=%v: the caster still knows a card of the library: %v", shuffle, !shuffle)
			}
		}
		top := seatOf(g, opp.ID).Library.Cards
		if !shuffle && (top[len(top)-1].InstanceID != lib[2] || top[len(top)-2].InstanceID != lib[0]) {
			t.Error("kept the order but the library is not as the restored prompt was answered")
		}
	}
}

func TestPortentLookWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(opp, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Portent", "Sorcery", portentOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	refuseUnknownPromptKey(t, g, game.PendingChoicePutInLibrary, me.ID)
}

// Sylvan Library: the pick of the two cards drawn this turn. It is restored
// with its candidates, and the payments after it run on the restored table.
func TestSylvanLibraryPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)
	if key := promptKeyOf(g, game.PendingChoiceChooseCards, owner.ID); key != sylvanLibraryChosen.Key() {
		t.Fatalf("the pick is written under %q, want %q", key, sylvanLibraryChosen.Key())
	}
	candidates := append([]uuid.UUID(nil), pick.ChooseCards...)
	life := owner.Life

	g = restoreRoundTrip(t, g, true)
	owner = seatOf(g, owner.ID)
	pick = chooseCardsChoiceFor(g, owner.ID)
	if pick == nil || len(pick.ChooseCards) != len(candidates) || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("the restored pick is %+v, want two of %v", pick, candidates)
	}
	chosen := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	if err := g.ResolveChooseCards(pick.ID, owner.ID, chosen); err != nil {
		t.Fatal(err)
	}
	g, first := restoreAtConfirm(t, g, owner.ID, sylvanLibraryAnswer.Key())
	owner = seatOf(g, owner.ID)
	if err := g.ResolveConfirm(first.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	g, second := restoreAtConfirm(t, g, owner.ID, sylvanLibraryAnswer.Key())
	owner = seatOf(g, owner.ID)
	if err := g.ResolveConfirm(second.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if owner.Life != life-4 || !owner.Hand.Contains(chosen[0]) || owner.Hand.Contains(chosen[1]) || len(g.PendingChoices) != 0 {
		t.Errorf("life %d → %d, kept %v, put back %v, %d prompts open; want one paid, one on top, finished",
			life, owner.Life, owner.Hand.Contains(chosen[0]), !owner.Hand.Contains(chosen[1]), len(g.PendingChoices))
	}
}

// The pick is still refused a card that is not among the cards drawn this
// turn once restored, and one that has left the hand since the prompt was
// asked (LiveZone), as the closure form's Zone did.
func TestSylvanLibraryRestoredPickKeepsItsRules(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)
	candidates := append([]uuid.UUID(nil), pick.ChooseCards...)
	g = restoreRoundTrip(t, g, true)
	owner = seatOf(g, owner.ID)
	pick = chooseCardsChoiceFor(g, owner.ID)
	if err := g.ResolveChooseCards(pick.ID, owner.ID, []uuid.UUID{candidates[0], uuid.New()}); err == nil {
		t.Error("the restored pick took a card that is not one of the candidates")
	}
	if err := g.ResolveChooseCards(pick.ID, owner.ID, candidates[:1]); err == nil {
		t.Error("the restored pick took one card, want exactly two")
	}
	// A candidate that has left the hand (here, back on top of the library)
	// is no longer a legal pick.
	gone := candidates[0]
	g.WithWriteLock(func() {
		c, err := owner.Hand.Remove(gone)
		if err != nil {
			t.Fatal(err)
		}
		owner.Library.PushTop(c)
	})
	if err := g.ResolveChooseCards(pick.ID, owner.ID, []uuid.UUID{gone, candidates[1]}); err == nil {
		t.Error("the restored pick took a card that has left the hand")
	}
}

func TestSylvanLibraryPickWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	owner, _ := sylvanOpening(t, g)
	refuseUnknownPromptKey(t, g, game.PendingChoiceChooseCards, owner.ID)
}

// Lazav, Familiar Stranger: the pick of a card from a graveyard, then the
// copy offer. Restored at the pick, the exile happens on the restored table
// and Lazav becomes the creature.
func TestLazavFamiliarStrangerPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lazav := dcSeed(g, me.ID, "Lazav, Familiar Stranger", "Legendary Creature — Shapeshifter", oracleLazavFamiliarStranger, 1, 4)
	giant := dcGraveyardCard(g, opp, "Hill Giant", "Creature — Giant", "oracle-giant", 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatal("Lazav must offer the graveyard exile")
	}
	if key := promptKeyOf(g, game.PendingChoiceChooseCards, me.ID); key != lazavFamiliarStrangerExile.Key() {
		t.Fatalf("the pick is written under %q, want %q", key, lazavFamiliarStrangerExile.Key())
	}

	g = restoreRoundTrip(t, g, true)
	me = seatOf(g, me.ID)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("the restored pick is %+v, want up to one card", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{giant}); err != nil {
		t.Fatal(err)
	}
	g, ask := restoreAtConfirm(t, g, me.ID, lazavFamiliarStrangerCopy.Key())
	if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != giant {
		t.Fatalf("the offer carries %v, want the exiled card", ask.OptionCarry)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if c := dcCard(t, g, lazav); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Errorf("Lazav did not become the Hill Giant: %q %d/%d", c.Name, c.Effective().Power, c.Effective().Toughness)
	}
}

// Declining the exile on the restored table does nothing and asks nothing.
func TestLazavFamiliarStrangerDeclinedPickAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dcSeed(g, me.ID, "Lazav, Familiar Stranger", "Legendary Creature — Shapeshifter", oracleLazavFamiliarStranger, 1, 4)
	giant := dcGraveyardCard(g, opp, "Hill Giant", "Creature — Giant", "oracle-giant", 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = seatOf(g, me.ID)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick after the restore")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !seatOf(g, opp.ID).Graveyard.Contains(giant) || latestConfirmFor(g, me.ID) != nil {
		t.Error("declining exiled the card or asked about a copy")
	}
}

func TestLazavFamiliarStrangerPickWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dcSeed(g, me.ID, "Lazav, Familiar Stranger", "Legendary Creature — Shapeshifter", oracleLazavFamiliarStranger, 1, 4)
	dcGraveyardCard(g, opp, "Hill Giant", "Creature — Giant", "oracle-giant", 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	refuseUnknownPromptKey(t, g, game.PendingChoiceChooseCards, me.ID)
}
