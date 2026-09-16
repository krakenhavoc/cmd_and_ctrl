package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// reveal_frame_test.go — CR 701.20 reveal, from the card end.
//
// The engine and protocol halves are pinned in their own packages.
// What is tested here is the thing those two cannot see: that the
// cards which PRINT the word "reveal" actually reach the broadcast
// frame, and that they reach it with the right cards named.
//
// Three shapes, one per way a catalog card reveals today:
//
//   - a tutor's "search your library for a card, REVEAL it, put it
//     into your hand" — the largest family by far, 59 card files;
//   - "reveal cards from the top until X" where the run ends in the
//     graveyard (Consuming Aberration);
//   - the same, where one card peels off to hand (Hermit Druid).

// revealsFor returns the reveal window as the named seat receives it,
// through the same per-viewer filter a real client is served by.
func revealsFor(g *game.Game, viewer uuid.UUID) []protocol.RevealView {
	return protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String()).Reveals
}

// revealedNames flattens a reveal window into the card names it
// announced, oldest reveal first.
func revealedNames(rs []protocol.RevealView) []string {
	var out []string
	for _, r := range rs {
		for _, c := range r.Cards {
			out = append(out, c.Name)
		}
	}
	return out
}

func containsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// TestTutorRevealReachesTheTable is the retrofit's proof. Before S22
// a tutor's "reveal it" was a KnownBy write and nothing else: the
// opponent could read the card once it landed in a hand, and had no
// way to know when it got there or that a tutor was what put it
// there. The knowledge half is asserted by
// TestSearchRevealsOnlyWhatWasTaken; this is the announcement half,
// and it has to hold the same line about what was NOT taken.
func TestTutorRevealReachesTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: true,
			Reason: "Test Tutor — a basic land, revealed, to hand",
		})
	})
	answerSearchNamed(t, g, me.ID, "Island")

	rs := revealsFor(g, opp.ID)
	if len(rs) != 1 {
		t.Fatalf("the opponent sees %d reveals, want 1: %+v", len(rs), rs)
	}
	names := revealedNames(rs)
	if !containsName(names, "Island") {
		t.Errorf("the tutored card was not announced; got %v", names)
	}
	if containsName(names, "Forest") {
		t.Error("a candidate that was NOT taken was announced to the table")
	}
	if rs[0].Reason != "Test Tutor — a basic land, revealed, to hand" {
		t.Errorf("reveal reason is %q, want the search's own banner copy", rs[0].Reason)
	}
}

// TestTutorWithoutRevealAnnouncesNothing is the control. Demonic
// Tutor and Entomb do not print "reveal", and the retrofit must not
// have made every search loud.
func TestTutorWithoutRevealAnnouncesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: false,
		})
	})
	answerSearchNamed(t, g, me.ID, "Island")

	if rs := revealsFor(g, opp.ID); len(rs) != 0 {
		t.Errorf("a search that does not say REVEAL announced %d reveals: %+v", len(rs), rs)
	}
}

// TestHermitDruidAnnouncesTheWholeRun. The card reveals every card it
// turns over, not only the land it keeps — and the milled ones are
// the half the old implementation left the table to infer from a pile
// of graveyard motion.
func TestHermitDruidAnnouncesTheWholeRun(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	druid := b21Push(g, me.ID, "Hermit Druid", "Creature — Human Druid", b21HermitDruidOracle, 1, 1, "G")
	seedSearchLibrary(me,
		game.Card{Name: "Top Bolt", TypeLine: "Instant"},
		game.Card{Name: "Middle Nonbasic", TypeLine: "Land"},
		game.Card{Name: "The Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Untouched Below", TypeLine: "Sorcery"},
	)
	advanceToMain(t, g)
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})

	names := revealedNames(revealsFor(g, opp.ID))
	for _, want := range []string{"Top Bolt", "Middle Nonbasic", "The Forest"} {
		if !containsName(names, want) {
			t.Errorf("%q was turned over and not announced; got %v", want, names)
		}
	}
	if containsName(names, "Untouched Below") {
		t.Errorf("a card below the land was announced; got %v", names)
	}
}

// TestConsumingAberrationAnnouncesWhatItMills retires the declared
// simplification on that card ("Reveals is not modelled as a reveal").
func TestConsumingAberrationAnnouncesWhatItMills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Consuming Aberration", "Creature — Horror",
		"9b55fb72-237d-4935-b645-8ebc6eb4140e", false)
	seedSearchLibrary(opp,
		game.Card{Name: "Their Bolt", TypeLine: "Instant"},
		game.Card{Name: "Their Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Their Untouched", TypeLine: "Sorcery"},
	)
	advanceToMain(t, g)

	// Any cast fires the trigger.
	castVanillaSorceryForRevealTest(t, g, me)
	passPriorityAroundTable(t, g)

	names := revealedNames(revealsFor(g, me.ID))
	for _, want := range []string{"Their Bolt", "Their Swamp"} {
		if !containsName(names, want) {
			t.Errorf("%q was milled by the trigger and not announced; got %v", want, names)
		}
	}
	if containsName(names, "Their Untouched") {
		t.Errorf("a card below the land was announced; got %v", names)
	}
}

// castVanillaSorceryForRevealTest puts a plain sorcery into a
// player's hand and casts it, purely to fire a "whenever you cast a
// spell" trigger.
func castVanillaSorceryForRevealTest(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       "Trigger Bait",
			TypeLine:   "Sorcery",
			Owner:      p.ID,
			Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}
