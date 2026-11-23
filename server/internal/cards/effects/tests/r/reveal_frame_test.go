package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTutorRevealReachesTheTable is the retrofit's proof. Before S22
// a tutor's "reveal it" was a KnownBy write and nothing else: the
// opponent could read the card once it landed in a hand, and had no
// way to know when it got there or that a tutor was what put it
// there. The knowledge half is asserted by
// TestSearchRevealsOnlyWhatWasTaken; this is the announcement half,
// and it has to hold the same line about what was NOT taken.
func TestTutorRevealReachesTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   effects.IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: true,
			Reason: "Test Tutor — a basic land, revealed, to hand",
		})
	})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")

	rs := effectstest.RevealsFor(g, opp.ID)
	if len(rs) != 1 {
		t.Fatalf("the opponent sees %d reveals, want 1: %+v", len(rs), rs)
	}
	names := effectstest.RevealedNames(rs)
	if !effectstest.ContainsName(names, "Island") {
		t.Errorf("the tutored card was not announced; got %v", names)
	}
	if effectstest.ContainsName(names, "Forest") {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   effects.IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: false,
		})
	})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")

	if rs := effectstest.RevealsFor(g, opp.ID); len(rs) != 0 {
		t.Errorf("a search that does not say REVEAL announced %d reveals: %+v", len(rs), rs)
	}
}

// TestHermitDruidAnnouncesTheWholeRun. The card reveals every card it
// turns over, not only the land it keeps — and the milled ones are
// the half the old implementation left the table to infer from a pile
// of graveyard motion.
func TestHermitDruidAnnouncesTheWholeRun(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	druid := effectstest.B21Push(g, me.ID, "Hermit Druid", "Creature — Human Druid", effectstest.B21HermitDruidOracle, 1, 1, "G")
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Top Bolt", TypeLine: "Instant"},
		game.Card{Name: "Middle Nonbasic", TypeLine: "Land"},
		game.Card{Name: "The Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Untouched Below", TypeLine: "Sorcery"},
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})

	names := effectstest.RevealedNames(effectstest.RevealsFor(g, opp.ID))
	for _, want := range []string{"Top Bolt", "Middle Nonbasic", "The Forest"} {
		if !effectstest.ContainsName(names, want) {
			t.Errorf("%q was turned over and not announced; got %v", want, names)
		}
	}
	if effectstest.ContainsName(names, "Untouched Below") {
		t.Errorf("a card below the land was announced; got %v", names)
	}
}

// TestConsumingAberrationAnnouncesWhatItMills retires the declared
// simplification on that card ("Reveals is not modelled as a reveal").
func TestConsumingAberrationAnnouncesWhatItMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Consuming Aberration's power/toughness are a CDA equal to the
	// number of cards in its controller's opponents' graveyards (CR
	// 604.3). pushCatalogPermanent now stamps a real battlefield entry
	// (#1698), so the layer engine actually computes that CDA instead
	// of leaving the helper's literal 1/1 in place — and with every
	// graveyard empty at the top of the game, a freshly-entered
	// Aberration is a real 0/0 that dies to CR 704.5f before it ever
	// sees a spell cast. Seed one card into an opponent's graveyard
	// first so it has 1 toughness and survives to trigger.
	opp.Graveyard.PushTop(game.Card{Name: "Already Dead", TypeLine: "Instant"})
	effectstest.PushCatalogPermanent(g, me.ID, "Consuming Aberration", "Creature — Horror",
		"9b55fb72-237d-4935-b645-8ebc6eb4140e", false)
	effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "Their Bolt", TypeLine: "Instant"},
		game.Card{Name: "Their Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Their Untouched", TypeLine: "Sorcery"},
	)
	effectstest.AdvanceToMain(t, g)

	// Any cast fires the trigger.
	effectstest.CastVanillaSorceryForRevealTest(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)

	names := effectstest.RevealedNames(effectstest.RevealsFor(g, me.ID))
	for _, want := range []string{"Their Bolt", "Their Swamp"} {
		if !effectstest.ContainsName(names, want) {
			t.Errorf("%q was milled by the trigger and not announced; got %v", want, names)
		}
	}
	if effectstest.ContainsName(names, "Their Untouched") {
		t.Errorf("a card below the land was announced; got %v", names)
	}
}
