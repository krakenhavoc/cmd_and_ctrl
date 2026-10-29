package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// simultaneous_replacements_test.go — #4521: a replacement whose source
// is destroyed in the same wipe still replaces the deaths of everything
// destroyed with it (CR 608.2f, CR 614.4). Liesa is pushed FIRST in each
// test, so her leg of the batch moves first: the position that used to
// take her effect away from every later leg.

// liesaOptionID is the ID of Liesa's effect on an open ordering prompt,
// read off the prompt's labels as the client reads them.
func liesaOptionID(t *testing.T, g *game.Game, order *game.PendingChoice, liesa uuid.UUID) game.ReplacementEffectID {
	t.Helper()
	for _, id := range order.ReplacementEffectIDs {
		var label string
		var src uuid.UUID
		g.ReadSnapshot(func() { label, src = g.ReplacementOptionMetaForEffect(id) })
		if src == liesa {
			if label != "Liesa, Forgotten Archangel: exile instead of dying" {
				t.Errorf("Liesa's option is labelled %q", label)
			}
			return id
		}
	}
	t.Fatalf("no option on the prompt names Liesa: %v", order.ReplacementEffectIDs)
	return 0
}

// orderWith puts first at the front of the prompt's IDs.
func orderWith(order *game.PendingChoice, first game.ReplacementEffectID) []game.ReplacementEffectID {
	out := []game.ReplacementEffectID{first}
	for _, id := range order.ReplacementEffectIDs {
		if id != first {
			out = append(out, id)
		}
	}
	return out
}

// Wrath of God over Liesa and three of an opponent's creatures: Liesa
// goes to her owner's graveyard (she is not an opponent's creature), and
// every one of theirs is exiled, including the ones destroyed after her.
func TestWrathOfGodWithLiesaExilesEveryOpponentCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	theirs := []uuid.UUID{
		b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2),
		b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1),
		b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4),
	}
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(liesa) {
		t.Error("Liesa is not an opponent's creature: she dies")
	}
	for i, id := range theirs {
		if !g.Exile.Contains(id) || opp.Graveyard.Contains(id) {
			t.Errorf("their creature %d went to the graveyard: Liesa's effect existed when the wipe began", i)
		}
	}
}

// The issue's own board: Wrath of God over Head of the Hunt and three
// opponents' creatures. Head of the Hunt is not an opponent's creature,
// so it dies; each of theirs is exiled and makes a Wolf (CR 603.12),
// including the ones destroyed after it left.
func TestWrathOfGodWithHeadOfTheHuntExilesAndMakesAWolfForEach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	head := seedHeadOfTheHunt(g, me.ID)
	theirs := []uuid.UUID{
		pushVanillaCreature(g, g.Seats[1].ID, "Bear A", 2, 2),
		pushVanillaCreature(g, g.Seats[2].ID, "Bear B", 2, 2),
		pushVanillaCreature(g, g.Seats[3].ID, "Bear C", 2, 2),
	}
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(head) {
		t.Error("Head of the Hunt is not an opponent's creature: it dies")
	}
	for i, id := range theirs {
		if !g.Exile.Contains(id) {
			t.Errorf("their creature %d is not in exile: Head of the Hunt's effect existed when the wipe began", i)
		}
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 3 {
		t.Errorf("%d Wolves, want one per creature exiled", got)
	}
}

// Two Liesas under two players, wiped together: each one's effect exiles
// the other, whichever leaves first. Ordering the legs cannot do this,
// since each source must outlast the other.
func TestTwoLiesasInOneWrathExileEachOther(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	theirs := b12Push(g, opp.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(mine) {
		t.Error("my Liesa was destroyed while theirs still applied: exiled")
	}
	if !g.Exile.Contains(theirs) {
		t.Error("their Liesa was destroyed while mine still applied: exiled")
	}
}

// A single-target destroy is unchanged, and nothing leaks out of a
// finished wipe: once Liesa has gone, a later destruction is a later
// event and her effect does not apply to it.
func TestLiesaAppliesOnlyWhileSheIsThereOutsideABatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	first := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	if !g.Exile.Contains(first) {
		t.Fatal("a single destroy while Liesa is there: exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(liesa) })
	if !me.Graveyard.Contains(liesa) {
		t.Fatal("Liesa destroyed alone dies")
	}
	later := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(later) })
	if !opp.Graveyard.Contains(later) || g.Exile.Contains(later) {
		t.Error("Liesa had already left: a later destruction is not replaced")
	}

	// And after a wipe that took her.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	b12Push(g2, me2.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	castCatalogSpell(t, g2, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g2)
	after := b12Creature(g2, opp2.ID, "Elf", "Creature — Elf", 1, 1)
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(after) })
	if !opp2.Graveyard.Contains(after) {
		t.Error("the wipe is over: its departed Liesa replaces nothing after it")
	}
}

// liesaWipeOrder casts the wipe over Liesa and an opponent's Needle, with
// one Graveyard Order Probe of theirs on the battlefield, and returns the
// Needle's CR 616 prompt: Liesa's departed effect against the probe. The
// probe replaces the move with itself, so the Needle can be exiled only
// by Liesa's effect.
func liesaWipeOrder(t *testing.T, wipe, oracle string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, *game.PendingChoice) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedReplacementPermanent(g, graveyardOrderProbeOracles[0], "Graveyard Order Probe A", opp.ID)
	castCatalogSpell(t, g, wipe, "Sorcery", oracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(liesa) {
		t.Fatal("Liesa's leg moved first and she died")
	}
	order := b06ReplacementOrderFor(g, opp.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want the Needle's controller asked to order Liesa and the probe (CR 616.1), got %+v", g.PendingChoices)
	}
	return g, opp, liesa, needle, order
}

// The order is still the affected creature's controller's to choose
// (CR 616.1), Liesa's departed effect is one of the choices, and it
// exiles the Needle after the pause in either order.
func TestADepartedLiesaIsOrderedByTheAffectedCreaturesController(t *testing.T) {
	for _, liesaFirst := range []bool{true, false} {
		g, opp, liesa, needle, order := liesaWipeOrder(t, "Wrath of God", wrathOfGodOracle)
		first := liesaOptionID(t, g, order, liesa)
		if !liesaFirst {
			for _, id := range order.ReplacementEffectIDs {
				if id != first {
					first = id
					break
				}
			}
		}
		if err := g.ResolveReplacementOrder(order.ID, opp.ID, orderWith(order, first)); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		if !g.Exile.Contains(needle) || opp.Graveyard.Contains(needle) {
			t.Errorf("Liesa first = %v: the departed Liesa did not exile the Needle after the pause", liesaFirst)
		}
	}
}

// A fire-and-forget leg holding a departed effect stays off the disk: its
// event does not carry the batch, so a restored table could not gather
// Liesa again. The keyed form below is a restore point.
func TestAFireAndForgetLegWithADepartedLiesaIsNotWritten(t *testing.T) {
	g, _, _, _, _ := liesaWipeOrder(t, "Wrath of God", wrathOfGodOracle)
	if g.CaptureSnapshot().Restorable() {
		t.Error("the prompt holds an effect a restored table could not gather")
	}
}

// Fumigate (a keyed batch) over Liesa, an opponent's Needle and their Elf,
// restored at the Needle's prompt: the restored prompt still offers
// Liesa's effect, the answer exiles the Needle, the Elf is exiled after
// it, and only Liesa was destroyed this way (CR 701.8a), so the caster
// gains 1.
func TestFumigateWithLiesaIsARestorePointMidBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	seedReplacementPermanent(g, graveyardOrderProbeOracles[0], "Graveyard Order Probe A", opp.ID)
	life := me.Life
	castCatalogSpell(t, g, "Fumigate", "Sorcery", fumigateOracle, nil)
	passPriorityAroundTable(t, g)
	if b06ReplacementOrderFor(g, opp.ID) == nil {
		t.Fatalf("want the Needle's ordering prompt, got %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me, opp = seatByID(g, me.ID), seatByID(g, opp.ID)
	if !me.Graveyard.Contains(liesa) || !g.Battlefield.Contains(needle) || !g.Battlefield.Contains(elf) {
		t.Fatal("the restored table is not paused at the Needle's leg")
	}
	order := b06ReplacementOrderFor(g, opp.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("the restored prompt lost an option: %+v", g.PendingChoices)
	}
	if err := g.ResolveReplacementOrder(order.ID, opp.ID, orderWith(order, liesaOptionID(t, g, order, liesa))); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if !g.Exile.Contains(needle) {
		t.Error("the restored answer did not fire the departed Liesa: the Needle is not exiled")
	}
	if !g.Exile.Contains(elf) || opp.Graveyard.Contains(elf) {
		t.Error("the leg after the restore lost Liesa's effect: the Elf is not exiled")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1: only Liesa was destroyed this way", life, me.Life)
	}
}

// Dauthi Voidwalker in the same Wrath as two of an opponent's creatures:
// it dies (its own controller's graveyard is not replaced), and both of
// theirs are exiled with a void counter, including the one after it left.
func TestWrathOfGodWithDauthiVoidwalkerVoidsEveryOpponentCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushCatalogPermanent(g, me.ID, "Dauthi Voidwalker", "Creature — Dauthi Rogue", dauthiOracle, false)
	theirs := []uuid.UUID{
		pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2),
		pushVanillaCreature(g, opp.ID, "Their Elf", 1, 1),
	}
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(walker) {
		t.Error("the Voidwalker goes to its own controller's graveyard")
	}
	for i, id := range theirs {
		if opp.Graveyard.Contains(id) || voidCounters(g, id) != 1 {
			t.Errorf("their creature %d: in graveyard %v, void counters %d, want exiled with 1", i, opp.Graveyard.Contains(id), voidCounters(g, id))
		}
	}
}
