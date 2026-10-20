package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch1_test.go — #4155 (CR 608.2c): the first batch of
// hand-written bodies written as steps, one test for each way of writing
// them, each with a real card. A later instruction waits for the answer to
// the prompt an earlier one asked: a search, a scry, a dredge offer.
//
//   - ResolveSteps, on Khârn the Betrayer's question (Lightning Helix), live
//     and across a restore;
//   - ResolveSteps, on a search (Vampiric Tutor), live and on an undo;
//   - ResolveSteps, on a scry and the draw its answer runs (Read the Bones);
//   - ResolveSteps, on a dredge offer (Night's Whisper), live and on an undo;
//   - a row's Do (Sensei's Divining Top), on a dredge offer;
//   - ModeSteps (Black Market Connections), on a dredge offer.
//
// A dredge offer and a search resume through closures, so the table that
// holds one is not a restore point; the undo (a clone) stands in for the
// restore there, as in resolution_order_test.go. Khârn's question is data,
// so Lightning Helix is the restore proof.

const lightningHelixOracle = "800c258a-cfc4-4a54-a667-065ea8dea69e"

// Lightning Helix — "deals 3 damage to any target and you gain 3 life" —
// aimed at Khârn the Betrayer, whose prevention asks its controller for an
// opponent (#2040). The life gain is the next instruction, so it waits for
// the answer, live and on a table restored while the question is open.
func TestLightningHelixLifeGainWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := kharnPush(g, me.ID, me.ID)
		life := me.Life
		d1658Cast(t, g, "Lightning Helix", "Instant", lightningHelixOracle, game.CastSpellParams{
			Targets: cardRefs(kharn),
		})
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+3 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+3)
		}
		if got := kharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}

// Vampiric Tutor — "Search your library for a card, then shuffle and put
// that card on top. You lose 2 life." The life loss is the next
// instruction, so it waits for the search's answer, live and on an undo
// taken while the search is open. (The search resumes through a closure,
// so that table is not a restore point.)
func TestVampiricTutorLifeLossWaitsForTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := pushLibraryCardForTest(me, game.Card{Name: "Vamp Needle"})
	life := me.Life
	castCatalogSpell(t, g, "Vampiric Tutor", "Instant", oaVampiricTutor, nil)
	passPriorityAroundTable(t, g)
	if searchPromptFor(g, me.ID) == nil {
		t.Fatal("the search is not open")
	}
	if me.Life != life {
		t.Fatalf("life %d with the search open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerSearchByID(t, table, p.ID, needle)
		if p.Life != life-2 {
			t.Fatalf("life %d after the answer, want %d", p.Life, life-2)
		}
	}
}

// Read the Bones — "Scry 2, then draw two cards. You lose 2 life." The
// life loss waits for the scry's answer and for the draw that answer runs.
func TestReadTheBonesLifeLossWaitsForTheScry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Second", "Third")
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Read the Bones", "Sorcery", readTheBonesOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the scry is not open")
	}
	if me.Life != life {
		t.Fatalf("life %d with the scry open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		open := scryChoiceFor(table, p.ID)
		seen := open.ScryCards
		if err := table.ResolveScry(open.ID, p.ID, nil, seen); err != nil {
			t.Fatal(err)
		}
		if p.Life != life-2 {
			t.Fatalf("life %d after the answer, want %d", p.Life, life-2)
		}
		if p.Hand.Size() != hand+2 || !p.Hand.Contains(seen[0]) || !p.Hand.Contains(seen[1]) {
			t.Fatalf("hand %d after the answer, want %d holding the two cards kept on top", p.Hand.Size(), hand+2)
		}
	}
}

// Night's Whisper — "You draw two cards and you lose 2 life." With
// Stinkweed Imp in the graveyard the first draw offers dredge (CR 702.52),
// and the life loss waits for the answer, live and on an undo taken while
// the offer is open.
func TestNightsWhisperLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	castCatalogSpell(t, g, "Night's Whisper", "Sorcery", nightsWhisperOracle, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-2 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-2)
		}
	}
}

// Sensei's Divining Top — "{T}: Draw a card, then put this artifact on top
// of its owner's library." The row's Do waits: the Top stays on the
// battlefield while dredge is offered for the draw, then goes on top of the
// library once the offer is answered.
func TestSenseisTopTuckWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	top := pushCatalogPermanent(g, me.ID, "Sensei's Divining Top", "Artifact", senseisTopOracle, false)
	if err := activateOn(t, g, me.ID, top, 1, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if findBattlefieldByName(g, "Sensei's Divining Top") == uuid.Nil {
		t.Fatal("the Top left the battlefield before the draw's offer was answered")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if findBattlefieldByName(table, "Sensei's Divining Top") != uuid.Nil {
			t.Fatal("the Top is still on the battlefield after the offer was answered")
		}
		if got := libraryTopNames(p, 1); len(got) != 1 || got[0] != "Sensei's Divining Top" {
			t.Fatalf("library top %v, want the Top", got)
		}
		if !p.Hand.Contains(imp) {
			t.Fatal("dredge did not return the Imp")
		}
	}
}

// Black Market Connections' second bullet — "Draw a card. You lose 2
// life." — is a ModeSteps: the life loss waits for the dredge offer the
// draw makes, and only then does the walk move on.
func TestBlackMarketConnectionsLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedPermanentWithOracle(g, me.ID, "Black Market Connections", "Enchantment", blackMarketConnectionsOracle)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	mode := latestChoiceOfKind(g, game.PendingChoiceModePick)
	if mode == nil {
		t.Fatal("no mode prompt at the first main phase")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-2 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-2)
		}
	}
}
